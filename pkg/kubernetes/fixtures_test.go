package kubernetes_test

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

var clusterPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  annotations:
    pod-policies.kyverno.io/autogen-controllers: Deployment
    policies.kyverno.io/severity: medium
    policies.kyverno.io/category: Pod Security Standards (Default)
    policies.kyverno.io/description: HostPath volumes let pods use host directories
      and volumes in containers. Using host resources can be used to access shared
      data or escalate privileges and should not be allowed.
  creationTimestamp: "2021-03-31T13:42:01Z"
  name: disallow-host-path
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

var minClusterPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  creationTimestamp: "2021-03-31T13:42:01Z"
  name: disallow-host-path
  uid: 953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7
spec:
  rules:
  - match:
      resources:
        kinds:
        - Pod
    validate:
      message:
      pattern:
        spec:
          =(volumes):
          - X(hostPath): "null"
`

var nsPolicy = `
apiVersion: kyverno.io/v1
kind: Policy
metadata:
  creationTimestamp: "2021-03-31T13:42:01Z"
  name: disallow-host-path
  resourceVersion: "61655872"
  uid: "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7"
  namespace: "test"
spec:
  background: true
  rules:
  - match:
      resources:
        kinds:
        - Pod
    validate:
      message:
      pattern:
        spec:
          =(volumes):
          - X(hostPath): "null"
          rules:
          - match:
              resources:
                kinds:
                - Pod
            validate:
              message:
              pattern:
                spec:
                  =(volumes):
                  - X(hostPath): "null"
  - match:
      resources:
        kinds:
        - Pod
`

var minPolicy = `
apiVersion: kyverno.io/v1
kind: Policy
metadata:
  creationTimestamp: "2021-03-31T13:42:01Z"
  name: disallow-host-path
  namespace: test
  uid: 953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7
spec:
  background: true
  rules:
  - match:
      resources:
        kinds:
        - Pod
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

var veriyPolicy = `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: check-image
spec:
  validationFailureAction: enforce
  background: false
  webhookTimeoutSeconds: 30
  failurePolicy: Fail
  rules:
    - name: check-image
      match:
        resources:
          kinds:
            - Pod
      verifyImages:
      - image: "ghcr.io/kyverno/test-verify-image:*"
        repository: "registry.io/signatures"
        key: |-
          -----BEGIN PUBLIC KEY-----
          MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8nXRh950IZbRj8Ra/N9sbqOPZrfM
          5/KAQN0/KjHcorm/J5yctVd7iEcnessRQjU917hmKO6JWVGHpDguIyakZA==
          -----END PUBLIC KEY-----
        attestations:
          - predicateType: https://example.com/CodeReview/v1
            conditions:
              - all:
                - key: "{{ repo.uri }}"
                  operator: Equals
                  value: "https://git-repo.com/org/app"            
                - key: "{{ repo.branch }}"
                  operator: Equals
                  value: "main"
                - key: "{{ reviewers }}"
                  operator: In
                  value: ["ana@example.com", "bob@example.com"]
`

func convert(policy string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dec.Decode([]byte(policy), nil, obj)

	return obj
}
