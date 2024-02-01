package smtphandler

import "errors"

var (
	ErrInternal = errors.New("internal error")
	ErrReceiveMessage = errors.New("cant receive message from channel")
	ErrOpenConnection = errors.New("cant open smtp connection")
	ErrSendMessage = errors.New("cant send message")
)