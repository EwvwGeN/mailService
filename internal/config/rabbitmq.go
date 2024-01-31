package config

type RabbitMQConfig struct {
	Scheme          string          `mapstructure:"scheme"`
	Host            string          `mapstructure:"host"`
	Port            int             `mapstructure:"port"`
	Username        string          `mapstructure:"username"`
	Password        string          `mapstructure:"password"`
	VirtualHost     string          `mapstructure:"virtual_host"`
	ConnectionName  string          `mapstructure:"connection_name"`
	ExchangerConfig ExchangerConfig `mapstructure:"exchanger"`
	BindingConfig   BindingConfig   `mapstructure:"binding"`
	QueueConfig     QueueConfig     `mapstructure:"queue"`
	ConsumerConfig  ConsumerConfig  `mapstructure:"consumer"`
}

type ExchangerConfig struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`
}

type BindingConfig struct {
	Key string `mapstructure:"key"`
}

type QueueConfig struct {
	Name string `mapstructure:"name"`
}

type ConsumerConfig struct {
	Tag     string `mapstructure:"tag"`
	AutoAck bool   `mapstructure:"auto_ack"`
}