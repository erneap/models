package svcs

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/erneap/models/config"
)

type SmtpServer struct {
	Host     string
	Port     string
	Password string
	From     string
}

func (s *SmtpServer) Address() string {
	return s.Host + ":" + s.Port
}

func (s *SmtpServer) Send(to []string, subject, body string) error {

	toLine := strings.Join(to, ",")
	message := []byte("To: " + toLine + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)

	err := smtp.SendMail(s.Address(), auth, s.From, to, message)
	return err
}

func SendMail(to []string, subject, body string) error {
	smtpServer := SmtpServer{
		Host:     config.Config("SMTP_SERVER"),
		Port:     config.Config("SMTP_PORT"),
		Password: config.Config("SMTP_PASS"),
		From:     config.Config("SMTP_FROM"),
	}

	err := smtpServer.Send(to, subject, body)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
