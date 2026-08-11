package email

import (
	"errors"
	"os"
	"strconv"

	"github.com/wneessen/go-mail"
)

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
	if host == "" {
		host = "smtp.gmail.com"
	}
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
	if host == "" {
		host = "smtp.gmail.com"
	}
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
