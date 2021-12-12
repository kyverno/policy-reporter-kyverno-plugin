package cmd

import (
	"context"
	"errors"
	"flag"

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

			client, err := resolver.PolicyClient()
			if err != nil {
				return err
			}

			server := resolver.APIServer(client.GetFoundResources())

			if c.REST.Enabled {
				resolver.RegisterStoreListener()
				server.RegisterREST()
			}

			if c.Metrics.Enabled {
				resolver.RegisterMetricsListener()
				server.RegisterMetrics()
			}

			g := &errgroup.Group{}

			g.Go(server.Start)

			g.Go(func() error {
				eventChan := client.StartWatching(ctx)

				resolver.EventPublisher().Publish(eventChan)

				return errors.New("event publisher stoped")
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
