package config

// API configuration
type API struct {
	Port int `mapstructure:"port"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled"`
}

// Metrics configuration
type Metrics struct {
	Enabled bool `mapstructure:"enabled"`
}

// Config of the Policyer
type Config struct {
	API        API     `mapstructure:"api"`
	REST       REST    `mapstructure:"rest"`
	Metrics    Metrics `mapstructure:"metrics"`
	Kubeconfig string  `mapstructure:"kubeconfig"`
}
