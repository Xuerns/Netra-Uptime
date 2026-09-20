package email

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wneessen/go-mail"
)

// EmailSender defines the generic email-sending contract for PulseWatch
type EmailSender interface {
	SendEmail(to []string, subject, htmlContent string) (string, error)
	SendAlert(to []string, subject, htmlContent string) error
	SendExpiryReminder(to []string, monitorName, target, expiryType string, daysRemaining int, expiresAt string) (string, error)
}

// SMTPSender implements EmailSender using SMTP via go-mail
type SMTPSender struct {
	from     string
	host     string
	port     int
	password string
}

// NewSMTPSender creates a new SMTPSender configured from environment variables
func NewSMTPSender() (*SMTPSender, error) {
	from := os.Getenv("SMTP_EMAIL")
	if from == "" {
		from = os.Getenv("EMAIL_FROM")
	}
	if from == "" {
		return nil, errors.New("SMTP_EMAIL or EMAIL_FROM not found")
	}

	pass := os.Getenv("SMTP_PASSWORD")
	if pass == "" {
		return nil, errors.New("SMTP_PASSWORD not found")
	}

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &SMTPSender{
		from:     from,
		host:     host,
		port:     port,
		password: pass,
	}, nil
}

func (s *SMTPSender) newClient() (*mail.Client, error) {
	return mail.NewClient(
		s.host,
		mail.WithPort(s.port),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(s.from),
		mail.WithPassword(s.password),
	)
}

// SendEmail sends a generic HTML email to multiple recipients via SMTP
func (s *SMTPSender) SendEmail(to []string, subject, htmlContent string) (string, error) {
	if len(to) == 0 {
		return "", errors.New("no recipients provided")
	}

	message := mail.NewMsg()
	if err := message.From(s.from); err != nil {
		return "", fmt.Errorf("failed to set from address: %w", err)
	}

	if err := message.To(to...); err != nil {
		return "", fmt.Errorf("failed to set to address: %w", err)
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, htmlContent)

	client, err := s.newClient()
	if err != nil {
		return "", fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSend(message); err != nil {
		return "", fmt.Errorf("failed to send mail: %w", err)
	}

	return message.GetMessageID(), nil
}

// SendAlert sends an alert notification via SMTP
func (s *SMTPSender) SendAlert(to []string, subject, htmlContent string) error {
	_, err := s.SendEmail(to, subject, htmlContent)
	return err
}

// SendExpiryReminder sends SSL/Domain expiry notifications using styled HTML email
func (s *SMTPSender) SendExpiryReminder(to []string, monitorName, target, expiryType string, daysRemaining int, expiresAt string) (string, error) {
	badgeColor := "#f59e0b" // yellow/warning for 30 & 14 days
	if daysRemaining <= 7 {
		badgeColor = "#ef4444" // red/danger for 7 & 3 days
	}

	subject := fmt.Sprintf("Warning: %s for %s expires in %d days!", expiryType, monitorName, daysRemaining)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#0e0e0f;font-family:'Inter','Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0e0e0f;padding:40px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;background-color:#131314;border:1px solid #2a2a2b;border-radius:12px;padding:40px 32px;">

        <tr><td align="center" style="padding-bottom:24px;">
          <span style="font-size:22px;font-weight:700;color:#4edea3;letter-spacing:-0.02em;">PulseWatch</span>
        </td></tr>

        <tr><td align="center" style="padding-bottom:12px;">
          <span style="display:inline-block;background-color:%s20;border:1px solid %s40;color:%s;font-size:13px;font-weight:600;padding:6px 16px;border-radius:20px;">
            Expiration Warning (%d Days Remaining)
          </span>
        </td></tr>

        <tr><td align="center" style="padding-bottom:16px;">
          <h1 style="margin:0;font-size:20px;font-weight:600;color:#e5e2e3;letter-spacing:-0.01em;">%s Expiration Alert</h1>
        </td></tr>

        <tr><td align="left" style="padding-bottom:24px;">
          <p style="margin:0 0 12px;font-size:14px;line-height:1.6;color:#bbcabf;">
            Your monitored service <strong style="color:#ffffff;">%s</strong> (%s) has an upcoming %s expiration date.
          </p>
          <div style="background-color:#1a1a1b;border:1px solid #2a2a2b;border-radius:8px;padding:16px;margin-top:12px;">
            <p style="margin:0 0 6px;font-size:13px;color:#86948a;">Target: <strong style="color:#e5e2e3;">%s</strong></p>
            <p style="margin:0 0 6px;font-size:13px;color:#86948a;">Expiration Date: <strong style="color:%s;">%s</strong></p>
            <p style="margin:0;font-size:13px;color:#86948a;">Time Left: <strong style="color:%s;">%d Days</strong></p>
          </div>
        </td></tr>

        <tr><td align="center" style="padding-bottom:24px;">
          <p style="margin:0;font-size:13px;line-height:1.5;color:#86948a;">
            Please renew your %s soon to avoid service disruption or security warnings for your users.
          </p>
        </td></tr>

        <tr><td style="padding-bottom:20px;">
          <hr style="border:none;border-top:1px solid #2a2a2b;margin:0;">
        </td></tr>

        <tr><td align="center">
          <p style="margin:0;font-size:12px;color:#86948a;">© 2026 PulseWatch SaaS. Automated Monitoring Alert.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, badgeColor, badgeColor, badgeColor, daysRemaining, expiryType, monitorName, target, strings.ToLower(expiryType), target, badgeColor, expiresAt, badgeColor, daysRemaining, strings.ToLower(expiryType))

	return s.SendEmail(to, subject, htmlBody)
}

// Standalone functions for backward compatibility with auth & verification handlers:

// SendAlert sends alert using standalone configuration
func SendAlert(to []string, subject, htmlContent string) error {
	sender, err := NewSMTPSender()
	if err != nil {
		return err
	}
	return sender.SendAlert(to, subject, htmlContent)
}

func MailSender(to, link string) error {
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASSWORD")
	message := mail.NewMsg()

	if err := message.From(from); err != nil {
		return errors.New("Failed to set from Address")
	}

	if err := message.To(to); err != nil {
		return errors.New("Failed to set to Address")
	}

	message.Subject("Reset your Pulsewatch password")

	htmlBody := `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#0e0e0f;font-family:'Inter','Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#0e0e0f;padding:40px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#131314;border:1px solid #2a2a2b;border-radius:12px;padding:40px 32px;">

        <tr><td align="center" style="padding-bottom:32px;">
          <span style="font-size:22px;font-weight:700;color:#4edea3;letter-spacing:-0.02em;">Pulsewatch</span>
        </td></tr>

        <tr><td align="center" style="padding-bottom:8px;">
          <h1 style="margin:0;font-size:22px;font-weight:600;color:#e5e2e3;letter-spacing:-0.01em;">Reset your password</h1>
        </td></tr>

        <tr><td align="center" style="padding-bottom:32px;">
          <p style="margin:0;font-size:15px;line-height:1.6;color:#bbcabf;">
            We received a request to reset the password for your account. Click the button below to choose a new password.
          </p>
        </td></tr>

        <tr><td align="center" style="padding-bottom:24px;">
          <a href="` + link + `" target="_blank" style="display:inline-block;background-color:#10b981;color:#003824;font-size:15px;font-weight:600;text-decoration:none;padding:12px 32px;border-radius:8px;">
            Reset Password
          </a>
        </td></tr>

        <tr><td align="center" style="padding-bottom:32px;">
          <p style="margin:0;font-size:13px;line-height:1.5;color:#86948a;">
            This link will expire in <strong style="color:#bbcabf;">15 minutes</strong>. If you didn't request a password reset, you can safely ignore this email.
          </p>
        </td></tr>

        <tr><td style="padding-bottom:24px;">
          <hr style="border:none;border-top:1px solid #2a2a2b;margin:0;">
        </td></tr>

        <tr><td>
          <p style="margin:0 0 8px;font-size:12px;color:#86948a;">If the button doesn't work, copy and paste this link into your browser:</p>
          <p style="margin:0;font-size:12px;word-break:break-all;">
            <a href="` + link + `" style="color:#4edea3;text-decoration:underline;">` + link + `</a>
          </p>
        </td></tr>
      </table>

      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:480px;padding:24px 0;">
        <tr><td align="center">
          <p style="margin:0;font-size:12px;color:#86948a;">© 2026 Pulsewatch. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

	message.SetBodyString(mail.TypeTextHTML, htmlBody)

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, _ := strconv.Atoi(portStr)

	client, err := mail.NewClient(host, mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithUsername(from), mail.WithPassword(pass))

	if err != nil {
		return errors.New("Failed to create mail client")
	}

	if err := client.DialAndSend(message); err != nil {
		return errors.New("Failed to send mail")
	}

	return nil
}

func SendVerificationCodeEmail(to, code string) error {
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASSWORD")
	message := mail.NewMsg()

	if err := message.From(from); err != nil {
		return errors.New("Failed to set from Address")
	}

	if err := message.To(to); err != nil {
		return errors.New("Failed to set to Address")
	}

	message.Subject("Your Verification Code")

	htmlBody := `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#0e0e0f;font-family:'Inter','Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#0e0e0f;padding:40px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#131314;border:1px solid #2a2a2b;border-radius:12px;padding:40px 32px;">

        <tr><td align="center" style="padding-bottom:32px;">
          <span style="font-size:22px;font-weight:700;color:#4edea3;letter-spacing:-0.02em;">Pulsewatch</span>
        </td></tr>

        <tr><td align="center" style="padding-bottom:8px;">
          <h1 style="margin:0;font-size:22px;font-weight:600;color:#e5e2e3;letter-spacing:-0.01em;">Verification Code</h1>
        </td></tr>

        <tr><td align="center" style="padding-bottom:32px;">
          <p style="margin:0;font-size:15px;line-height:1.6;color:#bbcabf;">
            Here is your 6-digit verification code. Please enter it to verify your action.
          </p>
        </td></tr>

        <tr><td align="center" style="padding-bottom:24px;">
          <div style="display:inline-block;background-color:#1A1A1B;border:1px solid #2A2A2B;color:#4edea3;font-size:28px;letter-spacing:6px;font-weight:700;padding:16px 32px;border-radius:8px;">
            ` + code + `
          </div>
        </td></tr>

        <tr><td align="center" style="padding-bottom:32px;">
          <p style="margin:0;font-size:13px;line-height:1.5;color:#86948a;">
            This code will expire in <strong style="color:#bbcabf;">15 minutes</strong>. If you didn't request this code, you can safely ignore this email.
          </p>
        </td></tr>

        <tr><td style="padding-bottom:24px;">
          <hr style="border:none;border-top:1px solid #2a2a2b;margin:0;">
        </td></tr>

      </table>

      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:480px;padding:24px 0;">
        <tr><td align="center">
          <p style="margin:0;font-size:12px;color:#86948a;">© 2026 Pulsewatch. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

	message.SetBodyString(mail.TypeTextHTML, htmlBody)

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, _ := strconv.Atoi(portStr)

	client, err := mail.NewClient(host, mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithUsername(from), mail.WithPassword(pass))

	if err != nil {
		return errors.New("Failed to create mail client")
	}

	if err := client.DialAndSend(message); err != nil {
		return errors.New("Failed to send mail")
	}

	return nil
}
