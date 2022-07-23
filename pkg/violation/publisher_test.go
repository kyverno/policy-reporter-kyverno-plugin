package violation_test

import (
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
)

func Test_PublishPolicyViolation(t *testing.T) {
	vChan := make(chan violation.PolicyViolation)

	publisher := violation.NewPublisher()
	publisher.RegisterListener(func(pv violation.PolicyViolation) {
		vChan <- pv
	})

	go func() {
		publisher.Publish(violation.PolicyViolation{Event: violation.Event{Name: "test"}})
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
