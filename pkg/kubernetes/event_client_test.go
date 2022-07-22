package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	eventsv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	createdAt = v1.NewTime(time.Now().Add(1 * time.Minute))
	updatedAt = v1.NewTime(time.Now().Add(2 * time.Minute))

	baseEvent = &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:              "nginx.12345",
			Namespace:         "default",
			UID:               "27d99ed6-2535-4a6b-a33a-dea49062fdcd",
			CreationTimestamp: createdAt,
		},
		InvolvedObject: corev1.ObjectReference{
			Name: "require-request-and-limits",
			UID:  "1f2b937f-8ae9-4071-aa37-4d35c87965a3",
		},
		Source: corev1.EventSource{
			Component: "kyverno-admission",
		},
		Message:       "Pod test/nginx: [require-resource-request] fail (blocked)",
		Reason:        "PolicyViolation",
		Type:          "Warning",
		LastTimestamp: updatedAt,
	}

	basePolicy = kyverno.Policy{
		Name:              "require-request-and-limits",
		Kind:              "ClusterPolicy",
		Category:          "Best Practices",
		Severity:          "medium",
		UID:               "1f2b937f-8ae9-4071-aa37-4d35c87965a3",
		CreationTimestamp: time.Now(),
		Rules: []*kyverno.Rule{
			{
				Name: "require-resource-request",
			},
		},
	}
)

func NewEventFakeCilent() (k8s.Interface, eventsv1.EventInterface) {
	client := fake.NewSimpleClientset()

	return client, client.CoreV1().Events("default")
}

type eventStore struct {
	store []kyverno.PolicyViolation
	rwm   *sync.RWMutex
}

func (s *eventStore) Add(r kyverno.PolicyViolation) {
	s.rwm.Lock()
	s.store = append(s.store, r)
	s.rwm.Unlock()
}

func (s *eventStore) Get(index int) kyverno.PolicyViolation {
	return s.store[index]
}

func (s *eventStore) List() []kyverno.PolicyViolation {
	return s.store
}

func newEventStore(size int) *eventStore {
	return &eventStore{
		store: make([]kyverno.PolicyViolation, 0, size),
		rwm:   &sync.RWMutex{},
	}
}

func Test_EventWatcher(t *testing.T) {
	ctx := context.Background()
	kclient, pclient := NewEventFakeCilent()
	policyStore := kyverno.NewPolicyStore()
	policyStore.Add(basePolicy)

	client := kubernetes.NewEventClient(kclient, policyStore, time.Millisecond, "default")

	store := newEventStore(3)
	eventChan := client.StartWatching(ctx)

	t.Run("AddListener", func(t *testing.T) {
		_, _ = pclient.Create(ctx, baseEvent, v1.CreateOptions{})

		violation := <-eventChan

		store.Add(violation)

		if len(store.List()) != 1 {
			t.Error("Should receive Add Event")
		}

		checkViolationPolicy(violation, t)
		checkViolationResource(violation, t)
		checkViolationEvent(violation, baseEvent, t)
	})

	t.Run("UpdateListener", func(t *testing.T) {
		event := baseEvent.DeepCopy()
		event.LastTimestamp = v1.Now()

		_, _ = pclient.Update(ctx, event, v1.UpdateOptions{})

		violation := <-eventChan

		store.Add(violation)

		if len(store.List()) != 2 {
			t.Error("Should receive Update Event")
		}

		checkViolationPolicy(violation, t)
		checkViolationResource(violation, t)
		checkViolationEvent(violation, event, t)
	})

	t.Run("ClusterResource Event", func(t *testing.T) {
		event := baseEvent.DeepCopy()
		event.Message = "Namespace test: [require-resource-request] fail (blocked)"
		event.UID = "58ee457c-465b-482a-a965-b206fe8567bd"

		_, _ = pclient.Update(ctx, event, v1.UpdateOptions{})

		violation := <-eventChan

		store.Add(violation)

		if len(store.List()) != 3 {
			t.Error("Should receive Add Event")
		}

		checkViolationPolicy(violation, t)
		checkViolationEvent(violation, event, t)

		if violation.Resource.Kind != "Namespace" {
			t.Errorf("expected Resource.Kind to be '%s', got %s", "Namespace", violation.Resource.Kind)
		}
		if violation.Resource.Name != "test" {
			t.Errorf("expected Resource.Name to be '%s', got %s", "test", violation.Resource.Name)
		}
		if violation.Resource.Namespace != "" {
			t.Errorf("expected Resource.Namespace to be '%s', got %s", "", violation.Resource.Namespace)
		}
	})

	t.Run("Ignore none blocked events", func(t *testing.T) {
		event := baseEvent.DeepCopy()
		event.Message = "Namespace test: [require-resource-request] fail"
		event.LastTimestamp = v1.Now()

		_, _ = pclient.Create(ctx, event, v1.CreateOptions{})
	})
}

func Test_NotBlockedEvent(t *testing.T) {
	ctx := context.Background()
	kclient, pclient := NewEventFakeCilent()
	policyStore := kyverno.NewPolicyStore()
	policyStore.Add(basePolicy)

	client := kubernetes.NewEventClient(kclient, policyStore, time.Millisecond, "default")

	eventChan := client.StartWatching(ctx)

	event := baseEvent.DeepCopy()
	event.Message = "Namespace test: [require-resource-request] fail"
	event.LastTimestamp = v1.Now()

	_, _ = pclient.Create(ctx, event, v1.CreateOptions{})
	time.Sleep(1 * time.Millisecond)
	_, _ = pclient.Update(ctx, event, v1.UpdateOptions{})

	go func() {
		<-eventChan
		t.Error("Should not receive new event")
	}()

	time.Sleep(1 * time.Second)
}

func Test_UnknownPolicy(t *testing.T) {
	ctx := context.Background()
	kclient, pclient := NewEventFakeCilent()
	policyStore := kyverno.NewPolicyStore()
	policyStore.Add(basePolicy)

	client := kubernetes.NewEventClient(kclient, policyStore, time.Millisecond, "default")

	eventChan := client.StartWatching(ctx)

	event := baseEvent.DeepCopy()
	event.InvolvedObject.Name = "unknown"
	event.InvolvedObject.UID = "4baaf7cc-4f7c-4746-b8e3-1dd7cc002c75"

	_, _ = pclient.Create(ctx, event, v1.CreateOptions{})
	time.Sleep(1 * time.Millisecond)
	_, _ = pclient.Update(ctx, event, v1.UpdateOptions{})

	go func() {
		<-eventChan
		t.Error("Should not receive new event")
	}()

	time.Sleep(1 * time.Second)
}

func checkViolationPolicy(violation kyverno.PolicyViolation, t *testing.T) {
	if violation.Policy.Category != basePolicy.Category {
		t.Errorf("expected Category to be '%s', got %s", basePolicy.Category, violation.Policy.Category)
	}
	if violation.Policy.Severity != basePolicy.Severity {
		t.Errorf("expected Severity to be '%s', got %s", basePolicy.Severity, violation.Policy.Severity)
	}
	if violation.Policy.Name != basePolicy.Name {
		t.Errorf("expected Policy to be '%s', got %s", basePolicy.Name, violation.Policy.Name)
	}
	if violation.Policy.Rule != basePolicy.Rules[0].Name {
		t.Errorf("expected Rule to be '%s', got %s", basePolicy.Rules[0].Name, violation.Policy.Rule)
	}
	if violation.Policy.Message != basePolicy.Rules[0].ValidateMessage {
		t.Errorf("expected Message to be '%s', got %s", basePolicy.Rules[0].ValidateMessage, violation.Policy.Message)
	}
}

func checkViolationResource(violation kyverno.PolicyViolation, t *testing.T) {
	if violation.Resource.Kind != "Pod" {
		t.Errorf("expected Resource.Kind to be '%s', got %s", "Pod", violation.Resource.Kind)
	}
	if violation.Resource.Name != "nginx" {
		t.Errorf("expected Resource.Name to be '%s', got %s", "nginx", violation.Resource.Name)
	}
	if violation.Resource.Namespace != "test" {
		t.Errorf("expected Resource.Namespace to be '%s', got %s", "test", violation.Resource.Namespace)
	}
}

func checkViolationEvent(violation kyverno.PolicyViolation, event *corev1.Event, t *testing.T) {
	if violation.Event.Name != event.Name {
		t.Errorf("expected Event.Name to be '%s', got %s", event.Name, violation.Event.Name)
	}
	if violation.Event.UID != string(event.UID) {
		t.Errorf("expected Event.UID to be '%s', got %s", event.UID, violation.Event.UID)
	}
}
