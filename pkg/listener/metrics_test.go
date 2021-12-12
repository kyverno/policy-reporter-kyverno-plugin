package listener_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/listener"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func Test_PolicyMetricGeneration(t *testing.T) {
	pol1 := NewPolicy()
	pol2 := NewPolicy()
	pol2.Name = pol2.Name + "-updated"

	handler := listener.NewPolicyMetricsListener()

	t.Run("Added Metric", func(t *testing.T) {
		handler(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: pol1, OldPolicy: &kyverno.Policy{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "kyverno_policy")

		metricResult := results.GetMetric()
		if err = testResultMetricLabels(metricResult[0], pol1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: pol1, OldPolicy: &kyverno.Policy{}})
		handler(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: pol2, OldPolicy: pol1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "kyverno_policy")

		metricResult := results.GetMetric()
		if err = testResultMetricLabels(metricResult[0], pol2); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(kyverno.LifecycleEvent{Type: kyverno.Added, NewPolicy: pol1, OldPolicy: &kyverno.Policy{}})
		handler(kyverno.LifecycleEvent{Type: kyverno.Updated, NewPolicy: pol2, OldPolicy: pol1})
		handler(kyverno.LifecycleEvent{Type: kyverno.Deleted, NewPolicy: pol2, OldPolicy: &kyverno.Policy{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "kyverno_policy")
		if results != nil {
			t.Error("policy_report_kyverno_policy shoud no longer exist", *results.Name)
		}
	})
}

func testResultMetricLabels(metric *io_prometheus_client.Metric, policy *kyverno.Policy) error {
	rule := policy.Rules[0]

	if name := *metric.Label[0].Name; name != "background" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != strconv.FormatBool(policy.Background) {
		return fmt.Errorf("Unexpected Background Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "category" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != policy.Category {
		return fmt.Errorf("Unexpected Category Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "kind" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != policy.Kind {
		return fmt.Errorf("Unexpected Kind Label Value: %s", value)
	}

	if name := *metric.Label[3].Name; name != "namespace" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[3].Value; value != policy.Namespace {
		return fmt.Errorf("Unexpected Namespace Label Value: %s", value)
	}

	if name := *metric.Label[4].Name; name != "policy" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[4].Value; value != policy.Name {
		return fmt.Errorf("Unexpected Policy Label Value: %s", value)
	}

	if name := *metric.Label[5].Name; name != "rule" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[5].Value; value != rule.Name {
		return fmt.Errorf("Unexpected Rule Label Value: %s", value)
	}

	if name := *metric.Label[6].Name; name != "severity" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[6].Value; value != policy.Severity {
		return fmt.Errorf("Unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[7].Name; name != "type" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[7].Value; value != rule.Type {
		return fmt.Errorf("Unexpected Type Label Value: %s", value)
	}

	if name := *metric.Label[8].Name; name != "validationFailureAction" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[8].Value; value != policy.ValidationFailureAction {
		return fmt.Errorf("Unexpected ValidationFailureAction Label Value: %s", value)
	}

	return nil
}

func findMetric(metrics []*io_prometheus_client.MetricFamily, name string) *io_prometheus_client.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
