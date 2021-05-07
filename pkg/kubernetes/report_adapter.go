package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type PolicyAdapter interface {
	ListClusterPolicies() (*unstructured.UnstructuredList, error)
	ListPolicies() (*unstructured.UnstructuredList, error)
	WatchClusterPolicies() (watch.Interface, error)
	WatchPolicies() (watch.Interface, error)
}

type k8sPolicyAdapter struct {
	client          dynamic.Interface
	policies        schema.GroupVersionResource
	clusterPolicies schema.GroupVersionResource
}

func (k *k8sPolicyAdapter) ListClusterPolicies() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(k.clusterPolicies).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyAdapter) ListPolicies() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(k.policies).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyAdapter) WatchClusterPolicies() (watch.Interface, error) {
	return k.client.Resource(k.clusterPolicies).Watch(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyAdapter) WatchPolicies() (watch.Interface, error) {
	return k.client.Resource(k.policies).Watch(context.Background(), metav1.ListOptions{})
}

// NewPolicAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyAdapter(dynamic dynamic.Interface, version string) PolicyAdapter {
	if version == "" {
		version = "v1"
	}

	return &k8sPolicyAdapter{
		client: dynamic,
		policies: schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  version,
			Resource: "clusterpolicies",
		},
		clusterPolicies: schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  version,
			Resource: "policies",
		},
	}
}
