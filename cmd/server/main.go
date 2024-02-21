package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	c "github.com/EwvwGeN/mailService/internal/config"
	l "github.com/EwvwGeN/mailService/internal/logger"
	"github.com/EwvwGeN/mailService/internal/queue"
	smtphandler "github.com/EwvwGeN/mailService/internal/smtpHandlers"
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

	smtpH := smtphandler.NewSMTPHandler(context.Background(), logger, cfg.SMTP)
	consumer, err := queue.NewConsumer(context.Background(), logger, cfg.RabbitMQ)
	if err != nil {
		logger.Error("failed to create the consumer", slog.String("error", err.Error()))
		panic(err)
	}
	msgChan, err := consumer.Start()
	if err != nil {
		logger.Error("failed to start the consumer", slog.String("error", err.Error()))
		panic(err)
	}
	closer := make(chan struct{})
	errChan := smtpH.Start(context.Background(), closer, msgChan)

	stopChecker := make(chan os.Signal, 1)
	signal.Notify(stopChecker, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <- stopChecker:
		break
	case err = <- errChan:
		logger.Error("error while work of the smtp handler", slog.String("error", err.Error()))
		break
	}
	logger.Info("stopping service")
	err = consumer.Shutdown()
	if err != nil {
		logger.Error("failed to close the consumer", slog.String("error", err.Error()))
	}
	logger.Info("consumer closed")
	closer <- struct{}{}
	logger.Info("service stopped")

}