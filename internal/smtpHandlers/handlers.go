package smtphandlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EwvwGeN/mailService/internal/config"
	"gopkg.in/gomail.v2"
)

type SMTPHandler struct {
	logger *slog.Logger
	config config.SMTPConfig
	dialer *gomail.Dialer
}

func NewSMTPHandler(ctx context.Context, cfg config.SMTPConfig, lg *slog.Logger) *SMTPHandler {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return &SMTPHandler{
		logger: lg,
		config: cfg,
		dialer: dialer,
	}
}

func (s *SMTPHandler) Start(ctx context.Context, closer chan struct{}) {
	msgChan := make(chan *gomail.Message)
	errChan := make(chan error)

	var closed *bool

	receiverGroupe := new(sync.WaitGroup)

	for i := 0; i < s.config.NodeCfg.NodeCount; i++ {
		receiverGroupe.Add(1)
		go s.MessageReceiver(ctx, receiverGroupe, msgChan, errChan, closed)
	}

	for {
		select {
			case <- closer: {
				*closed = true
				receiverGroupe.Wait()
				return
			}
			case err := <-errChan: {
				if err != nil {
					s.logger.Error("error occurred", slog.String("error", err.Error()))
				}
				if s.config.NodeCfg.AlwaysRestart {
					s.logger.Info("receiver restart attempt")
					receiverGroupe.Add(1)
					go s.MessageReceiver(ctx, receiverGroupe, msgChan, errChan, closed)
					continue
				}
				if s.config.NodeCfg.CancelOnError {
					s.logger.Info("close all receivers")
					*closed = true
				}
			}
		}
	}
}

func (s *SMTPHandler) MessageReceiver(ctx context.Context, rg *sync.WaitGroup, msgChan chan *gomail.Message, errChan chan error, closed *bool) {
	defer rg.Done()
	var (
		sendCloser gomail.SendCloser
		err error
	)
	open := false
	
	wg := new(sync.WaitGroup)

	for err == nil && !*closed {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				s.logger.Error(ErrReceiveMessage.Error(), slog.String("error", ErrInternal.Error()))
				err = ErrReceiveMessage
				break
			}
			if !open {
				if sendCloser, err = s.dialer.Dial(); err != nil {
					s.logger.Error(ErrOpenConnection.Error(), slog.String("error", err.Error()))
					break
				}
				open = true
			}
			wg.Add(1)
			s.SendMessage(ctx, wg, sendCloser, msg)
		case <-time.After(30 * time.Second):
			if open {
				if err = sendCloser.Close(); err != nil {
					s.logger.Error("cant close connection", slog.String("error", ErrInternal.Error()))
					break
				}
				open = false
			}
		}
	}
	wg.Wait()
	sendCloser.Close()
	errChan <- err
}

func (s *SMTPHandler) SendMessage(ctx context.Context, wg *sync.WaitGroup, sender gomail.Sender, msg *gomail.Message) {
	defer wg.Done()
	for i := 0; i < s.config.RetriesCount; i++ {
		err := gomail.Send(sender, msg)
		if err != nil {
			s.logger.Error(
				ErrSendMessage.Error(),
				slog.String("error", err.Error()),
				slog.String("retries_left", fmt.Sprintf("%d", s.config.RetriesCount - i - 1)),
			)
			continue
		}
		return
	}
	
}