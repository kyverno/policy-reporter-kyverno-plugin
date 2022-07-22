package kyverno_test

import (
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_PublishLifecycleEvents(t *testing.T) {
	var event kyverno.LifecycleEvent

	wg := sync.WaitGroup{}
	wg.Add(1)

	publisher := kyverno.NewEventPublisher()
	publisher.RegisterListener(func(le kyverno.LifecycleEvent) {
		event = le
		wg.Done()
	})

	publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: kyverno.Policy{}, OldPolicy: kyverno.Policy{}})

	wg.Wait()

	if event.Type != kyverno.Updated {
		t.Error("Expected Event to be published to the listener")
	}
}

func Test_GetReisteredListeners(t *testing.T) {
	publisher := kyverno.NewEventPublisher()
	publisher.RegisterListener(func(le kyverno.LifecycleEvent) {})

	if len(publisher.GetListener()) != 1 {
		t.Error("Expected to get one registered listener back")
	}
}
