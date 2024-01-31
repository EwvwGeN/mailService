package queue

import "errors"

var (
	ErrStartupConnection = errors.New("failed to connect to RabbitMQ")
	ErrOpenChannel = errors.New("failed to open a channel")
	ErrDeclareQueue = errors.New("failed to declare a queue")
	ErrExchangeDeclare = errors.New("failed to declare an exchange")
	ErrExchangeBind = errors.New("failed to bind to exchange")
	ErrStartConsume = errors.New("failed to start of consumption")
)