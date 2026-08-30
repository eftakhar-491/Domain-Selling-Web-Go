package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Email    string
	Password string
}

// getConfig loads SMTP configuration from environment variables
func getConfig() *SMTPConfig {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return &SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Email:    os.Getenv("SMTP_EMAIL"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

// sendMail handles email transmission over both Implicit TLS (port 465) and STARTTLS (port 587/25)
func sendMail(cfg *SMTPConfig, to []string, message []byte) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Email, cfg.Password, cfg.Host)

	// Port 465 uses Implicit TLS / SSL connection
	if strings.TrimSpace(cfg.Port) == "465" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect via TLS on port %s: %w", cfg.Port, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate SMTP: %w", err)
		}

		if err = client.Mail(cfg.Email); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to open data writer: %w", err)
		}

		if _, err = w.Write(message); err != nil {
			return fmt.Errorf("failed to write email body: %w", err)
		}

		if err = w.Close(); err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		return client.Quit()
	}

	// Port 587, 25, or others: use standard STARTTLS / Plain net/smtp.SendMail
	return smtp.SendMail(addr, auth, cfg.Email, to, message)
}

// SendResetCode sends a password reset OTP code via email
func SendResetCode(toEmail, code string) error {
	cfg := getConfig()

	if cfg.Host == "" || cfg.Email == "" || cfg.Password == "" {
		log.Println("⚠️ SMTP not configured. Reset code for", toEmail, "is:", code)
		return nil // Don't fail — just log in dev mode
	}

	from := cfg.Email
	to := []string{toEmail}

	subject := "Password Reset Code"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f4f4f7; margin: 0; padding: 0; }
        .container { max-width: 480px; margin: 40px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #6366f1, #8b5cf6); padding: 32px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; font-weight: 600; }
        .body { padding: 32px; text-align: center; }
        .body p { color: #555; font-size: 15px; line-height: 1.6; }
        .code-box { background: #f0f0ff; border: 2px dashed #6366f1; border-radius: 10px; padding: 20px; margin: 24px 0; }
        .code { font-size: 36px; font-weight: 700; letter-spacing: 8px; color: #6366f1; font-family: 'Courier New', monospace; }
        .footer { padding: 20px 32px; text-align: center; background: #fafafa; }
        .footer p { color: #999; font-size: 12px; margin: 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 Password Reset</h1>
        </div>
        <div class="body">
            <p>We received a request to reset your password. Use the code below to proceed:</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p>This code is valid for <strong>5 minutes</strong>.</p>
            <p style="color: #999; font-size: 13px;">If you didn't request this, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>© 2026 Project Setup. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, code)

	// Build MIME message
	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		from, toEmail, subject, htmlBody,
	)

	if err := sendMail(cfg, to, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Println("📧 Reset code sent to:", toEmail)
	return nil
}

// SendNotificationEmail sends a generic notification email
func SendNotificationEmail(toEmail, subject, body string) error {
	cfg := getConfig()

	if cfg.Host == "" || cfg.Email == "" || cfg.Password == "" {
		log.Printf("⚠️ SMTP not configured. Notification to %s [%s]: %s\n", toEmail, subject, body)
		return nil
	}

	from := cfg.Email
	to := []string{toEmail}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f4f4f7; margin: 0; padding: 0; }
        .container { max-width: 520px; margin: 40px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #3b82f6, #6366f1); padding: 24px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 20px; font-weight: 600; }
        .body { padding: 24px; color: #444; font-size: 15px; line-height: 1.6; }
        .footer { padding: 16px 24px; text-align: center; background: #fafafa; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 Alert Notification</h1>
        </div>
        <div class="body">
            <p>%s</p>
        </div>
        <div class="footer">
            <p>© 2026 Project Setup. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, body)

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		from, toEmail, subject, htmlBody,
	)

	if err := sendMail(cfg, to, []byte(message)); err != nil {
		return fmt.Errorf("failed to send notification email: %w", err)
	}

	log.Printf("📧 Notification email sent to %s [%s]\n", toEmail, subject)
	return nil
}
