package cmd

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/metrics"
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

			apiServer := resolver.APIServer()

			go func() {
				for {
					apiServer.Healthy()

					err := client.StartWatching()
					if err != nil {
						log.Printf("[ERROR] %s\n", err.Error())

						apiServer.Unhealthy()
					}

					time.Sleep(time.Second * 2)
				}
			}()

			go func() { errorChan <- apiServer.Start() }()

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
	cmd.PersistentFlags().IntP("apiPort", "a", 8080, "http port for the rest api")

	flag.Parse()

	return cmd
}
