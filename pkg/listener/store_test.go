package listener_test

import (
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/listener"
)

func Test_StoreListener(t *testing.T) {
	store := kyverno.NewPolicyStore()
	pol := NewPolicy()

	t.Run("Save Policy", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: pol, OldPolicy: kyverno.Policy{}})

		if _, ok := store.Get(pol.UID); !ok {
			t.Error("Expected Report to be stored")
		}
	})
	t.Run("Remove Deleted Policy", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(kyverno.LifecycleEvent{Type: kyverno.Deleted, NewPolicy: pol, OldPolicy: kyverno.Policy{}})

		if _, ok := store.Get(pol.UID); ok {
			t.Error("Expected Report to be removed")
		}
	})
}
