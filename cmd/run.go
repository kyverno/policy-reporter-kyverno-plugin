package cmd

import (
	"flag"
	"net/http"

	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/config"
	"github.com/fjogeleit/policy-reporter-kyverno-plugin/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run PolicyReporter Watcher & HTTP Metrics Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			var k8sConfig *rest.Config
			if c.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c, k8sConfig)

			client, err := resolver.PolicyClient()
			if err != nil {
				return err
			}

			client.RegisterCallback(metrics.CreatePolicyMetricsCallback())

			errorChan := make(chan error)

			if c.API.Enabled {
				go func() { errorChan <- resolver.APIServer().Start() }()
			}

			go func() { errorChan <- client.StartWatching() }()

			go func() {
				http.Handle("/metrics", promhttp.Handler())

				errorChan <- http.ListenAndServe(":2113", nil)
			}()

			return <-errorChan
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("apiPort", "a", 0, "http port for the optional rest api")

	flag.Parse()

	return cmd
}
