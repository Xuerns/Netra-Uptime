package worker

import (
	"PulseWatch/internal/model"
	"PulseWatch/internal/notification"
	"PulseWatch/internal/worker/checker"
	"PulseWatch/pkg/email"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ProcessSSLDomainCheck performs SSL and Domain expiry checks for a monitor and sends email reminders if needed
func ProcessSSLDomainCheck(db *gorm.DB, emailSender email.EmailSender, dispatcher *notification.Dispatcher, monitor *model.Monitor) {
	if !monitor.EnableSSLDomainCheck || monitor.Type != "http" {
		return
	}

	now := time.Now()
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second

	// 1. SSL Expiry Check
	sslRes := checker.CheckSSLExpiry(monitor.Target, timeout)
	monitor.LastSSLCheckedAt = &now
	monitor.SSLStatus = sslRes.Status
	monitor.SSLExpiresAt = sslRes.ExpiresAt
	monitor.SSLDaysLeft = sslRes.DaysLeft
	monitor.SSLError = sslRes.ErrorMsg

	// 2. Domain Expiry Check
	domainRes := checker.CheckDomainExpiry(monitor.Target, timeout)
	monitor.LastDomainCheckedAt = &now
	monitor.DomainStatus = domainRes.Status
	monitor.DomainExpiresAt = domainRes.ExpiresAt
	monitor.DomainDaysLeft = domainRes.DaysLeft
	monitor.DomainError = domainRes.ErrorMsg

	// 3. Save updated fields to Database
	db.Model(monitor).Select(
		"SSLStatus", "SSLExpiresAt", "SSLDaysLeft", "SSLError", "LastSSLCheckedAt",
		"DomainStatus", "DomainExpiresAt", "DomainDaysLeft", "DomainError", "LastDomainCheckedAt",
	).Updates(monitor)

	// 4. Handle Notification Reminders for 30, 14, 7, 3 Days
	checkAndSendExpiryReminders(db, emailSender, dispatcher, monitor)
}

func checkAndSendExpiryReminders(db *gorm.DB, emailSender email.EmailSender, dispatcher *notification.Dispatcher, monitor *model.Monitor) {
	if emailSender == nil {
		return
	}

	reminderThresholds := []int{30, 14, 7, 3}

	// 1. SSL Reminder Check
	if monitor.SSLDaysLeft != nil {
		daysLeft := *monitor.SSLDaysLeft
		for _, threshold := range reminderThresholds {
			if daysLeft <= threshold {
				// Check if we already sent reminder for this threshold
				if monitor.LastSSLSentReminderDay == nil || *monitor.LastSSLSentReminderDay > threshold {
					sendExpiryNotification(db, emailSender, dispatcher, monitor, "SSL Certificate", daysLeft, monitor.SSLExpiresAt, threshold)
					monitor.LastSSLSentReminderDay = &threshold
					db.Model(monitor).Update("LastSSLSentReminderDay", threshold)
				}
				break
			}
		}
	}

	// 2. Domain Reminder Check
	if monitor.DomainDaysLeft != nil {
		daysLeft := *monitor.DomainDaysLeft
		for _, threshold := range reminderThresholds {
			if daysLeft <= threshold {
				// Check if we already sent reminder for this threshold
				if monitor.LastDomainSentReminderDay == nil || *monitor.LastDomainSentReminderDay > threshold {
					sendExpiryNotification(db, emailSender, dispatcher, monitor, "Domain Name", daysLeft, monitor.DomainExpiresAt, threshold)
					monitor.LastDomainSentReminderDay = &threshold
					db.Model(monitor).Update("LastDomainSentReminderDay", threshold)
				}
				break
			}
		}
	}
}

func sendExpiryNotification(db *gorm.DB, emailSender email.EmailSender, dispatcher *notification.Dispatcher, monitor *model.Monitor, expiryType string, daysLeft int, expiresAt *time.Time, threshold int) {
	recipients := getExpiryNotificationRecipients(db, monitor, threshold)

	expiresStr := "Unknown"
	if expiresAt != nil {
		expiresStr = expiresAt.Format("2006-01-02")
	}

	// Send to PICs/admins via Resend email
	if len(recipients) > 0 {
		if resendSender, ok := emailSender.(*email.ResendSender); ok {
			_, err := resendSender.SendExpiryReminder(recipients, monitor.Name, monitor.Target, expiryType, daysLeft, expiresStr)
			if err != nil {
				log.Printf("[SSL/Domain Expiry Error] Failed to send email to %v: %v", recipients, err)
			} else {
				log.Printf("[SSL/Domain Expiry] Email reminder sent for %s (%s) to %v", monitor.Name, expiryType, recipients)
			}
		}
	} else {
		log.Printf("[SSL/Domain Expiry] No email recipients found for monitor %s", monitor.Name)
	}

	// Also dispatch to org NotificationChannels (slack, webhook, email channels)
	if dispatcher != nil {
		subject := fmt.Sprintf("[PulseWatch] %s Expiry Warning: %s (%d days left)", expiryType, monitor.Name, daysLeft)
		body := fmt.Sprintf("<h3>%s Expiry Warning</h3><p>Monitor <b>%s</b> (%s): %s expires in <b>%d days</b> on %s.</p>",
			expiryType, monitor.Name, monitor.Target, expiryType, daysLeft, expiresStr)
		go dispatcher.Dispatch(monitor, nil, "expiry.warning", subject, body)
	}
}

func getExpiryNotificationRecipients(db *gorm.DB, monitor *model.Monitor, threshold int) []string {
	emailMap := make(map[string]bool)

	// Fetch PICs
	var pics []model.MonitorPIC
	db.Preload("User").Where("monitor_id = ?", monitor.ID).Find(&pics)

	for _, pic := range pics {
		if pic.User.Email != "" {
			emailMap[pic.User.Email] = true
		}
	}

	// For critical threshold (7 and 3 days), or if no PIC assigned, escalate to Org Admins & Owners
	if len(emailMap) == 0 || threshold <= 7 {
		var members []model.OrganizationMember
		db.Preload("User").Where("organization_id = ? AND role IN ('owner', 'admin')", monitor.OrganizationID).Find(&members)
		for _, member := range members {
			if member.User.Email != "" {
				emailMap[member.User.Email] = true
			}
		}
	}

	var emails []string
	for email := range emailMap {
		emails = append(emails, email)
	}

	return emails
}
