package config

// BasicAuth configuration
type BasicAuth struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	SecretRef string `mapstructure:"secretRef"`
}

// API configuration
type API struct {
	Port      int       `mapstructure:"port"`
	Logging   bool      `mapstructure:"logging"`
	BasicAuth BasicAuth `mapstructure:"basicAuth"`
}

type Logging struct {
	LogLevel    int8   `mapstructure:"logLevel"`
	Encoding    string `mapstructure:"encoding"`
	Development bool   `mapstructure:"development"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled"`
}

// Metrics configuration
type Metrics struct {
	Enabled bool `mapstructure:"enabled"`
}

// Results configuration
type Results struct {
	MaxPerReport   int  `mapstructure:"maxPerReport"`
	KeepOnlyLatest bool `mapstructure:"keepOnlyLatest"`
}

// LeaderElection configuration
type LeaderElection struct {
	LockName        string `mapstructure:"lockName"`
	PodName         string `mapstructure:"podName"`
	LeaseDuration   int    `mapstructure:"leaseDuration"`
	RenewDeadline   int    `mapstructure:"renewDeadline"`
	RetryPeriod     int    `mapstructure:"retryPeriod"`
	ReleaseOnCancel bool   `mapstructure:"releaseOnCancel"`
	Enabled         bool   `mapstructure:"enabled"`
}

// BlockReports configuration
type BlockReports struct {
	Enabled        bool    `mapstructure:"enabled"`
	Results        Results `mapstructure:"results"`
	Source         string  `mapstructure:"source"`
	EventNamespace string  `mapstructure:"eventNamespace"`
}

// Config of the Policyer
type Config struct {
	API            API            `mapstructure:"api"`
	REST           REST           `mapstructure:"rest"`
	Metrics        Metrics        `mapstructure:"metrics"`
	Kubeconfig     string         `mapstructure:"kubeconfig"`
	BlockReports   BlockReports   `mapstructure:"blockReports"`
	LeaderElection LeaderElection `mapstructure:"leaderElection"`
	Logging        Logging        `mapstructure:"logging"`
	Namespace      string         `mapstructure:"namespace"`
}
