package config

import (
	"time"

	"github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	k8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/listener"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config       *Config
	k8sConfig    *rest.Config
	mapper       k8s.Mapper
	policyStore  *kyverno.PolicyStore
	policyClient kyverno.PolicyClient
	eventClient  kyverno.EventClient
	polrClient   policyreport.Client
	publisher    kyverno.EventPublisher
}

// APIServer resolver method
func (r *Resolver) APIServer(synced func() bool) api.Server {
	return api.NewServer(
		r.PolicyStore(),
		r.config.API.Port,
		synced,
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

	policyClient := k8s.NewPolicyClient(client, r.Mapper(), r.EventPublisher())

	r.policyClient = policyClient

	return policyClient, nil
}

// EventClient resolver method
func (r *Resolver) EventClient() (kyverno.EventClient, error) {
	if r.eventClient != nil {
		return r.eventClient, nil
	}

	clientset, err := kubernetes.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	eventClient := k8s.NewEventClient(clientset, r.PolicyStore(), 5*time.Second, r.config.BlockReports.EventNamespace)

	r.eventClient = eventClient

	return eventClient, nil
}

// PolicyReportClient resolver method
func (r *Resolver) PolicyReportClient() (policyreport.Client, error) {
	if r.polrClient != nil {
		return r.polrClient, nil
	}

	client, err := v1alpha2.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	policyreportClient := k8s.NewPolicyReportClient(
		client,
		r.config.BlockReports.Results.MaxPerReport,
		r.config.BlockReports.Source,
		r.config.BlockReports.Results.KeepOnlyLatest,
	)

	r.polrClient = policyreportClient

	return policyreportClient, nil
}

// Mapper resolver method
func (r *Resolver) Mapper() k8s.Mapper {
	if r.mapper != nil {
		return r.mapper
	}

	r.mapper = k8s.NewMapper()

	return r.mapper
}

// RegisterStoreListener resolver method
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
