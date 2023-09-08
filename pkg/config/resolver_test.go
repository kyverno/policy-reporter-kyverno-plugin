package config_test

import (
	"context"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
)

var testConfig = &config.Config{}

func Test_ResolvePolicyClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	client1, err := resolver.PolicyClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, err := resolver.PolicyClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if client1 != client2 {
		t.Error("A second call resolver.PolicyClient() should return the cached first client")
	}
}

func Test_ResolveClientset(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	client1, err := resolver.Clientset()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, err := resolver.Clientset()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if client1 != client2 {
		t.Error("A second call resolver.Clientset() should return the cached first client")
	}
}

func Test_ResolveLeaderElection(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	client1, err := resolver.LeaderElectionClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, err := resolver.LeaderElectionClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if client1 != client2 {
		t.Error("A second call resolver.LeaderElectionClient() should return the cached first client")
	}
}

func Test_ResolveViolationPublisher(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	publisher1 := resolver.ViolationPublisher()
	publisher2 := resolver.ViolationPublisher()

	if publisher1 != publisher2 {
		t.Error("A second call resolver.ViolationPublisher() should return the cached first publisher")
	}
}

func Test_ResolvePolicyReportClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	client1, err := resolver.PolicyReportClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, err := resolver.PolicyReportClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if client1 != client2 {
		t.Error("A second call resolver.PolicyReportClient() should return the cached first client")
	}
}

func Test_ResolveEventClient(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	client1, err := resolver.EventClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	client2, err := resolver.EventClient()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}

	if client1 != client2 {
		t.Error("A second call resolver.EventClient() should return the cached first client")
	}
}

func Test_ResolvePolicyStore(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	store1 := resolver.PolicyStore()
	store2 := resolver.PolicyStore()
	if store1 != store2 {
		t.Error("A second call resolver.PolicyStore() should return the cached first store")
	}
}

func Test_ResolvePolicyMapper(t *testing.T) {
	resolver := config.NewResolver(&config.Config{}, &rest.Config{})

	mapper1 := resolver.Mapper()
	mapper2 := resolver.Mapper()
	if mapper1 != mapper2 {
		t.Error("A second call resolver.Mapper() should return the cached first mapper")
	}
}

func Test_ResolveAPIServer(t *testing.T) {
	resolver := config.NewResolver(testConfig, &rest.Config{})

	server := resolver.APIServer(context.Background(), func() bool { return true })
	if server == nil {
		t.Error("Error: Should return API Server")
	}
}

func Test_ResolveClientsetWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(&config.Config{}, k8sConfig)

	_, err := resolver.Clientset()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveLeaderElectionWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(&config.Config{}, k8sConfig)

	_, err := resolver.LeaderElectionClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(&config.Config{}, k8sConfig)

	_, err := resolver.PolicyClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolveEventClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(&config.Config{}, k8sConfig)

	_, err := resolver.EventClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_ResolvePolicyReportClientWithInvalidK8sConfig(t *testing.T) {
	k8sConfig := &rest.Config{}
	k8sConfig.Host = "invalid/url"

	resolver := config.NewResolver(&config.Config{}, k8sConfig)

	_, err := resolver.PolicyReportClient()
	if err == nil {
		t.Error("Error: 'host must be a URL or a host:port pair' was expected")
	}
}

func Test_RegisterStoreListener(t *testing.T) {
	t.Run("Register StoreListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterStoreListener()

		if len(resolver.EventPublisher().GetListener()) != 1 {
			t.Error("Expected one Listener to be registered")
		}
	})
}

func Test_RegisterMetricsListener(t *testing.T) {
	t.Run("Register MetricsListener", func(t *testing.T) {
		resolver := config.NewResolver(testConfig, &rest.Config{})
		resolver.RegisterMetricsListener()

		if len(resolver.EventPublisher().GetListener()) != 1 {
			t.Error("Expected one Listener to be registered")
		}
	})
}
