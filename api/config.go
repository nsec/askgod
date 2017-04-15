package api

// URL: /1.0/config
// Access: admin

// Config represents the Askgod configuration
type Config struct {
	Daemon   ConfigDaemon   `yaml:"daemon" json:"daemon"`
	Database ConfigDatabase `yaml:"database" json:"database"`
	Scoring  ConfigScoring  `yaml:"scoring" json:"scoring"`
	Teams    ConfigTeams    `yaml:"teams" json:"teams"`
	Subnets  ConfigSubnets  `yaml:"subnets" json:"subnets"`
}

// ConfigDaemon represents the Daemon part of the Askgod configuration
type ConfigDaemon struct {
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
	ClusterPeers     []string `yaml:"cluster_peers" json:"cluster_peers"`
	HAProxyHeader    bool     `yaml:"haproxy_header" json:"haproxy_header"`
	HTTPPort         int      `yaml:"http_port" json:"http_port"`
	HTTPSPort        int      `yaml:"https_port" json:"https_port"`
	HTTPSCertificate string   `yaml:"https_certificate" json:"https_certificate"`
	HTTPSKey         string   `yaml:"https_key" json:"https_key"`
	LogLevel         string   `yaml:"log_level" json:"log_level"`
	LogFile          string   `yaml:"log_file" json:"log_file"`
}

// ConfigDatabase represents the Daemon part of the Askgod configuration
type ConfigDatabase struct {
	Driver      string `yaml:"driver" json:"driver"`
	Host        string `yaml:"host" json:"host"`
	Username    string `yaml:"username" json:"username"`
	Password    string `yaml:"password" json:"password"`
	Name        string `yaml:"name" json:"name"`
	Connections int    `yaml:"connections" json:"connections"`
}

// ConfigScoring represents the Daemon part of the Askgod configuration
type ConfigScoring struct {
	EventName  string `yaml:"event_name" json:"event_name"`
	HideOthers bool   `yaml:"hide_others" json:"hide_others"`
	ReadOnly   bool   `yaml:"read_only" json:"read_only"`
}

// ConfigTeams represents the Daemon part of the Askgod configuration
type ConfigTeams struct {
	SelfRegister bool `yaml:"self_register" json:"self_register"`
	SelfUpdate   bool `yaml:"self_update" json:"self_update"`
}

// ConfigSubnets represents the Daemon part of the Askgod configuration
type ConfigSubnets struct {
	Admins []string `yaml:"admins" json:"admins"`
	Teams  []string `yaml:"teams" json:"teams"`
	Guests []string `yaml:"guests" json:"guests"`
}
