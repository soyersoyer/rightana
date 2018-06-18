package mail

import (
	"fmt"

	"github.com/go-mail/mail"
)

// SMTPConfig holds the configuration
type SMTPConfig struct {
	Hostname string
	Port     int
	User     string
	Password string
	Sender   string
	AppURL   string
	AppName  string
}

var (
	config SMTPConfig
)

// Configure sets the config variables
func Configure(smtpConfig SMTPConfig) {
	config = smtpConfig
}

// SendResetPassword sends an password change email to the user
func SendResetPassword(userName, recipient, displayName, resetKey string, expireMinutes int) error {
	body, err := getResetPasswordBody(userName, displayName, resetKey, expireMinutes)
	if err != nil {
		return err
	}
	return SendUserEmail(recipient, userName, "Reset your password", body)
}

// SendVerifyEmail sends an email verification request to the user
func SendVerifyEmail(userName, recipient, displayName, verificationKey string) error {
	body, err := getVerifyEmailBody(userName, displayName, verificationKey)
	if err != nil {
		return err
	}
	return SendUserEmail(recipient, userName, "Verify your email address", body)
}

// SendUserEmail sends a html email to an user
func SendUserEmail(recipient, name, subject, htmlBody string) error {
	m := mail.NewMessage()
	m.SetHeader("From", config.Sender)
	m.SetHeader("To", fmt.Sprintf("%s <%s>", name, recipient))
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)
	d := mail.NewDialer(
		config.Hostname,
		config.Port,
		config.User,
		config.Password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	return d.DialAndSend(m)
}
