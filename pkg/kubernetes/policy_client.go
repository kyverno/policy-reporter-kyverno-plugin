package kubernetes

import (
	"context"
	"log"
	"sync"
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
		Resource: "clusterpolicies",
	}

	clusterPolicySchema = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "policies",
	}
)

type policyClient struct {
	client                dynamic.Interface
	found                 map[string]bool
	mapper                Mapper
	mx                    *sync.Mutex
	restartWatchOnFailure time.Duration
}

// GetFoundResources as Map of Names
func (c *policyClient) GetFoundResources() map[string]bool {
	return c.found
}

// StartWatching returns a stream of incomming LifecyclePolicyEvents from the Kubernetes API
func (c *policyClient) StartWatching(ctx context.Context) <-chan kyverno.LifecycleEvent {
	eventChan := make(chan kyverno.LifecycleEvent)

	for _, version := range []schema.GroupVersionResource{policySchema, clusterPolicySchema} {
		go func(v schema.GroupVersionResource) {
			for {
				factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(c.client, time.Hour, corev1.NamespaceAll, nil)
				c.watchPolicyCRD(ctx, v, factory, eventChan)
				time.Sleep(2 * time.Second)
			}
		}(version)
	}

	for {
		if len(c.found) == 2 {
			break
		}
	}

	return eventChan
}

func (c *policyClient) watchPolicyCRD(
	ctx context.Context,
	resource schema.GroupVersionResource,
	factory dynamicinformer.DynamicSharedInformerFactory,
	eventChan chan<- kyverno.LifecycleEvent,
) {
	informer := factory.ForResource(resource).Informer()
	ctx, cancel := context.WithCancel(ctx)

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
		c.mx.Lock()
		delete(c.found, resource.String())
		c.mx.Unlock()
		cancel()

		log.Printf("[WARNING] Resource registration failed: %s\n", resource.String())
	})

	go c.handleCRDRegistration(ctx, informer, resource)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				eventChan <- kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: &kyverno.Policy{}}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				eventChan <- kyverno.LifecycleEvent{Type: kyverno.Deleted, NewPolicy: c.mapper.MapPolicy(item.Object), OldPolicy: &kyverno.Policy{}}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*unstructured.Unstructured); ok {
				newPolicy := c.mapper.MapPolicy(item.Object)

				var oldPolicy *kyverno.Policy
				if oldItem, ok := oldObj.(*unstructured.Unstructured); ok {
					oldPolicy = c.mapper.MapPolicy(oldItem.Object)
				}

				eventChan <- kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: newPolicy, OldPolicy: oldPolicy}
			}
		},
	})

	informer.Run(ctx.Done())
}

func (c *policyClient) handleCRDRegistration(ctx context.Context, informer cache.SharedIndexInformer, r schema.GroupVersionResource) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if informer.HasSynced() {
				c.mx.Lock()
				c.found[r.String()] = true
				c.mx.Unlock()

				log.Printf("[INFO] Resource registered: %s\n", r.String())
				return
			}
		}
	}
}

// NewPolicyClient creates a new PolicyClient based on the kubernetes go-client
func NewPolicyClient(client dynamic.Interface, mapper Mapper, restartWatchOnFailure time.Duration) kyverno.PolicyClient {
	return &policyClient{
		client:                client,
		mapper:                mapper,
		found:                 make(map[string]bool),
		mx:                    new(sync.Mutex),
		restartWatchOnFailure: restartWatchOnFailure,
	}
}
