package kubernetes

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kyverno "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/kyverno/v1"
	pr "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/typed/kyverno/v1"
)

type PolicyClient struct {
	client pr.KyvernoV1Interface
}

func NewPolicyClient(client pr.KyvernoV1Interface) *PolicyClient {
	return &PolicyClient{client}
}

func (p *PolicyClient) CusterPolicies(ctx context.Context) ([]kyverno.ClusterPolicy, error) {
	list, err := p.client.ClusterPolicies().List(ctx, v1.ListOptions{})
	if err != nil {
		return make([]kyverno.ClusterPolicy, 0, 0), err
	}

	return list.Items, nil
}

func (p *PolicyClient) Policies(ctx context.Context, namespace string) ([]kyverno.Policy, error) {
	list, err := p.client.Policies(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return make([]kyverno.Policy, 0, 0), err
	}

	return list.Items, nil
}
