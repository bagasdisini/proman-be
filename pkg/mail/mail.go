package mail

import (
	"gopkg.in/gomail.v2"
	"proman-backend/internal/config"
	"proman-backend/pkg/log"
)

func SendMail(cc, receiver []string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", config.Mail.SenderName)
	mailer.SetHeader("To", receiver...)
	mailer.SetHeader("Cc", cc...)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(config.Mail.Host, config.Mail.Port, config.Mail.AuthMail, config.Mail.AuthPass)

	err := dialer.DialAndSend(mailer)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
