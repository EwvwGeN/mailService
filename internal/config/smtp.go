package config

type SMTPConfig struct {
	Host         string     `mapstructure:"host"`
	Port         int        `mapstructure:"port"`
	Username     string     `mapstructure:"username"`
	Password     string     `mapstructure:"password"`
	NodeCfg      NodeConfig `mapstructure:"node"`
	RetriesCount int        `mapstructure:"retries_count"`
}

type NodeConfig struct {
	NodeCount     int  `mapstructure:"node_count"`
	AlwaysRestart bool `mapstructure:"alws_restart"`
	CancelOnError bool `mapstructure:"cancel_on_error"`
}