package listener_test

import (
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

func NewPolicy() kyverno.Policy {
	return kyverno.Policy{
		Name:     "disallow-host-path",
		UID:      "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7",
		Severity: "medium",
		Category: "Pod Security Standards (Default)",
		Background: func(val bool) *bool {
			return &val
		}(false),
		Rules: []*kyverno.Rule{
			{
				Name:            "host-path",
				ValidateMessage: "HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set.",
			},
		},
	}
}
