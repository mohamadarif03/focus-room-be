package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(to, token string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	appURL := os.Getenv("APP_URL")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	if smtpEmail == "" || smtpPassword == "" || appURL == "" {
		return fmt.Errorf("SMTP_EMAIL, SMTP_PASSWORD, or APP_URL not set")
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	verifyLink := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", appURL, token)

	subject := "Subject: CobyLearnAI - Email Verification\r\n"
	mime := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	body := fmt.Sprintf(`
<html>
<body>
	<h3>Welcome to CobyLearnAI!</h3>
	<p>Please click the link below to verify your email address:</p>
	<a href="%s">Verify Email</a>
	<p>Or copy this link:</p>
	<p>%s</p>
	<p>Thank you,<br/>CobyLearnAI Team</p>
</body>
</html>
`, verifyLink, verifyLink)

	msg := []byte(subject + mime + body)

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		smtpEmail,
		[]string{to},
		msg,
	)

	return err
}
