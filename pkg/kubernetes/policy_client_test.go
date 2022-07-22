package kubernetes_test

import (
	"context"
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

var (
	clusterPolicySchema = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "clusterpolicies",
	}

	policySchema = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "policies",
	}

	gvrToListKind = map[schema.GroupVersionResource]string{
		policySchema:        "PolicyList",
		clusterPolicySchema: "ClusterPolicyList",
	}
)

type store struct {
	store []kyverno.LifecycleEvent
	rwm   *sync.RWMutex
}

func (s *store) Add(r kyverno.LifecycleEvent) {
	s.rwm.Lock()
	s.store = append(s.store, r)
	s.rwm.Unlock()
}

func (s *store) Get(index int) kyverno.LifecycleEvent {
	return s.store[index]
}

func (s *store) List() []kyverno.LifecycleEvent {
	return s.store
}

func newStore(size int) *store {
	return &store{
		store: make([]kyverno.LifecycleEvent, 0, size),
		rwm:   &sync.RWMutex{},
	}
}

func NewFakeCilent() (dynamic.Interface, dynamic.ResourceInterface, dynamic.ResourceInterface) {
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind)

	return client, client.Resource(clusterPolicySchema), client.Resource(policySchema).Namespace("test")
}

func Test_PolicyWatcher(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	kclient, _, pclient := NewFakeCilent()

	store := newStore(3)
	eventChan := make(chan kyverno.LifecycleEvent)

	publisher := kyverno.NewEventPublisher()
	publisher.RegisterListener(func(le kyverno.LifecycleEvent) {
		eventChan <- le
	})

	client := kubernetes.NewPolicyClient(kclient, kubernetes.NewMapper(), publisher)
	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("HasSynced", func(t *testing.T) {
		if !client.HasSynced() {
			t.Errorf("Expected sync success")
		}
	})

	t.Run("AddListener", func(t *testing.T) {
		pclient.Create(ctx, convert(minPolicy), metav1.CreateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 1 {
			t.Error("Should receive Add Event")
		}
	})

	t.Run("UpdateListener", func(t *testing.T) {
		pclient.Update(ctx, convert(nsPolicy), metav1.UpdateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 2 {
			t.Error("Should receive Update Event")
		}
	})

	t.Run("DeleteListener", func(t *testing.T) {
		pclient.Delete(ctx, "disallow-host-path", metav1.DeleteOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 3 {
			t.Error("Should receive Delete Event")
		}
	})
}

func Test_ClusterPolicyWatcher(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	kclient, pclient, _ := NewFakeCilent()

	store := newStore(3)
	eventChan := make(chan kyverno.LifecycleEvent)

	publisher := kyverno.NewEventPublisher()
	publisher.RegisterListener(func(le kyverno.LifecycleEvent) {
		eventChan <- le
	})

	client := kubernetes.NewPolicyClient(kclient, kubernetes.NewMapper(), publisher)
	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("HasSynced", func(t *testing.T) {
		if !client.HasSynced() {
			t.Errorf("Expected sync success")
		}
	})

	t.Run("AddListener", func(t *testing.T) {
		pclient.Create(ctx, convert(clusterPolicy), metav1.CreateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 1 {
			t.Error("Should receive Add Event")
		}
	})

	t.Run("UpdateListener", func(t *testing.T) {
		pclient.Update(ctx, convert(minClusterPolicy), metav1.UpdateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 2 {
			t.Error("Should receive Update Event")
		}
	})

	t.Run("DeleteListener", func(t *testing.T) {
		pclient.Delete(ctx, "disallow-host-path", metav1.DeleteOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 3 {
			t.Error("Should receive Delete Event")
		}
	})
}
