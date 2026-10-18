package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/apierror"
	"netra-uptime/pkg/email"
	"netra-uptime/pkg/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// WebhookPayload is the JSON body sent to custom webhook endpoints.
type WebhookPayload struct {
	Event     string    `json:"event"`   // "monitor.down" | "monitor.up"
	Monitor   string    `json:"monitor"` // monitor name
	Target    string    `json:"target"`  // monitor target URL
	Status    string    `json:"status"`  // "down" | "up"
	Message   string    `json:"message"` // human-readable description
	Timestamp time.Time `json:"timestamp"`
}

// Dispatcher sends notifications to all active NotificationChannels for an org.
// It is safe to call concurrently from multiple goroutines.
type Dispatcher struct {
	db          *gorm.DB
	emailSender email.EmailSender
	httpClient  *http.Client
	sem         chan struct{}
}

// NewDispatcher creates a Dispatcher. The shared http.Client uses a 20s timeout
// to prevent slow webhooks from blocking goroutines.
func NewDispatcher(db *gorm.DB, emailSender email.EmailSender) *Dispatcher {
	return &Dispatcher{
		db:          db,
		emailSender: emailSender,
		httpClient:  utils.NewSafeHTTPClient(20 * time.Second),
		sem:         make(chan struct{}, 50), // max 50 concurrent outgoing notification deliveries
	}
}

// Dispatch sends the notification to every active channel registered for the
// monitor's organization and logs the outcome.
func (d *Dispatcher) Dispatch(monitor *model.Monitor, incidentID *string, event, subject, htmlBody string) {
	var channels []model.NotificationChannel
	if err := d.db.
		Where("organization_id = ? AND is_active = true", monitor.OrganizationID).
		Find(&channels).Error; err != nil {
		log.Printf("[Dispatcher] Failed to fetch channels for org %s: %v\n", monitor.OrganizationID, err)
		return
	}

	if len(channels) == 0 {
		return
	}

	monitorID := monitor.ID
	payload := WebhookPayload{
		Event:     event,
		Monitor:   monitor.Name,
		Target:    monitor.Target,
		Status:    monitor.Status,
		Message:   subject,
		Timestamp: time.Now(),
	}

	for _, ch := range channels {
		ch := ch // capture loop variable for goroutine
		d.sem <- struct{}{}
		go func() {
			defer func() { <-d.sem }()
			config, err := unmarshalConfig(ch.Config)
			if err != nil {
				log.Printf("[Dispatcher] Invalid config for channel %s: %v\n", ch.ID, err)
				return
			}

			channelID := ch.ID
			logEntry := &model.NotificationLog{
				MonitorID: &monitorID,
				ChannelID: &channelID,
				Channel:   ch.Type,
				Message:   fmt.Sprintf("[%s] %s → %s", ch.Type, monitor.Name, subject),
				SentAt:    time.Now(),
			}

			// Parse incidentID if provided
			if incidentID != nil {
				// We store it for context but don't require it
				_ = incidentID
			}

			var dispatchErr error
			switch ch.Type {
			case "email":
				dispatchErr = d.dispatchEmail(config, subject, htmlBody)
				logEntry.Channel = "email_channel"
			case "slack":
				dispatchErr = d.dispatchSlack(config, payload)
			case "webhook":
				dispatchErr = d.dispatchWebhook(config, payload)
			case "discord":
				dispatchErr = d.dispatchDiscord(config, payload)
			default:
				log.Printf("[Dispatcher] Unknown channel type: %s\n", ch.Type)
				return
			}

			if dispatchErr != nil {
				log.Printf("[Dispatcher] Failed to dispatch to channel %s (%s): %v\n", ch.Name, ch.Type, dispatchErr)
				logEntry.Status = "failed"
				logEntry.ErrorDetail = apierror.SanitizeDispatchError(dispatchErr)
			} else {
				log.Printf("[Dispatcher] Dispatched to channel %s (%s) for monitor %s\n", ch.Name, ch.Type, monitor.Name)
				logEntry.Status = "sent"
			}

			d.db.Create(logEntry)
		}()
	}
}

// ─── Channel-specific dispatchers ────────────────────────────────────────────

func (d *Dispatcher) dispatchEmail(config map[string]string, subject, htmlBody string) error {
	if d.emailSender == nil {
		return fmt.Errorf("email sender not configured")
	}
	to, ok := config["email"]
	if !ok || to == "" {
		return fmt.Errorf("missing email in config")
	}
	_, err := d.emailSender.SendEmail([]string{to}, subject, htmlBody)
	return err
}

func (d *Dispatcher) dispatchSlack(config map[string]string, payload WebhookPayload) error {
	webhookURL, ok := config["webhook_url"]
	if !ok || webhookURL == "" {
		return fmt.Errorf("missing webhook_url in Slack config")
	}

	// SSRF check (defence-in-depth, also validated at channel creation)
	if err := validateExternalURL(webhookURL); err != nil {
		return fmt.Errorf("SSRF protection: %w", err)
	}

	// Slack Incoming Webhook expects {"text": "..."} format.
	// Use a richer Block Kit message for better UX.
	emoji := "✅"
	color := "#10B981" // Green

	statusLower := strings.ToLower(payload.Status)
	if statusLower == "down" {
		emoji = "🚨"
		color = "#EF4444" // Red
	} else if statusLower == "degraded" || statusLower == "unresponsive" {
		color = "#F59E0B" // Amber
		if strings.Contains(strings.ToLower(payload.Message), "timeout") || strings.Contains(strings.ToLower(payload.Message), "unresponsive") {
			emoji = "⏳"
		} else {
			emoji = "⚠️"
		}
	}

	unixTime := payload.Timestamp.Unix()
	formattedTime := fmt.Sprintf("<!date^%d^{date_short_pretty} at {time}|%s>",
		unixTime, payload.Timestamp.Format("02 Jan 2006 15:04:05 UTC"))

	slackBody := map[string]interface{}{
		"text": fmt.Sprintf("%s *%s* — %s", emoji, payload.Monitor, payload.Message),
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"text": fmt.Sprintf(">Target: `%s`\n>Status: *%s*\n>Time: %s",
					payload.Target, strings.ToUpper(payload.Status), formattedTime),
			},
		},
	}
	return d.postJSON(webhookURL, slackBody)
}

func (d *Dispatcher) dispatchWebhook(config map[string]string, payload WebhookPayload) error {
	targetURL, ok := config["url"]
	if !ok || targetURL == "" {
		return fmt.Errorf("missing url in webhook config")
	}

	// SSRF check (defence-in-depth)
	if err := validateExternalURL(targetURL); err != nil {
		return fmt.Errorf("SSRF protection: %w", err)
	}

	return d.postJSON(targetURL, payload)
}

func (d *Dispatcher) dispatchDiscord(config map[string]string, payload WebhookPayload) error {
	webhookURL, ok := config["webhook_url"]
	if !ok || webhookURL == "" {
		return fmt.Errorf("missing webhook_url in Discord config")
	}

	if err := validateExternalURL(webhookURL); err != nil {
		return fmt.Errorf("SSRF protection: %w", err)
	}

	color := 3066993 // Green (0x2ECC71) for UP
	statusLower := strings.ToLower(payload.Status)
	if statusLower == "down" {
		color = 15158332 // Red (0xE74C3C) for DOWN
	} else if statusLower == "degraded" || statusLower == "unresponsive" {
		color = 16096779 // Amber (0xF59E0B) for UNRESPONSIVE / DEGRADED
	}

	statusDisplay := strings.ToUpper(payload.Status)
	if (statusLower == "degraded" || statusLower == "unresponsive") && strings.Contains(strings.ToLower(payload.Message), "unresponsive") {
		statusDisplay = "UNRESPONSIVE"
	}

	discordBody := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       payload.Message,
				"description": fmt.Sprintf("Monitor **%s** (`%s`) status is **%s**.", payload.Monitor, payload.Target, statusDisplay),
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Monitor", "value": payload.Monitor, "inline": true},
					{"name": "Target", "value": fmt.Sprintf("`%s`", payload.Target), "inline": true},
					{"name": "Status", "value": statusDisplay, "inline": true},
				},
				"footer":    map[string]string{"text": "Netra Uptime Alerts"},
				"timestamp": payload.Timestamp.Format(time.RFC3339),
			},
		},
	}
	return d.postJSON(webhookURL, discordBody)
}

// postJSON marshals v to JSON and POSTs it to url with a context timeout.
func (d *Dispatcher) postJSON(targetURL string, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NetraUptime-Notifier/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response from webhook: %d", resp.StatusCode)
	}
	return nil
}
