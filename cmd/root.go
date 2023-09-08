package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
)

// NewCLI creates a new instance of the root CLI
func NewCLI() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kyverno-plugin",
		Short: "Generates Kyverno Policy Metrics",
		Long: `Generates Prometheus Metrics from Kyveno Policies.
		Creates REST APIs for Policies to use with other tools like Policy Reporter UI`,
	}

	rootCmd.AddCommand(newRunCMD())

	return rootCmd
}

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	v := viper.New()

	v.SetDefault("api.port", 8080)
	v.SetDefault("blockReports.source", "Kyverno Event")
	v.SetDefault("blockReports.results.maxPerReport", 100)

	v.SetDefault("leaderElection.releaseOnCancel", true)
	v.SetDefault("leaderElection.leaseDuration", 15)
	v.SetDefault("leaderElection.renewDeadline", 10)
	v.SetDefault("leaderElection.retryPeriod", 2)

	cfgFile := ""

	configFlag := cmd.Flags().Lookup("config")
	if configFlag != nil {
		cfgFile = configFlag.Value.String()
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("config")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		zap.L().Info("no seperate configuration file found")
	}

	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	if flag := cmd.Flags().Lookup("port"); flag != nil {
		v.BindPFlag("api.port", flag)
	}

	if flag := cmd.Flags().Lookup("rest-enabled"); flag != nil {
		v.BindPFlag("rest.enabled", flag)
	}

	if flag := cmd.Flags().Lookup("metrics-enabled"); flag != nil {
		v.BindPFlag("metrics.enabled", flag)
	}

	if flag := cmd.Flags().Lookup("lease-name"); flag != nil {
		v.BindPFlag("leaderElection.lockName", flag)
	}

	if err := v.BindEnv("leaderElection.podName", "POD_NAME"); err != nil {
		zap.L().Warn("failed to bind env POD_NAME")
	}

	if err := v.BindEnv("namespace", "POD_NAMESPACE"); err != nil {
		zap.L().Warn("failed to bind env POD_NAMESPACE")
	}

	c := &config.Config{}

	err := v.Unmarshal(c)

	return c, err
}
