package kubernetes

import (
	"fmt"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var (
	policySchema = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "policies",
	}

	clusterPolicySchema = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "clusterpolicies",
	}
)

type policyClient struct {
	publisher kyverno.EventPublisher
	fatcory   dynamicinformer.DynamicSharedInformerFactory
	mapper    Mapper
	synced    bool
}

func (c *policyClient) HasSynced() bool {
	return c.synced
}

func (c *policyClient) Run(stopper chan struct{}) error {
	policyInformer := c.configurePolicy()
	clusterPolicyInformaer := c.configureClusterPolicy()

	c.fatcory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, policyInformer.HasSynced) {
		return fmt.Errorf("failed to sync policies")
	}

	if !cache.WaitForCacheSync(stopper, clusterPolicyInformaer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policies")
	}

	c.synced = true

	return nil
}

func (c *policyClient) configurePolicy() cache.SharedIndexInformer {
	polrInformer := c.fatcory.ForResource(policySchema).Informer()
	polrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: kyverno.Policy{}})
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Deleted, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: kyverno.Policy{}})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*unstructured.Unstructured); ok {
				newPolicy := c.mapper.MapPolicy(item.Object)

				var oldPolicy kyverno.Policy
				if oldItem, ok := oldObj.(*unstructured.Unstructured); ok {
					oldPolicy = c.mapper.MapPolicy(oldItem.Object)
				}

				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: newPolicy, OldPolicy: oldPolicy})
			}
		},
	})

	polrInformer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		c.synced = false
	})

	return polrInformer
}

func (c *policyClient) configureClusterPolicy() cache.SharedIndexInformer {
	polrInformer := c.fatcory.ForResource(clusterPolicySchema).Informer()
	polrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: kyverno.Policy{}})
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Deleted, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: kyverno.Policy{}})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*unstructured.Unstructured); ok {
				newPolicy := c.mapper.MapPolicy(item.Object)

				var oldPolicy kyverno.Policy
				if oldItem, ok := oldObj.(*unstructured.Unstructured); ok {
					oldPolicy = c.mapper.MapPolicy(oldItem.Object)
				}

				c.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: newPolicy, OldPolicy: oldPolicy})
			}
		},
	})

	polrInformer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		c.synced = false
	})

	return polrInformer
}

// NewPolicyClient creates a new PolicyClient based on the kubernetes go-client
func NewPolicyClient(client dynamic.Interface, mapper Mapper, publisher kyverno.EventPublisher) kyverno.PolicyClient {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Hour, corev1.NamespaceAll, nil)

	return &policyClient{
		fatcory:   factory,
		mapper:    mapper,
		publisher: publisher,
	}
}
