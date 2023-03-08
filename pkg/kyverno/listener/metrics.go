package listener

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
)

// NewPolicyMetricsListener for Policy kyverno.LifecycleEvent
func NewPolicyMetricsListener() kyverno.PolicyListener {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kyverno_policy",
		Help: "List of all Policies",
	}, []string{"namespace", "kind", "policy", "rule", "type", "background", "severity", "category", "validationFailureAction"})

	prometheus.Register(policyGauge)
	cache := NewCache()

	return func(event kyverno.LifecycleEvent) {
		switch event.Type {
		case kyverno.Added:
			for _, rule := range event.Policy.Rules {
				policyGauge.With(generateResultLabels(event.Policy, rule)).Set(1)
			}

			cache.Add(event.Policy)
		case kyverno.Updated:
			items := cache.GetReportLabels(event.Policy.GetID())
			for _, rule := range items {
				policyGauge.Delete(rule.Labels)
			}

			for _, rule := range event.Policy.Rules {
				policyGauge.With(generateResultLabels(event.Policy, rule)).Set(1)
			}

			cache.Add(event.Policy)
		case kyverno.Deleted:
			items := cache.GetReportLabels(event.Policy.GetID())

			for _, rule := range items {
				policyGauge.Delete(rule.Labels)
			}

			cache.Remove(event.Policy.GetID())
		}
	}
}

func generateResultLabels(policy kyverno.Policy, rule *kyverno.Rule) prometheus.Labels {
	labels := prometheus.Labels{
		"namespace":               policy.Namespace,
		"kind":                    policy.Kind,
		"policy":                  policy.Name,
		"severity":                policy.Severity,
		"category":                policy.Category,
		"background":              "",
		"validationFailureAction": policy.ValidationFailureAction,
		"rule":                    rule.Name,
		"type":                    rule.Type,
	}

	if policy.Background != nil {
		labels["background"] = strconv.FormatBool(*policy.Background)
	}

	return labels
}
