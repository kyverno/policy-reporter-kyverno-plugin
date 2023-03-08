package policyreport

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
)

const (
	ClusterPolicyReport = "kyverno-cpolr-blocked"
)

func GeneratePolicyReportName(namespace string) string {
	return fmt.Sprintf("polr-ns-%s-blocked", namespace)
}

func GeneratePolicyReportResultID(eventID string, lastTimestamp time.Time) string {
	id := fmt.Sprintf("%s_%d", eventID, lastTimestamp.Unix())

	h := sha1.New()
	h.Write([]byte(id))

	return fmt.Sprintf("%x", h.Sum(nil))
}

type Client interface {
	ProcessViolation(context.Context, violation.PolicyViolation) error
}
