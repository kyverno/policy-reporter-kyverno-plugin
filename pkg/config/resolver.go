package config

import (
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config       *Config
	k8sConfig    *rest.Config
	mapper       kubernetes.Mapper
	policyStore  *kyverno.PolicyStore
	policyClient kyverno.PolicyClient
}

// APIServer resolver method
func (r *Resolver) APIServer() api.Server {
	return api.NewServer(
		r.PolicyStore(),
		r.config.API.Port,
	)
}

// PolicyStore resolver method
func (r *Resolver) PolicyStore() *kyverno.PolicyStore {
	if r.policyStore != nil {
		return r.policyStore
	}

	r.policyStore = kyverno.NewPolicyStore()

	return r.policyStore
}

// PolicyClient resolver method
func (r *Resolver) PolicyClient() (kyverno.PolicyClient, error) {
	if r.policyClient != nil {
		return r.policyClient, nil
	}

	policyAPI, err := r.policyReportAPI()
	if err != nil {
		return nil, err
	}

	client := kubernetes.NewPolicyClient(
		policyAPI,
		r.PolicyStore(),
		r.Mapper(),
	)

	r.policyClient = client

	return client, nil
}

// Mapper resolver method
func (r *Resolver) Mapper() kubernetes.Mapper {
	if r.mapper != nil {
		return r.mapper
	}

	r.mapper = kubernetes.NewMapper()

	return r.mapper
}

func (r *Resolver) policyReportAPI() (kubernetes.PolicyAdapter, error) {
	client, err := dynamic.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewPolicyAdapter(client, ""), nil
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}
