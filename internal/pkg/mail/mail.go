package mail

import (
	"gopkg.in/gomail.v2"
	"proman-backend/config"
	"proman-backend/internal/pkg/log"
)

func SendMail(cc, receiver []string, subject string, body string) error {
	if !config.Mail.Enable {
		log.Info("Mail API is disabled")
		return nil
	}

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

	log.Infof("Mail sent to %v", receiver)
	return nil
}
