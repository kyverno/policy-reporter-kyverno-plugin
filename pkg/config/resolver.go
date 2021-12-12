package config

import (
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/listener"
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
	publisher    kyverno.EventPublisher
}

// APIServer resolver method
func (r *Resolver) APIServer(foundResources map[string]bool) api.Server {
	return api.NewServer(
		r.PolicyStore(),
		r.config.API.Port,
		foundResources,
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

// EventPublisher resolver method
func (r *Resolver) EventPublisher() kyverno.EventPublisher {
	if r.publisher != nil {
		return r.publisher
	}

	s := kyverno.NewEventPublisher()
	r.publisher = s

	return r.publisher
}

// PolicyClient resolver method
func (r *Resolver) PolicyClient() (kyverno.PolicyClient, error) {
	if r.policyClient != nil {
		return r.policyClient, nil
	}

	client, err := dynamic.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	policyClient := kubernetes.NewPolicyClient(client, r.Mapper(), 5*time.Second)

	r.policyClient = policyClient

	return policyClient, nil
}

// Mapper resolver method
func (r *Resolver) Mapper() kubernetes.Mapper {
	if r.mapper != nil {
		return r.mapper
	}

	r.mapper = kubernetes.NewMapper()

	return r.mapper
}

// RegisterSendResultListener resolver method
func (r *Resolver) RegisterStoreListener() {
	r.EventPublisher().RegisterListener(listener.NewStoreListener(r.PolicyStore()))
}

// RegisterMetricsListener resolver method
func (r *Resolver) RegisterMetricsListener() {
	r.EventPublisher().RegisterListener(listener.NewPolicyMetricsListener())
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}
