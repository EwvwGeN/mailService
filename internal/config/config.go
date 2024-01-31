package config

import (
	"fmt"
	p "path"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Port int `mapstructure:"port"`
	SMTP SMTPConfig `mapstructure:"smtp"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AutomaticEnv()
	if path != "" {
		dir := p.Dir(path)
		file := p.Base(path)
		fileParts := strings.Split(file, ".")
		if len(fileParts) != 2 {
			return nil, fmt.Errorf("incorrect config file: %s", file)
		}
		viper.AddConfigPath(dir)
		viper.SetConfigName(fileParts[0])
		viper.SetConfigType(fileParts[1])
		err := viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
        return nil, err
    }
	return &config, nil
}