package listener

import (
	"strconv"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewPolicyMetricsListener for Policy kyverno.LifecycleEvent
func NewPolicyMetricsListener() kyverno.PolicyListener {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kyverno_policy",
		Help: "List of all Policies",
	}, []string{"namespace", "kind", "policy", "rule", "type", "background", "severity", "category", "validationFailureAction"})

	prometheus.Register(policyGauge)

	return func(event kyverno.LifecycleEvent) {
		switch event.Type {
		case kyverno.Added:
			for _, rule := range event.NewPolicy.Rules {
				policyGauge.With(generateResultLabels(event.NewPolicy, rule)).Set(1)
			}
		case kyverno.Updated:
			for _, rule := range event.OldPolicy.Rules {
				policyGauge.Delete(generateResultLabels(event.OldPolicy, rule))
			}

			for _, rule := range event.NewPolicy.Rules {
				policyGauge.With(generateResultLabels(event.NewPolicy, rule)).Set(1)
			}
		case kyverno.Deleted:
			for _, rule := range event.NewPolicy.Rules {
				policyGauge.Delete(generateResultLabels(event.NewPolicy, rule))
			}
		}
	}
}

func generateResultLabels(policy *kyverno.Policy, rule *kyverno.Rule) prometheus.Labels {
	labels := prometheus.Labels{
		"namespace":               policy.Namespace,
		"kind":                    policy.Kind,
		"policy":                  policy.Name,
		"severity":                policy.Severity,
		"category":                policy.Category,
		"background":              strconv.FormatBool(policy.Background),
		"validationFailureAction": policy.ValidationFailureAction,
		"rule":                    rule.Name,
		"type":                    rule.Type,
	}

	return labels
}
