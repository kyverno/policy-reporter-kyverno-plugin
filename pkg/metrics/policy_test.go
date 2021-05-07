package metrics_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/watch"
)

var policy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  annotations:
    pod-policies.kyverno.io/autogen-controllers: Deployment
    meta.helm.sh/release-name: kyverno
    meta.helm.sh/release-namespace: kyverno
    policies.kyverno.io/severity: medium
    policies.kyverno.io/category: Pod Security Standards (Default)
    policies.kyverno.io/description: HostPath volumes let pods use host directories
      and volumes in containers. Using host resources can be used to access shared
      data or escalate privileges and should not be allowed.
  creationTimestamp: "2021-03-31T13:42:01Z"
  generation: 1
  labels:
    app.kubernetes.io/managed-by: Helm
  name: disallow-host-path
  resourceVersion: "61655872"
  uid: 953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7
spec:
  background: true
  rules:
  - match:
      resources:
        kinds:
        - Pod
    name: host-path
    validate:
      message: HostPath volumes are forbidden. The fields spec.volumes[*].hostPath
        must not be set.
      pattern:
        spec:
          =(volumes):
          - X(hostPath): "null"
  validationFailureAction: audit
`

func NewPolicy() kyverno.Policy {
	obj := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, &obj)

	return kubernetes.NewMapper().MapPolicy(obj.Object)
}

func Test_PolicyMetricGeneration(t *testing.T) {
	pol1 := NewPolicy()
	pol2 := NewPolicy()
	pol2.Name = pol2.Name + "-updated"

	handler := metrics.CreatePolicyMetricsCallback()

	t.Run("Added Metric", func(t *testing.T) {
		handler(watch.Added, pol1, kyverno.Policy{})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_kyverno_policy")

		metricResult := results.GetMetric()
		if err = testResultMetricLabels(metricResult[0], pol1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(watch.Added, pol1, kyverno.Policy{})
		handler(watch.Modified, pol2, pol1)

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_kyverno_policy")

		metricResult := results.GetMetric()
		if err = testResultMetricLabels(metricResult[0], pol2); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(watch.Added, pol1, kyverno.Policy{})
		handler(watch.Modified, pol2, pol1)
		handler(watch.Deleted, pol2, pol2)

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_kyverno_policy")
		if results != nil {
			t.Error("policy_report_kyverno_policy shoud no longer exist", *results.Name)
		}
	})
}

func testResultMetricLabels(metric *io_prometheus_client.Metric, policy kyverno.Policy) error {
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
