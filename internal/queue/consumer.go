package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EwvwGeN/mailService/internal/config"
	"github.com/EwvwGeN/mailService/internal/structs"
	amqp "github.com/rabbitmq/amqp091-go"
)

type consumer struct {
	logger *slog.Logger
	cfg config.RabbitMQConfig
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(ctx context.Context, lg *slog.Logger, cfg config.RabbitMQConfig) (*consumer, error) {
	c := &consumer{
		logger: lg.With(slog.String("op", "queue")),
		cfg: cfg,
		conn:    nil,
		channel: nil,
	}

	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(c.cfg.ConnectionName)
	amqpURI := amqp.URI{
		Scheme: c.cfg.Scheme,
		Host: c.cfg.Host,
		Port: c.cfg.Port,
		Username: c.cfg.Username,
		Password: c.cfg.Password,
		Vhost: c.cfg.VirtualHost,
	}
	c.logger.Info("rabbitMQ dialing", slog.String("URI", amqpURI.String()))
	c.conn, err = amqp.DialConfig(amqpURI.String(), config)
	if err != nil {
		c.logger.Error(ErrStartupConnection.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrStartupConnection.Error(), err)
	}

	c.logger.Info("got Connection")
	c.logger.Info("getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		c.logger.Error(ErrOpenChannel.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrOpenChannel.Error(), err)
	}

	c.logger.Info("got Channel")
	c.logger.Info("declaring Exchange", slog.String("exchange", c.cfg.ExchangerConfig.Name))
	if err = c.channel.ExchangeDeclare(
		c.cfg.ExchangerConfig.Name,
		c.cfg.ExchangerConfig.Type,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		c.logger.Error(ErrExchangeDeclare.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrExchangeDeclare.Error(), err)
	}

	c.logger.Info("declared Exchange", slog.String("exchange", c.cfg.ExchangerConfig.Name))
	c.logger.Info("declaring Queue", slog.String("queue", c.cfg.QueueConfig.Name))
	queue, err := c.channel.QueueDeclare(
		c.cfg.QueueConfig.Name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Error(ErrDeclareQueue.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrDeclareQueue.Error(), err)
	}
	c.cfg.QueueConfig.Name = queue.Name
	c.logger.Info(
		"declared Queue",
		slog.String("queue", c.cfg.QueueConfig.Name),
		slog.Int("messages", queue.Messages),
		slog.Int("consumers", queue.Consumers),
	)

	c.logger.Info("binding to Exchange", slog.String("key", c.cfg.BindingConfig.Key))
	if err = c.channel.QueueBind(
		queue.Name,
		c.cfg.BindingConfig.Key,
		c.cfg.ExchangerConfig.Name,
		false,
		nil,
	); err != nil {
		c.logger.Error(ErrExchangeBind.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrExchangeBind.Error(), err)
	}

	c.logger.Info("queue bound to Exchang", slog.String("key", c.cfg.BindingConfig.Key))
	return c, nil
}

func (c *consumer) Start() (chan *structs.Message, error) {
	c.logger.Info("starting Consume", slog.String("tag", c.cfg.ConsumerConfig.Tag))
	deliveries, err := c.channel.Consume(
		c.cfg.QueueConfig.Name,
		c.cfg.ConsumerConfig.Tag,
		c.cfg.ConsumerConfig.AutoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Error(ErrStartConsume.Error(), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", ErrStartConsume.Error(), err)
	}
	outChan := make(chan *structs.Message)
	go func() {
		for d := range deliveries {
			c.logger.Info(
				"got delivery",
				slog.Int("byte", len(d.Body)),
			)
			emailTo, ok := d.Headers["To"]
			if !ok {
				c.logger.Error("incorrect headers of message")
				d.Ack(false)
				continue
			}
			subject, ok := d.Headers["Subject"]
			if !ok {
				c.logger.Error("incorrect headers of message")
				d.Ack(false)
				continue
			}
			msg := &structs.Message{
				Subject: subject.(string),
				EmailTo: emailTo.(string),
				Body: d.Body,
			}
			outChan <- msg
			if !c.cfg.ConsumerConfig.AutoAck {
				msg.AckFunc = func() {
					d.Ack(false)
				}
			}
		}
		c.logger.Info("cancel consuming")
	}()
	return outChan, nil
}

func (c *consumer) Shutdown() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer c.logger.Info("AMQP shutdown OK")

	return nil
}