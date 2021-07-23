package metrics

import (
	"strconv"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

// CreatePolicyMetricsCallback for PolicyReport watch.Events
func CreatePolicyMetricsCallback() kyverno.PolicyCallback {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report_kyverno_policy",
		Help: "List of all Policies",
	}, []string{"namespace", "kind", "policy", "rule", "type", "background", "severity", "category", "validationFailureAction"})

	prometheus.Register(policyGauge)

	return func(event watch.EventType, policy kyverno.Policy, oldPolicy kyverno.Policy) {
		switch event {
		case watch.Added:
			for _, rule := range policy.Rules {
				policyGauge.With(generateResultLabels(policy, rule)).Set(1)
			}
		case watch.Modified:
			for _, rule := range oldPolicy.Rules {
				policyGauge.Delete(generateResultLabels(oldPolicy, rule))
			}

			for _, rule := range policy.Rules {
				policyGauge.With(generateResultLabels(policy, rule)).Set(1)
			}
		case watch.Deleted:
			for _, rule := range policy.Rules {
				policyGauge.Delete(generateResultLabels(policy, rule))
			}
		}
	}
}

func generateResultLabels(policy kyverno.Policy, rule kyverno.Rule) prometheus.Labels {
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
