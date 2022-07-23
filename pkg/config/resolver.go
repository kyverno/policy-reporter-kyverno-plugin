package config

import (
	"time"

	"github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	k8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno/listener"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/leaderelection"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
	prk8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
	vk8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation/kubernetes"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Resolver manages dependencies
type Resolver struct {
	config       *Config
	clientset    *kubernetes.Clientset
	k8sConfig    *rest.Config
	mapper       k8s.Mapper
	leaderClient *leaderelection.Client
	policyStore  *kyverno.PolicyStore
	policyClient kyverno.PolicyClient
	eventClient  violation.EventClient
	polrClient   policyreport.Client
	publisher    *kyverno.EventPublisher
	vPulisher    *violation.Publisher
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
func (r *Resolver) EventPublisher() *kyverno.EventPublisher {
	if r.publisher != nil {
		return r.publisher
	}

	r.publisher = kyverno.NewEventPublisher()

	return r.publisher
}

// EventPublisher resolver method
func (r *Resolver) ViolationPublisher() *violation.Publisher {
	if r.vPulisher != nil {
		return r.vPulisher
	}

	r.vPulisher = violation.NewPublisher()

	return r.vPulisher
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

	policyClient := k8s.NewClient(client, r.Mapper(), r.EventPublisher())

	r.policyClient = policyClient

	return policyClient, nil
}

// EventClient resolver method
func (r *Resolver) Clientset() (*kubernetes.Clientset, error) {
	if r.clientset != nil {
		return r.clientset, nil
	}

	clientset, err := kubernetes.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	r.clientset = clientset

	return r.clientset, nil
}

// LeaderElectionClient resolver method
func (r *Resolver) LeaderElectionClient() (*leaderelection.Client, error) {
	if r.leaderClient != nil {
		return r.leaderClient, nil
	}

	clientset, err := r.Clientset()
	if err != nil {
		return nil, err
	}

	r.leaderClient = leaderelection.New(
		clientset.CoordinationV1(),
		r.config.LeaderElection.LockName,
		r.config.LeaderElection.Namespace,
		r.config.LeaderElection.PodName,
		time.Duration(r.config.LeaderElection.LeaseDuration)*time.Second,
		time.Duration(r.config.LeaderElection.RenewDeadline)*time.Second,
		time.Duration(r.config.LeaderElection.RetryPeriod)*time.Second,
		r.config.LeaderElection.ReleaseOnCancel,
	)

	return r.leaderClient, nil
}

// EventClient resolver method
func (r *Resolver) EventClient() (violation.EventClient, error) {
	if r.eventClient != nil {
		return r.eventClient, nil
	}

	clientset, err := r.Clientset()
	if err != nil {
		return nil, err
	}

	r.eventClient = vk8s.NewClient(clientset, r.ViolationPublisher(), r.PolicyStore(), r.config.BlockReports.EventNamespace)

	return r.eventClient, nil
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

	policyreportClient := prk8s.NewClient(
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
