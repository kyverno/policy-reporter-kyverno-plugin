package kubernetes

import (
	"errors"
	"log"
	"sync"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kyverno"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type policyReportClient struct {
	policyAPI PolicyAdapter
	store     *kyverno.PolicyStore
	callbacks []kyverno.PolicyCallback
	mapper    Mapper
	started   bool
}

func (c *policyReportClient) RegisterCallback(cb kyverno.PolicyCallback) {
	c.callbacks = append(c.callbacks, cb)
}

func (c *policyReportClient) FetchPolicies() ([]kyverno.Policy, error) {
	var items []kyverno.Policy

	policies, err := c.policyAPI.ListPolicies()
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return items, err
	}

	for _, item := range policies.Items {
		items = append(items, c.mapper.MapPolicy(item.Object))
	}

	clusterPolicies, err := c.policyAPI.ListClusterPolicies()
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return items, err
	}

	for _, item := range clusterPolicies.Items {
		items = append(items, c.mapper.MapPolicy(item.Object))
	}

	return items, nil
}

func (c *policyReportClient) StartWatching() error {
	if c.started {
		return errors.New("PolicyClient.StartWatching was already started")
	}

	c.started = true
	errorChan := make(chan error)
	resultChan := make(chan watch.Event)

	go func() {
		for {
			result, err := c.policyAPI.WatchPolicies()
			if err != nil {
				c.started = false
				close(resultChan)
				errorChan <- err
				return
			}

			for result := range result.ResultChan() {
				resultChan <- result
			}
		}
	}()

	go func() {
		for {
			result, err := c.policyAPI.WatchClusterPolicies()
			if err != nil {
				c.started = false
				close(resultChan)
				errorChan <- err
				return
			}

			for result := range result.ResultChan() {
				resultChan <- result
			}
		}
	}()

	go func() {
		for result := range resultChan {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				report := c.mapper.MapPolicy(item.Object)
				c.executeHandler(result.Type, report)
			}
		}
	}()

	return <-errorChan
}

func (c *policyReportClient) executeHandler(e watch.EventType, pr kyverno.Policy) {
	opr, ok := c.store.Get(pr.GetIdentifier())
	if !ok {
		opr = kyverno.Policy{}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(c.callbacks))

	for _, cb := range c.callbacks {
		go func(
			callback kyverno.PolicyCallback,
			event watch.EventType,
			creport kyverno.Policy,
			oreport kyverno.Policy,
		) {
			callback(event, creport, oreport)
			wg.Done()
		}(cb, e, pr, opr)
	}

	wg.Wait()

	if e == watch.Deleted {
		c.store.Remove(pr.GetIdentifier())
		return
	}

	c.store.Add(pr)
}

// NewPolicyClient creates a new PolicyReportClient based on the kubernetes go-client
func NewPolicyClient(
	client PolicyAdapter,
	store *kyverno.PolicyStore,
	mapper Mapper,
) kyverno.PolicyClient {
	return &policyReportClient{
		policyAPI: client,
		store:     store,
		mapper:    mapper,
	}
}
