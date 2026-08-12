package email

import (
	"fmt"
	"os"
	"strings"

	"github.com/resend/resend-go/v3"
)

type EmailSender interface {
	SendEmail(to []string, subject, htmlContent string) (string, error)
}

type ResendSender struct {
	client *resend.Client
	from   string
}

func NewResendSender() (*ResendSender, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Resend API not found")
	}

	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		return nil, fmt.Errorf("EMAIL_FROM not found")
	}

	client := resend.NewClient(apiKey)

	return &ResendSender{
		client: client,
		from:   fromEmail,
	}, nil
}

func (s *ResendSender) SendEmail(to []string, subject, htmlContent string) (string, error) {
	params := resend.SendEmailRequest{
		From:    s.from,
		To:      to,
		Subject: subject,
		Html:    htmlContent,
	}

	sent, err := s.client.Emails.Send(&params)

	if err != nil {
		return "", err
	}
	return sent.Id, nil
}

func (s *ResendSender) SendExpiryReminder(to []string, monitorName, target, expiryType string, daysRemaining int, expiresAt string) (string, error) {
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
