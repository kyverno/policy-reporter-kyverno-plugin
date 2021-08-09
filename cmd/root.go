package cmd

import (
	"log"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		log.Println("[INFO] No separete configuration file found")
	}

	if flag := cmd.Flags().Lookup("kubeconfig"); flag != nil {
		v.BindPFlag("kubeconfig", flag)
	}

	if flag := cmd.Flags().Lookup("apiPort"); flag != nil {
		v.BindPFlag("api.port", flag)
	}

	c := &config.Config{}

	err := v.Unmarshal(c)

	return c, err
}
