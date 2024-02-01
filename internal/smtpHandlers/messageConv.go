package smtphandler

import (
	"github.com/EwvwGeN/mailService/internal/structs"
	"gopkg.in/gomail.v2"
)

func ConvertMessage(ourMsg *structs.Message) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("Subject", ourMsg.Subject)
	m.SetHeader("To", ourMsg.EmailTo)
	m.SetBody("text/html", string(ourMsg.Body))
	return m
}