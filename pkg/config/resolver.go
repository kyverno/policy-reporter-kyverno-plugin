package config

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/api"
	v1 "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/typed/kyverno/v1"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno"
	k8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/kyverno/listener"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/leaderelection"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport"
	prk8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/policyreport/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting"
	rk8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/reporting/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/secrets"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
	vk8s "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation/kubernetes"
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
	logger       *zap.Logger
}

// SecretClient resolver method
func (r *Resolver) SecretClient() (secrets.Client, error) {
	clientset, err := r.Clientset()
	if err != nil {
		zap.L().Error("failed to create secret client, secretRefs can not be resolved", zap.Error(err))
		return nil, err
	}

	return secrets.NewClient(clientset.CoreV1().Secrets(r.config.Namespace)), nil
}

// APIServer resolver method
func (r *Resolver) APIServer(ctx context.Context, synced func() bool) api.Server {
	var logger *zap.Logger
	if r.config.API.Logging {
		logger, _ = r.Logger()
	}

	authConfig := &r.config.API.BasicAuth
	if authConfig.SecretRef != "" {
		r.loadSecretRef(ctx, authConfig)
	}

	var auth *api.BasicAuth
	if authConfig.Username != "" && authConfig.Password != "" {
		auth = &api.BasicAuth{
			Username: authConfig.Username,
			Password: authConfig.Password,
		}

		zap.L().Info("API BasicAuth enabled")
	}

	return api.NewServer(
		r.PolicyStore(),
		r.Reporting(),
		r.config.API.Port,
		synced,
		auth,
		logger,
	)
}

func (r *Resolver) CRDMetadataClient() (metadata.Interface, error) {
	client, err := metadata.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
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

	client, err := r.CRDMetadataClient()
	if err != nil {
		return nil, err
	}

	queue, err := r.Queue()
	if err != nil {
		return nil, err
	}

	policyClient := k8s.NewClient(client, queue)

	r.policyClient = policyClient

	return policyClient, nil
}

// EventPublisher resolver method
func (r *Resolver) Queue() (*k8s.Queue, error) {
	client, err := r.CRDClient()
	if err != nil {
		return nil, err
	}

	return k8s.NewQueue(
		r.EventPublisher(),
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "policy-queue"),
		client,
		dynamic.NewForConfigOrDie(r.k8sConfig),
	), nil
}

func (r *Resolver) CRDClient() (v1.KyvernoV1Interface, error) {
	client, err := v1.NewForConfig(r.k8sConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
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

// Reporting resolver method
func (r *Resolver) Reporting() reporting.PolicyReportGenerator {
	return reporting.NewPolicyReportGenerator(
		rk8s.NewPolicyClient(v1.NewForConfigOrDie(r.k8sConfig)),
		rk8s.NewReportClient(v1alpha2.NewForConfigOrDie(r.k8sConfig)),
	)
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
		r.config.Namespace,
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

// Logger resolver method
func (r *Resolver) Logger() (*zap.Logger, error) {
	if r.logger != nil {
		return r.logger, nil
	}

	encoder := zap.NewProductionEncoderConfig()
	if r.config.Logging.Development {
		encoder = zap.NewDevelopmentEncoderConfig()
		encoder.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	ouput := "json"
	if r.config.Logging.Encoding != "json" {
		ouput = "console"
		encoder.EncodeCaller = nil
	}

	var sampling *zap.SamplingConfig
	if !r.config.Logging.Development {
		sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(r.config.Logging.LogLevel)),
		Development:       r.config.Logging.Development,
		Sampling:          sampling,
		Encoding:          ouput,
		EncoderConfig:     encoder,
		DisableStacktrace: !r.config.Logging.Development,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	r.logger = logger

	zap.ReplaceGlobals(logger)

	return r.logger, nil
}

// RegisterStoreListener resolver method
func (r *Resolver) RegisterStoreListener() {
	r.EventPublisher().RegisterListener(listener.NewStoreListener(r.PolicyStore()))
}

// RegisterMetricsListener resolver method
func (r *Resolver) RegisterMetricsListener() {
	r.EventPublisher().RegisterListener(listener.NewPolicyMetricsListener())
}

func (r *Resolver) loadSecretRef(ctx context.Context, auth *BasicAuth) {
	client, err := r.SecretClient()
	if err != nil {
		return
	}
	values, err := client.Get(ctx, auth.SecretRef)
	if err != nil {
		zap.L().Error("failed to load basic auth secret", zap.Error(err))
	}

	if values.Username != "" {
		auth.Username = values.Username
	}
	if values.Password != "" {
		auth.Password = values.Password
	}
}

// NewResolver constructor function
func NewResolver(config *Config, k8sConfig *rest.Config) Resolver {
	return Resolver{
		config:    config,
		k8sConfig: k8sConfig,
	}
}
