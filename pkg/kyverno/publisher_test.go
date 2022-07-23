package kyverno_test

import (
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_PublishLifecycleEvents(t *testing.T) {
	eChan := make(chan kyverno.LifecycleEvent)

	publisher := kyverno.NewEventPublisher()
	publisher.RegisterListener(func(le kyverno.LifecycleEvent) {
		eChan <- le
	})

	go func() {
		publisher.Publish(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: kyverno.Policy{}, OldPolicy: kyverno.Policy{}})
	}()

	event := <-eChan

	if event.Type != kyverno.Updated {
		t.Error("Expected Event to be published to the listener")
	}

	t.Run("GetListener", func(t *testing.T) {
		if len(publisher.GetListener()) != 1 {
			t.Error("Expected to get one registered listener back")
		}
	})
}
