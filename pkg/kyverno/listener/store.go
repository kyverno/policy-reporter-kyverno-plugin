package listener

import (
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// NewStoreListener for Policy kyverno.LifecycleEvent
func NewStoreListener(store *kyverno.PolicyStore) kyverno.PolicyListener {
	return func(event kyverno.LifecycleEvent) {
		if event.Type == kyverno.Deleted {
			store.Remove(event.Policy.GetID())
			return
		}

		store.Add(event.Policy)
	}
}
