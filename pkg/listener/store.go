package listener

import (
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func NewStoreListener(store *kyverno.PolicyStore) kyverno.PolicyListener {
	return func(event kyverno.LifecycleEvent) {
		if event.Type == kyverno.Deleted {
			store.Remove(event.NewPolicy.UID)
			return
		}

		store.Add(event.NewPolicy)
	}
}
