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

	subject := "Subject: Verifikasi Email Focus Room\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h3>Selamat Datang di Focus Room!</h3>
			<p>Silakan klik link di bawah ini untuk memverifikasi email Anda:</p>
			<a href="%s">Verifikasi Email</a>
			<p>Atau copy link ini: %s</p>
			<p>Terima kasih,</p>
			<p>Tim Focus Room</p>
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
