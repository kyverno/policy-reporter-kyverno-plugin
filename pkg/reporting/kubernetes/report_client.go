package kubernetes

import (
	"context"

	v1alpha2 "github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	pr "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var ReportLabel = map[string]string{
	"app.kubernetes.io/managed-by": "kyverno",
}

type ReportClient struct {
	client pr.Wgpolicyk8sV1alpha2Interface
}

func NewReportClient(client pr.Wgpolicyk8sV1alpha2Interface) *ReportClient {
	return &ReportClient{client}
}

func (c *ReportClient) PolicyReports(ctx context.Context) ([]v1alpha2.PolicyReport, error) {
	list, err := c.client.PolicyReports("").List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(ReportLabel)})
	if err != nil {
		return make([]v1alpha2.PolicyReport, 0, 0), err
	}

	return list.Items, nil
}

func (c *ReportClient) ClusterPolicyReports(ctx context.Context) ([]v1alpha2.ClusterPolicyReport, error) {
	list, err := c.client.ClusterPolicyReports().List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(ReportLabel)})
	if err != nil {
		return make([]v1alpha2.ClusterPolicyReport, 0, 0), err
	}

	return list.Items, nil
}
