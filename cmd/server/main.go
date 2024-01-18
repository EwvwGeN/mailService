package main

import (
	"flag"
	"fmt"
	"log/slog"

	c "github.com/EwvwGeN/mailService/internal/config"
	l "github.com/EwvwGeN/mailService/internal/logger"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}

func main() {
	flag.Parse()
	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("cant load config from path %s", configPath ))
	}
	logger := l.SetupLogger(cfg.LogLevel)

	logger.Info("config loaded")
	logger.Debug("config data", slog.Any("cfg", cfg))

}