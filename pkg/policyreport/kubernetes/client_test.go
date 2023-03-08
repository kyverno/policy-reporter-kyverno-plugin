package kubernetes_test

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
)

func NewPolicyReportFakeCilent() (v1alpha2client.Wgpolicyk8sV1alpha2Interface, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset().Wgpolicyk8sV1alpha2()

	return client, client.PolicyReports("test"), client.ClusterPolicyReports()
}

func Test_CreateNewPolicyReportForViolation(t *testing.T) {
	client, polrAPI, _ := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", false)
	ctx := context.Background()

	violation := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: time.Now(),
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.GeneratePolicyReportName("test"), v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 1 {
		t.Errorf("expected one result in the created PolicyReport")
	}

	checkResource(polr.Results[0], violation, t)
}

func Test_UpdateExistingPolicyReportWithoutKeepOnlyLatest(t *testing.T) {
	client, polrAPI, _ := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", true)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: now,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.GeneratePolicyReportName("test"), v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 1 {
		t.Errorf("expected only the latest result in the updated PolicyReport")
	}

	if polr.Results[0].Timestamp.Seconds != int64(updated.Unix()) {
		t.Errorf("expected time to be the latest updated")
	}
}

func Test_AddUpdateToPolicyReport(t *testing.T) {
	client, polrAPI, _ := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", false)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: now,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.GeneratePolicyReportName("test"), v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 2 {
		t.Errorf("expected both result in the updated PolicyReport")
	}

	checkResource(polr.Results[0], violation1, t)
	checkResource(polr.Results[1], violation2, t)
}

func Test_MaxResultLimitForResults(t *testing.T) {
	client, polrAPI, _ := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 1, "Kyverno Event", false)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: now,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name:      "nginx",
			Namespace: "test",
			Kind:      "Pod",
		},
		Policy: violation.Policy{
			Name:     "request-and-limit-required",
			Rule:     "require-reesource-request",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "nginx.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.GeneratePolicyReportName("test"), v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 1 {
		t.Errorf("expected only one result for the updated PolicyReport because of the result limit")
	}

	checkResource(polr.Results[0], violation2, t)
}

func Test_CreateNewClusterPolicyReportForViolation(t *testing.T) {
	client, _, polrAPI := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", false)
	ctx := context.Background()

	violation := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: time.Now(),
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.ClusterPolicyReport, v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 1 {
		t.Errorf("expected one result in the created ClusterPolicyReport")
	}

	checkResource(polr.Results[0], violation, t)
}

func Test_UpdateExistingClusterPolicyReportWithoutKeepOnlyLatest(t *testing.T) {
	client, _, polrAPI := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", true)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: now,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation3 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "dev",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "dev.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated.Add(2 * time.Minute),
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation3)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.ClusterPolicyReport, v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 2 {
		t.Errorf("expected only the latest result in the updated ClusterPolicyReport")
	}

	if polr.Results[0].Timestamp.Seconds != int64(updated.Unix()) {
		t.Errorf("expected time to be the latest updated")
	}
}

func Test_AddUpdateToClusterPolicyReport(t *testing.T) {
	client, _, polrAPI := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 10, "Kyverno Event", false)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: now,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.ClusterPolicyReport, v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 2 {
		t.Errorf("expected both result in the updated PolicyReport")
	}

	checkResource(polr.Results[0], violation1, t)
	checkResource(polr.Results[1], violation2, t)
}

func Test_MaxResultLimitForClusterPolicyReportResults(t *testing.T) {
	client, _, polrAPI := NewPolicyReportFakeCilent()
	polrClient := kubernetes.NewClient(client, 1, "Kyverno Event", false)
	ctx := context.Background()
	now := time.Now()
	updated := now.Add(5 * time.Minute)

	violation1 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   false,
	}

	err := polrClient.ProcessViolation(ctx, violation1)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	violation2 := violation.PolicyViolation{
		Resource: violation.Resource{
			Name: "test",
			Kind: "Namespace",
		},
		Policy: violation.Policy{
			Name:     "ns-labels-required",
			Rule:     "require-team-label",
			Message:  "message",
			Category: "Best Practices",
			Severity: "medium",
		},
		Event: violation.Event{
			Name: "test.12345",
			UID:  "2d81d080-d2a3-4f1d-aad8-27c2ceb2a3fa",
		},
		Timestamp: updated,
		Updated:   true,
	}

	err = polrClient.ProcessViolation(ctx, violation2)
	if err != nil {
		t.Fatalf("Unexpected failure: %s", err)
	}

	polr, err := polrAPI.Get(ctx, policyreport.ClusterPolicyReport, v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected failure: %s", err)
	}

	if len(polr.Results) != 1 {
		t.Errorf("expected only one result for the updated PolicyReport because of the result limit")
	}

	checkResource(polr.Results[0], violation2, t)
}

func checkResource(result v1alpha2.PolicyReportResult, violation violation.PolicyViolation, t *testing.T) {
	if result.Category != violation.Policy.Category {
		t.Errorf("expected Category to be '%s', got %s", violation.Policy.Category, result.Category)
	}
	if result.Message != violation.Policy.Message {
		t.Errorf("expected Message to be '%s', got %s", violation.Policy.Category, result.Category)
	}
	if result.Policy != violation.Policy.Name {
		t.Errorf("expected Policy to be '%s', got %s", violation.Policy.Name, result.Policy)
	}
	if result.Rule != violation.Policy.Rule {
		t.Errorf("expected Rule to be '%s', got %s", violation.Policy.Rule, result.Rule)
	}
	if string(result.Severity) != violation.Policy.Severity {
		t.Errorf("expected Severity to be '%s', got %s", violation.Policy.Severity, result.Severity)
	}

	resource := result.Resources[0]
	if resource.Name != violation.Resource.Name {
		t.Errorf("expected Resource.Name to be '%s', got %s", violation.Resource.Name, resource.Name)
	}
	if resource.Kind != violation.Resource.Kind {
		t.Errorf("expected Resource.Kind to be '%s', got %s", violation.Resource.Kind, resource.Kind)
	}
	if resource.Namespace != violation.Resource.Namespace {
		t.Errorf("expected Resource.Namespace to be '%s', got %s", violation.Resource.Namespace, resource.Namespace)
	}

	if result.Timestamp.Seconds != violation.Timestamp.Unix() {
		t.Errorf("expected Timestamp to be '%s', got %s", violation.Policy.Rule, result.Rule)
	}
}
