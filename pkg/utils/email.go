package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(to, token string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	if smtpEmail == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP credentials not set")
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	verifyLink := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify?token=%s", token)

	subject := "Subject: CobyLearnAI - Email Verification\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h3>Welcome to CobyLearnAI!</h3>
			<p>Please click the link below to verify your email address:</p>
			<a href="%s">Verify Email</a>
			<p>Or copy this link: %s</p>
			<p>Thank you,</p>
			<p>CobyLearnAI Team</p>
		</body>
		</html>
	`, verifyLink, verifyLink)

	msg := []byte(subject + mime + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}
