package kubernetes

import (
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	apiV1 "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/kyverno/v1"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

type policyClient struct {
	queue   *Queue
	factory metadatainformer.SharedInformerFactory
	pol     informers.GenericInformer
	cpol    informers.GenericInformer
	synced  bool
}

func (c *policyClient) HasSynced() bool {
	return c.synced
}

func (c *policyClient) Sync(stopper chan struct{}) error {
	policyInformer := c.configureInformer(c.pol.Informer())
	clusterPolicyInformaer := c.configureInformer(c.cpol.Informer())

	c.factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, policyInformer.HasSynced) {
		return fmt.Errorf("failed to sync policies")
	}

	if !cache.WaitForCacheSync(stopper, clusterPolicyInformaer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policies")
	}

	c.synced = true

	return nil
}

func (c *policyClient) Run(worker int, stopper chan struct{}) error {
	if err := c.Sync(stopper); err != nil {
		return err
	}

	c.queue.Run(worker, stopper)

	return nil
}

func (c *policyClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				c.queue.Add(item)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				c.queue.Add(item)
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if item, ok := newObj.(*v1.PartialObjectMetadata); ok {
				c.queue.Add(item)
			}
		},
	})

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		c.synced = false
	})

	return informer
}

// NewClient creates a new PolicyClient based on the kubernetes go-client
func NewClient(client metadata.Interface, queue *Queue) kyverno.PolicyClient {
	factory := metadatainformer.NewSharedInformerFactory(client, 15*time.Minute)
	pol := factory.ForResource(apiV1.SchemeGroupVersion.WithResource("policies"))
	cpol := factory.ForResource(apiV1.SchemeGroupVersion.WithResource("clusterpolicies"))

	return &policyClient{
		factory: factory,
		pol:     pol,
		cpol:    cpol,
		queue:   queue,
	}
}
