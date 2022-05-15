package policyreport_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
)

func Test_GeneratePolicyReportName(t *testing.T) {
	name := policyreport.GeneratePolicyReportName("test")
	if name != "polr-ns-test-blocked" {
		t.Errorf("expected PolicyReport name was 'polr-ns-test-blocked', got %s", name)
	}
}

func Test_GeneratePolicyReportResultID(t *testing.T) {
	id := policyreport.GeneratePolicyReportResultID("4baaf7cc-4f7c-4746-b8e3-1dd7cc002c75", time.Date(2022, time.May, 15, 13, 0, 0, 0, time.UTC))
	if id != "44b7690531491c5370899a1178555c20b98f836d" {
		t.Errorf("expected PolicyReportResult ID was '44b7690531491c5370899a1178555c20b98f836d', got %s", id)
	}

	id2 := policyreport.GeneratePolicyReportResultID("4baaf7cc-4f7c-4746-b8e3-1dd7cc002c75", time.Date(2022, time.May, 15, 14, 0, 0, 0, time.UTC))
	if id == id2 {
		t.Error("expected PolicyReportResult ID changed by different time", id)
	}
}
