package config

// Server configuration
type API struct {
	Port int `mapstructure:"port"`
}

// Config of the PolicyReporter
type Config struct {
	API        API    `mapstructure:"api"`
	Kubeconfig string `mapstructure:"kubeconfig"`
}
