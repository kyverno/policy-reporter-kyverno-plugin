package kubernetes_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/watch"
)

type fakeClient struct {
	policies             []unstructured.Unstructured
	clusterPolicies      []unstructured.Unstructured
	policyWatcher        *watch.FakeWatcher
	clusterPolicyWatcher *watch.FakeWatcher
	policyError          error
	clusterPolicyError   error
}

func (f *fakeClient) ListClusterPolicies() (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{
		Items: f.clusterPolicies,
	}, f.clusterPolicyError
}

func (f *fakeClient) ListPolicies() (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{
		Items: f.policies,
	}, f.policyError
}

func (f *fakeClient) WatchClusterPolicies() (watch.Interface, error) {
	return f.clusterPolicyWatcher, f.clusterPolicyError
}

func (f *fakeClient) WatchPolicies() (watch.Interface, error) {
	return f.policyWatcher, f.policyError
}

func NewPolicyAdapter() *fakeClient {
	return &fakeClient{
		policies:             make([]unstructured.Unstructured, 0),
		clusterPolicies:      make([]unstructured.Unstructured, 0),
		policyWatcher:        watch.NewFake(),
		clusterPolicyWatcher: watch.NewFake(),
	}
}

func NewPolicy() unstructured.Unstructured {
	obj := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, &obj)

	return obj
}

func Test_FetchPolicies(t *testing.T) {
	fakeAdapter := NewPolicyAdapter()

	client := kubernetes.NewPolicyClient(
		fakeAdapter,
		kyverno.NewPolicyStore(),
		kubernetes.NewMapper(),
	)

	fakeAdapter.policies = append(fakeAdapter.policies, NewPolicy())

	policies, err := client.FetchPolicies()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(policies) != 1 {
		t.Fatal("Expected one Policy")
	}

	expected := kubernetes.NewMapper().MapPolicy(NewPolicy().Object)
	policy := policies[0]

	if policy.Name != expected.Name {
		t.Errorf("Expected Policy Name %s", expected.Name)
	}
}

func Test_FetchPoliciesError(t *testing.T) {
	fakeAdapter := NewPolicyAdapter()
	fakeAdapter.policyError = errors.New("")

	client := kubernetes.NewPolicyClient(
		fakeAdapter,
		kyverno.NewPolicyStore(),
		kubernetes.NewMapper(),
	)

	_, err := client.FetchPolicies()
	if err == nil {
		t.Error("Configured Error should be returned")
	}
}

func Test_PolicyWatcher(t *testing.T) {
	fakeAdapter := NewPolicyAdapter()
	store := kyverno.NewPolicyStore()

	client := kubernetes.NewPolicyClient(
		fakeAdapter,
		store,
		kubernetes.NewMapper(),
	)

	wg := sync.WaitGroup{}
	wg.Add(1)

	results := make([]kyverno.Policy, 0, 1)

	client.RegisterCallback(func(_ watch.EventType, p kyverno.Policy, o kyverno.Policy) {
		results = append(results, p)
		wg.Done()
	})

	go client.StartWatching()

	pol := NewPolicy()

	fakeAdapter.policyWatcher.Add(&pol)

	wg.Wait()

	if len(results) != 1 {
		t.Error("Should receive 1 Policy")
	}

	if len(store.List()) != 1 {
		t.Error("Should include a single Policy")
	}
}

func Test_PolicyModifyWatcher(t *testing.T) {
	fakeAdapter := NewPolicyAdapter()
	store := kyverno.NewPolicyStore()

	client := kubernetes.NewPolicyClient(
		fakeAdapter,
		store,
		kubernetes.NewMapper(),
	)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]kyverno.Policy, 0, 2)

	client.RegisterCallback(func(_ watch.EventType, p kyverno.Policy, o kyverno.Policy) {
		results = append(results, p)
		wg.Done()
	})

	go client.StartWatching()

	pol := NewPolicy()

	fakeAdapter.policyWatcher.Add(&pol)
	fakeAdapter.policyWatcher.Modify(&pol)

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive 2 Policy")
	}

	if len(store.List()) != 1 {
		t.Error("Should include a single Policy")
	}
}

func Test_PolicyDeleteWatcher(t *testing.T) {
	fakeAdapter := NewPolicyAdapter()
	store := kyverno.NewPolicyStore()

	client := kubernetes.NewPolicyClient(
		fakeAdapter,
		store,
		kubernetes.NewMapper(),
	)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]kyverno.Policy, 0, 2)

	client.RegisterCallback(func(_ watch.EventType, p kyverno.Policy, o kyverno.Policy) {
		results = append(results, p)
		wg.Done()
	})

	go client.StartWatching()

	pol := NewPolicy()

	fakeAdapter.policyWatcher.Add(&pol)
	fakeAdapter.policyWatcher.Delete(&pol)

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive 2 Policy")
	}

	if len(store.List()) != 0 {
		t.Error("Should be empty after deletion")
	}
}
