package kubernetes

import (
	"fmt"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	pr "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	reportLabels = map[string]string{
		"managed-by": "policy-reporter-kyverno-plugin",
	}
)

type policyReportClient struct {
	client         pr.Wgpolicyk8sV1alpha2Interface
	maxResults     int
	source         string
	keepOnlyLatest bool
}

func (p *policyReportClient) ProcessViolation(ctx context.Context, violation violation.PolicyViolation) error {
	if violation.Resource.Namespace == "" {
		return p.handleClusterScoped(ctx, violation)
	}

	return p.handleNamespaced(ctx, violation, violation.Resource.Namespace)
}

func (p *policyReportClient) handleNamespaced(ctx context.Context, violation violation.PolicyViolation, ns string) error {
	polr, err := p.client.PolicyReports(ns).Get(ctx, policyreport.GeneratePolicyReportName(ns), v1.GetOptions{})
	if err != nil {
		polr = &v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:      policyreport.GeneratePolicyReportName(ns),
				Namespace: ns,
				Labels:    reportLabels,
			},
		}

		polr, err = p.client.PolicyReports(ns).Create(ctx, polr, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create PolicyReport in namespace %s: %s", ns, err)
		}
	}

	if polr.Results == nil {
		polr.Results = []v1alpha2.PolicyReportResult{}
	}
	if len(polr.Results) >= p.maxResults {
		startIndex := len(polr.Results) - p.maxResults + 1

		polr.Summary.Fail--
		polr.Results = polr.Results[startIndex:]
	}

	if violation.Updated && p.keepOnlyLatest {
		index := prevIndex(polr.Results, violation)
		if index >= 0 {
			polr.Results = append(polr.Results[:index], polr.Results[index+1:]...)
			polr.Summary.Fail--
		}
	}

	polr.Summary.Fail++
	polr.Results = append(polr.Results, buildResult(violation, p.source))

	_, err = p.client.PolicyReports(ns).Update(ctx, polr, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update PolicyReport in namespace %s: %s", ns, err)
	}

	return nil
}

func (p *policyReportClient) handleClusterScoped(ctx context.Context, violation violation.PolicyViolation) error {
	polr, err := p.client.ClusterPolicyReports().Get(ctx, policyreport.ClusterPolicyReport, v1.GetOptions{})
	if err != nil {
		polr = &v1alpha2.ClusterPolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:   policyreport.ClusterPolicyReport,
				Labels: reportLabels,
			},
		}

		polr, err = p.client.ClusterPolicyReports().Create(ctx, polr, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ClusterPolicyReport: %s", err)
		}
	}

	if polr.Results == nil {
		polr.Results = []v1alpha2.PolicyReportResult{}
	}
	if len(polr.Results) >= p.maxResults {
		startIndex := len(polr.Results) - p.maxResults + 1

		polr.Summary.Fail--
		polr.Results = polr.Results[startIndex:]
	}

	if violation.Updated && p.keepOnlyLatest {
		index := prevIndex(polr.Results, violation)
		if index >= 0 {
			polr.Results = append(polr.Results[:index], polr.Results[index+1:]...)
			polr.Summary.Fail--
		}
	}

	polr.Summary.Fail++
	polr.Results = append(polr.Results, buildResult(violation, p.source))

	_, err = p.client.ClusterPolicyReports().Update(ctx, polr, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ClusterPolicyReport: %s", err)
	}

	return nil
}

func buildResult(violation violation.PolicyViolation, source string) v1alpha2.PolicyReportResult {
	return v1alpha2.PolicyReportResult{
		Source:   source,
		Policy:   violation.Policy.Name,
		Rule:     violation.Policy.Rule,
		Category: violation.Policy.Category,
		Severity: v1alpha2.PolicySeverity(violation.Policy.Severity),
		Message:  violation.Policy.Message,
		Result:   "fail",
		Resources: []corev1.ObjectReference{
			{
				Kind:      violation.Resource.Kind,
				Namespace: violation.Resource.Namespace,
				Name:      violation.Resource.Name,
			},
		},
		Timestamp: v1.Timestamp{Seconds: violation.Timestamp.Unix()},
		Properties: map[string]string{
			"eventName": violation.Event.Name,
			"resultID":  policyreport.GeneratePolicyReportResultID(violation.Event.UID, violation.Timestamp),
		},
	}
}

func prevIndex(results []v1alpha2.PolicyReportResult, violation violation.PolicyViolation) int {
	for index, result := range results {
		if result.Properties["eventName"] == violation.Event.Name {
			return index
		}
	}

	return -1
}

func NewClient(client pr.Wgpolicyk8sV1alpha2Interface, maxResults int, source string, keepOnlyLatest bool) policyreport.Client {
	return &policyReportClient{
		client:         client,
		maxResults:     maxResults,
		source:         source,
		keepOnlyLatest: keepOnlyLatest,
	}
}
