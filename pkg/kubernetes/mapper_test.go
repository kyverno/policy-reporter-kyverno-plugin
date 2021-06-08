package kubernetes_test

import (
	"strings"
	"testing"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
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

var minPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  creationTimestamp: "2021-03-31T13:42:01Z"
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
      message:
      pattern:
        spec:
          =(volumes):
          - X(hostPath): "null"
`

var genPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  annotations:
    pod-policies.kyverno.io/autogen-controllers: DaemonSet,Deployment,Job,StatefulSet,CronJob
  creationTimestamp: "2021-03-31T14:51:14Z"
  generation: 2
  labels:
    app.kubernetes.io/managed-by: Helm
  name: prod-env-deny-all-traffic
  uid: 5c569b21-9e51-48a2-b7b1-0af0518119e0
spec:
  background: false
  rules:
  - generate:
      data:
        spec:
          podSelector: {}
          policyTypes:
          - Ingress
          - Egress
      kind: NetworkPolicy
      name: deny-all-traffic
      namespace: '{{request.object.metadata.name}}'
    match:
      resources:
        kinds:
        - Namespace
        selector:
          matchLabels:
            env: production
    name: deny-all-traffic
  validationFailureAction: audit
`

var mutPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  annotations:
    pod-policies.kyverno.io/autogen-controllers: DaemonSet,Deployment,Job,StatefulSet,CronJob
  name: add-env-label
  uid: 139bd7d1-88d9-4a3c-8f4a-705067f59ee9
spec:
  background: true
  rules:
  - match:
      resources:
        kinds:
        - Namespace
        name: prod*
    mutate:
      patchStrategicMerge:
        metadata:
          labels:
            env: production
    name: add production label
  validationFailureAction: audit
`

func Test_MapPolicy(t *testing.T) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, obj)

	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(obj.Object)

	if pol.Kind != "ClusterPolicy" {
		t.Errorf("Expected Kind 'ClusterPolicy', got %s", pol.Kind)
	}
	if pol.Name != "disallow-host-path" {
		t.Errorf("Expected Name 'disallow-host-path', got %s", pol.Name)
	}
	if pol.Category != "Pod Security Standards (Default)" {
		t.Errorf("Expected Category 'Pod Security Standards (Default)', got %s", pol.Category)
	}
	if pol.Severity != "medium" {
		t.Errorf("Expected Severity 'medium', got %s", pol.Severity)
	}
	if len(pol.AutogenControllers) != 1 && pol.AutogenControllers[0] != "Deploymemt" {
		t.Errorf("Expected 1 Autogen 'Deployment', got %s", strings.Join(pol.AutogenControllers, ", "))
	}
	if !pol.Background {
		t.Errorf("Expected Background 'true', got false")
	}
	if pol.ValidationFailureAction != "audit" {
		t.Errorf("Expected ValidationFailureAction 'audit', got %s", pol.ValidationFailureAction)
	}
	if pol.UID != "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7" {
		t.Errorf("Expected UID '953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7', got %s", pol.UID)
	}

	rule := pol.Rules[0]
	if rule.Type != "validation" {
		t.Errorf("Expected Rule Type 'validation', got %s", rule.Type)
	}
	if rule.Name != "host-path" {
		t.Errorf("Expected Rule Name 'host-path', got %s", rule.Name)
	}
	if rule.ValidateMessage != "HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set." {
		t.Errorf("Expected Rule Message 'HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set.', go %s", rule.ValidateMessage)
	}
}

func Test_MapMinPolicy(t *testing.T) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(minPolicy), nil, obj)

	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(obj.Object)

	if pol.Kind != "ClusterPolicy" {
		t.Errorf("Expected Kind 'ClusterPolicy', go %s", pol.Kind)
	}
	if pol.Name != "disallow-host-path" {
		t.Errorf("Expected Name 'disallow-host-path', go %s", pol.Name)
	}
	if !pol.Background {
		t.Errorf("Expected Background 'true', go false")
	}
	if pol.UID != "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7" {
		t.Errorf("Expected UID '953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7', go %s", pol.UID)
	}

	rule := pol.Rules[0]
	if rule.Type != "validation" {
		t.Errorf("Expected Rule Type 'validation', got %s", rule.Type)
	}
	if rule.Name != "host-path" {
		t.Errorf("Expected Rule Name 'host-path', got %s", rule.Name)
	}
	if rule.ValidateMessage != "" {
		t.Errorf("Expected empty Rule Message, got %s", rule.ValidateMessage)
	}
}

func Test_MapGeneratePolicy(t *testing.T) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(genPolicy), nil, obj)

	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(obj.Object)

	rule := pol.Rules[0]
	if rule.Type != "generation" {
		t.Errorf("Expected Rule Type 'generation', go %s", rule.Type)
	}
}

func Test_MapMutatePolicy(t *testing.T) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(mutPolicy), nil, obj)

	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(obj.Object)

	rule := pol.Rules[0]
	if rule.Type != "mutation" {
		t.Errorf("Expected Rule Type 'generation', go %s", rule.Type)
	}
}
