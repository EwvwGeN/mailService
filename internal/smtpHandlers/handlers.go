package smtphandler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EwvwGeN/mailService/internal/config"
	"github.com/EwvwGeN/mailService/internal/structs"
	"gopkg.in/gomail.v2"
)

type SMTPHandler struct {
	logger *slog.Logger
	config config.SMTPConfig
	dialer *gomail.Dialer
}

func NewSMTPHandler(ctx context.Context, lg *slog.Logger, cfg config.SMTPConfig) *SMTPHandler {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return &SMTPHandler{
		logger: lg.With(slog.String("op", "smtp")),
		config: cfg,
		dialer: dialer,
	}
}

func (s *SMTPHandler) Start(ctx context.Context, closer chan struct{}, msgChan chan *structs.Message) (chan error) {
	var (
		errChan chan error = make(chan error)
		outErrChan chan error = make(chan error)
		closed bool = false
		err error
	)

	receiverGroupe := new(sync.WaitGroup)

	for i := 0; i < s.config.NodeCfg.NodeCount; i++ {
		receiverGroupe.Add(1)
		go s.MessageReceiver(ctx, receiverGroupe, msgChan, errChan, &closed)
	}

	go func() {
		for {
			select {
				case <- closer: {
					closed = true
					receiverGroupe.Wait()
					outErrChan <- err
					return
				}
				case err = <-errChan: {
					if err != nil {
						s.logger.Error("error occurred", slog.String("error", err.Error()))
					}
					if s.config.NodeCfg.AlwaysRestart {
						s.logger.Info("receiver restart attempt")
						receiverGroupe.Add(1)
						go s.MessageReceiver(ctx, receiverGroupe, msgChan, errChan, &closed)
						continue
					}
					if s.config.NodeCfg.CancelOnError {
						s.logger.Info("closing all receivers")
						closed = true
					}
				}
			}
		}
	}()
	return outErrChan
	
}

func (s *SMTPHandler) MessageReceiver(ctx context.Context, rg *sync.WaitGroup, msgChan chan *structs.Message, errChan chan error, closed *bool) {
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

func (s *SMTPHandler) SendMessage(ctx context.Context, wg *sync.WaitGroup, sender gomail.Sender, msg *structs.Message) {
	defer wg.Done()
	for i := 0; i < s.config.RetriesCount; i++ {
		m := ConvertMessage(msg)
		m.SetHeader("From", s.config.EmailFrom)
		err := gomail.Send(sender, m)
		if err != nil {
			s.logger.Error(
				ErrSendMessage.Error(),
				slog.String("error", err.Error()),
				slog.String("retries_left", fmt.Sprintf("%d", s.config.RetriesCount - i - 1)),
			)
			continue
		}
		msg.AckFunc()
		return
	}
	
}