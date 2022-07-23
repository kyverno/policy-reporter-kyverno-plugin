package kubernetes_test

import (
	"strings"
	"testing"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno/kubernetes"
)

func Test_MapPolicy(t *testing.T) {
	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(convert(clusterPolicy).Object)

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

func Test_MapMinClusterPolicy(t *testing.T) {
	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(convert(minClusterPolicy).Object)

	if pol.Kind != "ClusterPolicy" {
		t.Errorf("Expected Kind 'Policy', go %s", pol.Kind)
	}
	if pol.Name != "disallow-host-path" {
		t.Errorf("Expected Name 'disallow-host-path', go %s", pol.Name)
	}
	if pol.Background {
		t.Errorf("Expected Background 'false', go true")
	}
	if pol.UID != "953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7" {
		t.Errorf("Expected UID '953b1167-1ff5-4cf6-b636-3b7d0c0dd6c7', go %s", pol.UID)
	}

	rule := pol.Rules[0]
	if rule.Type != "validation" {
		t.Errorf("Expected Rule Type 'validation', got %s", rule.Type)
	}
	if rule.Name != "" {
		t.Errorf("Expected Rule Name is empty, got %s", rule.Name)
	}
	if rule.ValidateMessage != "" {
		t.Errorf("Expected empty Rule Message, got %s", rule.ValidateMessage)
	}
}

func Test_MapGeneratePolicy(t *testing.T) {
	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(convert(genPolicy).Object)

	rule := pol.Rules[0]
	if rule.Type != "generation" {
		t.Errorf("Expected Rule Type 'generation', got %s", rule.Type)
	}
}

func Test_MapMutatePolicy(t *testing.T) {
	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(convert(mutPolicy).Object)

	rule := pol.Rules[0]
	if rule.Type != "mutation" {
		t.Errorf("Expected Rule Type 'generation', got %s", rule.Type)
	}
}

func Test_MapVerifyImagePolicy(t *testing.T) {
	mapper := kubernetes.NewMapper()

	pol := mapper.MapPolicy(convert(veriyPolicy).Object)

	rule := pol.Rules[0]

	if len(rule.VerifyImages) == 0 {
		t.Fatalf("Expected VerifyImages information to be parsed")
	}

	verify := rule.VerifyImages[0]
	if verify.Repository != "registry.io/signatures" {
		t.Errorf("Expected Repo to be 'registry.io/signatures', got %s", verify.Repository)
	}
	if verify.Image != "ghcr.io/kyverno/test-verify-image:*" {
		t.Errorf("Expected Image to be 'ghcr.io/kyverno/test-verify-image:*', got %s", verify.Image)
	}
	if verify.Attestations == "" {
		t.Errorf("Expected Attestations not be empty")
	}

	key := strings.TrimSpace(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8nXRh950IZbRj8Ra/N9sbqOPZrfM
5/KAQN0/KjHcorm/J5yctVd7iEcnessRQjU917hmKO6JWVGHpDguIyakZA==
-----END PUBLIC KEY-----
  `)

	if verify.Key != key {
		t.Errorf("Expected Key to be \n'%s', got \n'%s'", key, verify.Key)
	}
}
