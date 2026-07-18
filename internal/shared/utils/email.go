package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
)

// SendPasswordResetEmail sends a branded reset link to the user.
func SendPasswordResetEmail(to, token string) {
	if config.AppConfig == nil {
		Log.Error("config not loaded, cannot send email")
		return
	}
	emailCfg := config.AppConfig.Email
	subject, body := PasswordResetEmail(emailCfg.FrontendURL, token)
	if err := sendEmail(to, subject, body, emailCfg); err != nil {
		Log.WithError(err).Error("Failed to send password reset email")
	}
}

// SendVerificationEmail sends a branded email verification link.
func SendVerificationEmail(to, token string) {
	if config.AppConfig == nil {
		Log.Error("config not loaded, cannot send email")
		return
	}
	emailCfg := config.AppConfig.Email
	subject, body := VerificationEmail(emailCfg.FrontendURL, token)
	if err := sendEmail(to, subject, body, emailCfg); err != nil {
		Log.WithError(err).Error("Failed to send verification email")
	}
}

// SendEmailDirect sends a pre-composed email using AppConfig. Intended for use
// by background job handlers where the full body is already built.
func SendEmailDirect(to, subject, bodyHTML string) {
	if config.AppConfig == nil {
		Log.Error("config not loaded, cannot send email")
		return
	}
	if err := sendEmail(to, subject, bodyHTML, config.AppConfig.Email); err != nil {
		Log.WithError(err).WithField("to", to).Error("failed to send email")
	}
}

// sendEmail now receives the email config as a parameter
func sendEmail(to, subject, bodyHTML string, emailCfg config.Email) error {
	host := emailCfg.Host
	port := emailCfg.Port
	auth := smtp.PlainAuth("", emailCfg.Username, emailCfg.Password, host)

	headers := map[string]string{
		"From":         emailCfg.From,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=utf-8",
	}

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + bodyHTML

	addr := fmt.Sprintf("%s:%d", host, port)
	if port == 587 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Quit()
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(emailCfg.From); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}
	return smtp.SendMail(addr, auth, emailCfg.From, []string{to}, []byte(msg))
}
