package utils

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

type ServerConfig struct {
	UseHTTPS     bool   `yaml:"useHTTPS"`
	Port         int    `yaml:"port"`
	CertFile     string `yaml:"certFile"`
	KeyFile      string `yaml:"keyFile"`
	ClientCAFile string `yaml:"clientCAFile"`
}

type PrometheusConfig struct {
	URL      string `yaml:"url"`
	Resource string `yaml:"resource"`
	Query    Query  `yaml:"query"`
}

type Query struct {
	Memory UsageRequest `yaml:"memory"`
	CPU    UsageRequest `yaml:"cpu"`
}

type UsageRequest struct {
	Usage   string `yaml:"usage"`
	Request string `yaml:"request"`
}
