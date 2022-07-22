package kyverno_test

import (
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func NewPolicy() kyverno.Policy {
	return kyverno.Policy{
		Name:     "disallow-host-path",
		UID:      "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7",
		Severity: "medium",
		Category: "Pod Security Standards (Default)",
		Rules: []*kyverno.Rule{
			{
				Name:            "host-path",
				ValidateMessage: "HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set.",
			},
		},
	}
}

func Test_PolicyStore(t *testing.T) {
	store := kyverno.NewPolicyStore()
	pol := NewPolicy()

	t.Run("Add/Get", func(t *testing.T) {
		_, ok := store.Get(pol.UID)
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(pol)
		_, ok = store.Get(pol.UID)
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("List", func(t *testing.T) {
		items := store.List()
		if len(items) != 1 {
			t.Errorf("Should return List with the added Report")
		}
	})

	t.Run("Delete/Get", func(t *testing.T) {
		_, ok := store.Get(pol.UID)
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(pol.UID)
		_, ok = store.Get(pol.UID)
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})
}
