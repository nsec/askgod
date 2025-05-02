package api

// URL: /1.0/config
// Access: admin

// Config represents the Askgod configuration.
type Config struct {
	ConfigPut `yaml:",inline"`
	Daemon    ConfigDaemon   `json:"daemon"   yaml:"daemon"`
	Database  ConfigDatabase `json:"database" yaml:"database"`
}

// ConfigPut represents the editable Askgod configuration.
type ConfigPut struct {
	Scoring ConfigScoring `json:"scoring" yaml:"scoring"`
	Teams   ConfigTeams   `json:"teams"   yaml:"teams"`
	Subnets ConfigSubnets `json:"subnets" yaml:"subnets"`
}

// ConfigDaemon represents the Daemon part of the Askgod configuration.
type ConfigDaemon struct {
	AllowedOrigins   []string `json:"allowed_origins"   yaml:"allowed_origins"`
	ClusterPeers     []string `json:"cluster_peers"     yaml:"cluster_peers"`
	HAProxyHeader    bool     `json:"haproxy_header"    yaml:"haproxy_header"`
	HTTPPort         int      `json:"http_port"         yaml:"http_port"`
	HTTPSPort        int      `json:"https_port"        yaml:"https_port"`
	HTTPSCertificate string   `json:"https_certificate" yaml:"https_certificate"`
	HTTPSKey         string   `json:"https_key"         yaml:"https_key"`
	PrometheusPort   int      `json:"prometheus_port"   yaml:"prometheus_port"`
	LogLevel         string   `json:"log_level"         yaml:"log_level"`
	LogFile          string   `json:"log_file"          yaml:"log_file"`
}

// ConfigDatabase represents the Daemon part of the Askgod configuration.
type ConfigDatabase struct {
	Driver      string `json:"driver"      yaml:"driver"`
	Host        string `json:"host"        yaml:"host"`
	Username    string `json:"username"    yaml:"username"`
	Password    string `json:"password"    yaml:"password"`
	Name        string `json:"name"        yaml:"name"`
	Connections int    `json:"connections" yaml:"connections"`
	TLS         bool   `json:"tls"         yaml:"tls"`
}

// ConfigScoring represents the Daemon part of the Askgod configuration.
type ConfigScoring struct {
	EventName  string   `json:"event_name"  yaml:"event_name"`
	HideOthers bool     `json:"hide_others" yaml:"hide_others"`
	ReadOnly   bool     `json:"read_only"   yaml:"read_only"`
	PublicTags []string `json:"public_tags" yaml:"public_tags"`
}

// ConfigTeams represents the Daemon part of the Askgod configuration.
type ConfigTeams struct {
	SelfRegister bool     `json:"self_register" yaml:"self_register"`
	SelfUpdate   bool     `json:"self_update"   yaml:"self_update"`
	Hidden       []string `json:"hidden"        yaml:"hidden"`
}

// ConfigSubnets represents the Daemon part of the Askgod configuration.
type ConfigSubnets struct {
	Admins []string `json:"admins" yaml:"admins"`
	Teams  []string `json:"teams"  yaml:"teams"`
	Guests []string `json:"guests" yaml:"guests"`
}
