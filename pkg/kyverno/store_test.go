package kyverno_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kubernetes"
	report "github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/kyverno"
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

func NewPolicy() report.Policy {
	obj := unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, &obj)

	return kubernetes.NewMapper().MapPolicy(obj.Object)
}

func Test_PolicyStore(t *testing.T) {
	store := report.NewPolicyStore()
	pol := NewPolicy()

	t.Run("Add/Get", func(t *testing.T) {
		_, ok := store.Get(pol.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(pol)
		_, ok = store.Get(pol.GetIdentifier())
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
		_, ok := store.Get(pol.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(pol.GetIdentifier())
		_, ok = store.Get(pol.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})
}
