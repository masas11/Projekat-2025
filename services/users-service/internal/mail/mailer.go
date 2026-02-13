package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"users-service/config"

	"gopkg.in/mail.v2"
)

var (
	mailConfig *config.Config
	mailer     *mail.Dialer
)

// Init initializes the mailer with SMTP configuration
func Init(cfg *config.Config) {
	mailConfig = cfg

	// Initialize mailer if SMTP host is provided
	// MailHog doesn't require authentication, so username/password are optional
	if cfg.SMTPHost != "" {
		// Use empty strings for username/password if not provided (for MailHog)
		username := cfg.SMTPUsername
		password := cfg.SMTPPassword
		if username == "" {
			username = "mailhog"
		}
		if password == "" {
			password = "mailhog"
		}
		mailer = mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, username, password)
		// MailHog doesn't use TLS, so disable StartTLS for MailHog
		if cfg.SMTPHost == "mailhog" {
			mailer.StartTLSPolicy = mail.NoStartTLS
			// Explicitly disable TLS for MailHog
			mailer.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         "",
			}
		} else {
			mailer.StartTLSPolicy = mail.MandatoryStartTLS
		}
		log.Printf("[EMAIL] SMTP configured: %s:%d (from: %s)", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	} else {
		log.Printf("[EMAIL] SMTP not configured - using mock mode")
	}
}

// sendEmail is a helper function to send emails
func sendEmail(to, subject, body string) error {
	// If SMTP is not configured, use mock mode
	if mailConfig == nil || mailConfig.SMTPHost == "" {
		log.Printf("[MOCK EMAIL] To: %s, Subject: %s", to, subject)
		log.Printf("[MOCK EMAIL] NOTE: SMTP not configured. Email not actually sent. Configure SMTP in docker-compose.yml to send real emails.")
		return nil
	}

	// Use net/smtp directly for MailHog (doesn't support TLS)
	if mailConfig.SMTPHost == "mailhog" {
		log.Printf("[EMAIL] Using MailHog direct SMTP for %s", to)
		return sendEmailViaMailHog(to, subject, body)
	}

	// Use gopkg.in/mail.v2 for other SMTP servers (with TLS)
	m := mail.NewMessage()
	m.SetHeader("From", mailConfig.SMTPFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Send email
	if err := mailer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL] Sent successfully to %s: %s", to, subject)
	return nil
}

// sendEmailViaMailHog sends email using net/smtp directly (no TLS required)
func sendEmailViaMailHog(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", mailConfig.SMTPHost, mailConfig.SMTPPort)
	log.Printf("[EMAIL] Connecting to MailHog at %s", addr)
	
	// Create email message
	from := mailConfig.SMTPFrom
	if from == "" {
		from = "noreply@musicstreaming.com"
	}
	
	msg := []byte(fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Send email via SMTP (no auth, no TLS for MailHog)
	// nil auth means no authentication
	err := smtp.SendMail(addr, nil, from, []string{to}, msg)
	if err != nil {
		log.Printf("[EMAIL ERROR] MailHog send failed: %v", err)
		return fmt.Errorf("failed to send email via MailHog: %w", err)
	}

	log.Printf("[EMAIL] Sent successfully to %s via MailHog: %s", to, subject)
	return nil
}

// SendOTP sends OTP code to user's email
func SendOTP(email, otp string) {
	subject := "Your OTP Code for Login"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.otp-code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; padding: 20px; background-color: white; border: 2px dashed #4CAF50; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Music Streaming Platform</h1>
				</div>
				<div class="content">
					<h2>Your OTP Code</h2>
					<p>Hello,</p>
					<p>You have requested to log in to your account. Please use the following OTP code:</p>
					<div class="otp-code">%s</div>
					<p>This code will expire in 5 minutes.</p>
					<p>If you did not request this code, please ignore this email.</p>
				</div>
				<div class="footer">
					<p>This is an automated message, please do not reply.</p>
				</div>
			</div>
		</body>
		</html>
	`, otp)

	if err := sendEmail(email, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] Failed to send OTP to %s: %v", email, err)
	}
}

// SendMagicLink sends magic link to user's email
func SendMagicLink(email, link string) {
	subject := "Your Magic Link for Account Recovery"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #2196F3; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Music Streaming Platform</h1>
				</div>
				<div class="content">
					<h2>Account Recovery</h2>
					<p>Hello,</p>
					<p>You have requested to recover your account. Click the button below to log in:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Recover Account</a>
					</p>
					<p>Or copy and paste this link into your browser:</p>
					<p style="word-break: break-all; color: #2196F3;">%s</p>
					<p>This link will expire in 15 minutes.</p>
					<p>If you did not request this link, please ignore this email.</p>
				</div>
				<div class="footer">
					<p>This is an automated message, please do not reply.</p>
				</div>
			</div>
		</body>
		</html>
	`, link, link)

	if err := sendEmail(email, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] Failed to send magic link to %s: %v", email, err)
	}
}

// SendVerificationEmail sends email verification link to user's email
func SendVerificationEmail(email, link string) {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #FF9800; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #FF9800; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Music Streaming Platform</h1>
				</div>
				<div class="content">
					<h2>Verify Your Email</h2>
					<p>Hello,</p>
					<p>Thank you for registering! Please verify your email address by clicking the button below:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Verify Email</a>
					</p>
					<p>Or copy and paste this link into your browser:</p>
					<p style="word-break: break-all; color: #FF9800;">%s</p>
					<p>If you did not create an account, please ignore this email.</p>
				</div>
				<div class="footer">
					<p>This is an automated message, please do not reply.</p>
				</div>
			</div>
		</body>
		</html>
	`, link, link)

	if err := sendEmail(email, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] Failed to send verification email to %s: %v", email, err)
	} else {
		// In mock mode, also log the verification link so user can copy it
		if mailConfig == nil || mailConfig.SMTPHost == "" || mailConfig.SMTPUsername == "" || mailConfig.SMTPPassword == "" {
			log.Printf("[MOCK EMAIL] Verification link for %s: %s", email, link)
		}
	}
}

// SendPasswordResetEmail sends password reset link to user's email
func SendPasswordResetEmail(email, link string) {
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #F44336; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #F44336; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Music Streaming Platform</h1>
				</div>
				<div class="content">
					<h2>Password Reset Request</h2>
					<p>Hello,</p>
					<p>You have requested to reset your password. Click the button below to reset it:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Reset Password</a>
					</p>
					<p>Or copy and paste this link into your browser:</p>
					<p style="word-break: break-all; color: #F44336;">%s</p>
					<p>This link will expire in 1 hour.</p>
					<p>If you did not request a password reset, please ignore this email and your password will remain unchanged.</p>
				</div>
				<div class="footer">
					<p>This is an automated message, please do not reply.</p>
				</div>
			</div>
		</body>
		</html>
	`, link, link)

	if err := sendEmail(email, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] Failed to send password reset email to %s: %v", email, err)
	}
}
