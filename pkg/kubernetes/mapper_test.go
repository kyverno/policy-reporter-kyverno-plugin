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

func Test_MapPolicy(t *testing.T) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, obj)

	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(obj.Object)

	if pol.Kind != "ClusterPolicy" {
		t.Errorf("Expected Kind 'ClusterPolicy', go %s", pol.Kind)
	}
	if pol.Name != "disallow-host-path" {
		t.Errorf("Expected Name 'disallow-host-path', go %s", pol.Name)
	}
	if pol.Category != "Pod Security Standards (Default)" {
		t.Errorf("Expected Category 'Pod Security Standards (Default)', go %s", pol.Category)
	}
	if pol.Severity != "medium" {
		t.Errorf("Expected Severity 'medium', go %s", pol.Severity)
	}
	if len(pol.AutogenControllers) != 1 && pol.AutogenControllers[0] != "Deploymemt" {
		t.Errorf("Expected 1 Autogen 'Deployment', go %s", strings.Join(pol.AutogenControllers, ", "))
	}
	if !pol.Background {
		t.Errorf("Expected Background 'true', go false")
	}
	if pol.ValidationFailureAction != "audit" {
		t.Errorf("Expected ValidationFailureAction 'audit', go %s", pol.ValidationFailureAction)
	}
	if pol.UID != "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7" {
		t.Errorf("Expected UID '953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7', go %s", pol.UID)
	}

	rule := pol.Rules[0]
	if rule.Type != "validation" {
		t.Errorf("Expected Rule Type 'validation', go %s", rule.Type)
	}
	if rule.Name != "host-path" {
		t.Errorf("Expected Rule Name 'host-path', go %s", rule.Name)
	}
	if rule.ValidateMessage != "HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set." {
		t.Errorf("Expected Rule Message 'HostPath volumes are forbidden. The fields spec.volumes[*].hostPath must not be set.', go %s", rule.ValidateMessage)
	}
}
