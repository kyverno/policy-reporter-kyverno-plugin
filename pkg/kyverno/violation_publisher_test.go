package kyverno_test

import (
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func Test_PublishPolicyViolation(t *testing.T) {
	vChan := make(chan kyverno.PolicyViolation)

	publisher := kyverno.NewViolationPublisher()
	publisher.RegisterListener(func(pv kyverno.PolicyViolation) {
		vChan <- pv
	})

	go func() {
		publisher.Publish(kyverno.PolicyViolation{Event: kyverno.ViolationEvent{Name: "test"}})
	}()

	violation := <-vChan

	if violation.Event.Name != "test" {
		t.Error("Expected Event to be published to the listener")
	}

	t.Run("GetListener", func(t *testing.T) {
		if len(publisher.GetListener()) != 1 {
			t.Error("Expected to get one registered listener back")
		}
	})
}
