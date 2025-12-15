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

func SendResetPasswordEmail(to, token string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	appURL := os.Getenv("APP_URL")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	if smtpEmail == "" || smtpPassword == "" || appURL == "" {
		return fmt.Errorf("SMTP_EMAIL, SMTP_PASSWORD, or APP_URL not set")
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	// NOTE: Frontend handling for reset password typically involves a page where user inputs new password.
	// So this link should point to the frontend reset password page.
	// Assuming frontend url is needed here, checking if handler uses FRONTEND_URL or APP_URL via env.
	// auth_handler.go uses FRONTEND_URL. Let's reuse that pattern if possible, but here we only have APP_URL in env loading (based on SendVerificationEmail specific code).
	// But wait, SendVerificationEmail uses APP_URL which points to /api/v1/auth/verify?token=... which is a backend endpoint that redirects.
	// For Reset Password, it usually points directly to Frontend.

	// Let's try to get FRONTEND_URL from env, defaulting to localhost:5173 like in handler.
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://coby-learn-ai.vercel.app"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	subject := "Subject: CobyLearnAI - Reset Password\r\n"
	mime := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	body := fmt.Sprintf(`
<html>
<body>
	<h3>Reset Password Request</h3>
	<p>You requested to reset your password. Please click the link below to set a new password:</p>
	<a href="%s">Reset Password</a>
	<p>Or copy this link:</p>
	<p>%s</p>
	<p>This link will expire in 1 hour.</p>
	<p>If you did not request this, please ignore this email.</p>
	<p>Thank you,<br/>CobyLearnAI Team</p>
</body>
</html>
`, resetLink, resetLink)

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
