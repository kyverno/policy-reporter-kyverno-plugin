package kubernetes

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	apiV1 "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/kyverno/v1"
	kyvernoV1 "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/typed/kyverno/v1"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

type Queue struct {
	mapper        Mapper
	publisher     *kyverno.EventPublisher
	queue         workqueue.RateLimitingInterface
	client        kyvernoV1.KyvernoV1Interface
	dynamicClient dynamic.Interface
	lock          *sync.Mutex
	cache         sets.Set[string]
}

func (q *Queue) Add(obj *v1.PartialObjectMetadata) error {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}

	q.queue.Add(key)

	return nil
}

func (q *Queue) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	defer q.queue.ShutDown()

	for i := 0; i < workers; i++ {
		go wait.Until(q.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (q *Queue) runWorker() {
	for q.processNextItem() {
	}
}

func (q *Queue) processNextItem() bool {
	obj, quit := q.queue.Get()
	if quit {
		return false
	}
	key := obj.(string)
	defer q.queue.Done(key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		q.queue.Forget(key)
		return true
	}

	var polr apiV1.PolicyInterface
	var cont *unstructured.Unstructured

	if namespace == "" {
		polr, err = q.client.ClusterPolicies().Get(context.Background(), name, v1.GetOptions{})
		cont, err = q.dynamicClient.Resource(apiV1.SchemeGroupVersion.WithResource("clusterpolicies")).Get(context.Background(), name, v1.GetOptions{})
	} else {
		polr, err = q.client.Policies(namespace).Get(context.Background(), name, v1.GetOptions{})
		cont, err = q.dynamicClient.Resource(apiV1.SchemeGroupVersion.WithResource("policies")).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	}

	if errors.IsNotFound(err) {
		if namespace == "" {
			polr = &apiV1.ClusterPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name: name,
				},
			}
		} else {
			polr = &apiV1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
		}

		func() {
			q.lock.Lock()
			defer q.lock.Unlock()
			q.cache.Delete(key)
		}()
		q.publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Deleted, Policy: q.mapper.MapPolicy(polr, nil)})

		return true
	}

	event := func() kyverno.Event {
		q.lock.Lock()
		defer q.lock.Unlock()
		event := kyverno.Added
		if q.cache.Has(key) {
			event = kyverno.Updated
		} else {
			q.cache.Insert(key)
		}
		return event
	}()

	q.handleErr(err, key)

	polr.SetAPIVersion(cont.Object["apiVersion"].(string))
	polr.SetKind(cont.Object["kind"].(string))

	q.publisher.Publish(kyverno.LifecycleEvent{Type: event, Policy: q.mapper.MapPolicy(polr, cont)})

	return true
}

func (q *Queue) handleErr(err error, key interface{}) {
	if err == nil {
		q.queue.Forget(key)
		return
	}

	if q.queue.NumRequeues(key) < 5 {
		zap.L().Error("process report", zap.Any("report", key), zap.Error(err))

		q.queue.AddRateLimited(key)
		return
	}

	q.queue.Forget(key)

	runtime.HandleError(err)
	zap.L().Warn("dropping report out of the queue", zap.Any("report", key), zap.Error(err))
}

func NewQueue(publisher *kyverno.EventPublisher, queue workqueue.RateLimitingInterface, client kyvernoV1.KyvernoV1Interface, dClient dynamic.Interface) *Queue {
	return &Queue{
		mapper:        NewMapper(),
		publisher:     publisher,
		queue:         queue,
		client:        client,
		dynamicClient: dClient,
		cache:         sets.New[string](),
		lock:          &sync.Mutex{},
	}
}
