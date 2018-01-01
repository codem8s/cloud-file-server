package config

// Config describes program configuration
type Config struct {
	Listen      string
	LogRequests bool `yaml:"logRequests"`
	Connectors  []ConnectorConfig
}

//ConnectorConfig describes connector configuration
type ConnectorConfig struct {
	Type       string
	URI        string
	Region     string
	PathPrefix string `yaml:"pathPrefix"`
}
