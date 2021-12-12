package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
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

var gvrToListKind = map[schema.GroupVersionResource]string{
	policySchema:        "PolicyList",
	clusterPolicySchema: "ClusterPolicyList",
}

func NewFakeCilent() (dynamic.Interface, dynamic.ResourceInterface) {
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind)

	return client, client.Resource(clusterPolicySchema)
}

func NewPolicyUnstructured(policy string) unstructured.Unstructured {
	obj := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, &obj)

	return obj
}

func Test_PolicyWatcher(t *testing.T) {
	ctx := context.Background()
	kclient, pclient := NewFakeCilent()

	client := kubernetes.NewPolicyClient(kclient, kubernetes.NewMapper(), time.Millisecond)

	store := newStore(3)
	eventChan := client.StartWatching(ctx)

	pol := NewPolicyUnstructured(policy)
	minPol := NewPolicyUnstructured(minPolicy)

	t.Run("GetFoundResources", func(t *testing.T) {
		found := client.GetFoundResources()
		if len(found) != 2 {
			t.Errorf("Expected 2 found resources, got %d", len(found))
		}
	})

	t.Run("AddListener", func(t *testing.T) {
		pclient.Create(ctx, &pol, v1.CreateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 1 {
			t.Error("Should receive Add Event")
		}
	})

	t.Run("UpdateListener", func(t *testing.T) {
		pclient.Update(ctx, &minPol, v1.UpdateOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 2 {
			t.Error("Should receive Update Event")
		}
	})

	t.Run("DeleteListener", func(t *testing.T) {
		pclient.Delete(ctx, "disallow-host-path", v1.DeleteOptions{})

		store.Add(<-eventChan)

		if len(store.List()) != 3 {
			t.Error("Should receive Delete Event")
		}
	})
}
