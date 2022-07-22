package cmd

import (
	"context"
	"flag"
	"log"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run Policyer Watcher & HTTP Metrics Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

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

			policyClient, err := resolver.PolicyClient()
			if err != nil {
				return err
			}

			server := resolver.APIServer(policyClient.HasSynced)

			if c.REST.Enabled || c.BlockReports.Enabled {
				resolver.RegisterStoreListener()
			}

			if c.REST.Enabled {
				server.RegisterREST()
			}

			if c.Metrics.Enabled {
				resolver.RegisterMetricsListener()
				server.RegisterMetrics()
			}

			g := &errgroup.Group{}

			g.Go(server.Start)

			if c.BlockReports.Enabled {
				log.Printf("[INFO] Block Reports enabled, max results per Report: %d\n", c.BlockReports.Results.MaxPerReport)
				eventClient, err := resolver.EventClient()
				if err != nil {
					return err
				}

				policyReportClient, err := resolver.PolicyReportClient()
				if err != nil {
					return err
				}

				g.Go(func() error {
					eventChan := eventClient.StartWatching(ctx)

					for violation := range eventChan {
						err := policyReportClient.ProcessViolation(ctx, violation)
						if err != nil {
							log.Printf("[ERROR] %s", err)
						}
					}

					return nil
				})
			}

			g.Go(func() error {

				stop := make(chan struct{})
				defer close(stop)

				return policyClient.Run(stop)
			})

			return g.Wait()
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("port", "p", 8080, "http port for the rest api")
	cmd.PersistentFlags().BoolP("metrics-enabled", "m", false, "Enable Metrics API")
	cmd.PersistentFlags().BoolP("rest-enabled", "r", false, "Enable REST API")

	flag.Parse()

	return cmd
}
