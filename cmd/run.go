package cmd

import (
	"context"
	"flag"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/violation"
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
			logger, err := resolver.Logger()

			policyClient, err := resolver.PolicyClient()
			if err != nil {
				return err
			}

			server := resolver.APIServer(cmd.Context(), policyClient.HasSynced)

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

			if c.BlockReports.Enabled {
				logger.Info("block reports enabled", zap.Int("resultsPerReport", c.BlockReports.Results.MaxPerReport))
				eventClient, err := resolver.EventClient()
				if err != nil {
					return err
				}

				policyReportClient, err := resolver.PolicyReportClient()
				if err != nil {
					return err
				}

				resolver.ViolationPublisher().RegisterListener(func(pv violation.PolicyViolation) {
					policyReportClient.ProcessViolation(ctx, pv)
				})

				var stop chan struct{}
				defer close(stop)

				leClient, err := resolver.LeaderElectionClient()
				if err != nil {
					return err
				}

				if c.LeaderElection.Enabled {
					leClient.RegisterOnStart(func(c context.Context) {
						logger.Info("started leadership")

						stop = make(chan struct{})

						if err = eventClient.Run(stop); err != nil {
							logger.Error("failed to run EventClient", zap.Error(err))
						}
					}).RegisterOnNew(func(currentID, lockID string) {
						if currentID != lockID {
							logger.Info("leadership", zap.String("leader", currentID))
						}
					}).RegisterOnStop(func() {
						logger.Info("stopped leadership")
						close(stop)
					})

					go leClient.Run(cmd.Context())
				} else {
					stop = make(chan struct{})
					if err = eventClient.Run(stop); err != nil {
						return err
					}
				}
			}

			g := &errgroup.Group{}

			g.Go(func() error {
				stop := make(chan struct{})
				defer close(stop)
				logger.Info("start client", zap.Int("worker", 5))

				return policyClient.Run(5, stop)
			})

			logger.Info("server starting")
			g.Go(server.Start)

			return g.Wait()
		},
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().IntP("port", "p", 8080, "http port for the rest api")
	cmd.PersistentFlags().BoolP("metrics-enabled", "m", false, "Enable Metrics API")
	cmd.PersistentFlags().BoolP("rest-enabled", "r", false, "Enable REST API")
	cmd.PersistentFlags().String("lease-name", "policy-reporter-kyverno-plugin", "name of the LeaseLock")

	flag.Parse()

	return cmd
}
