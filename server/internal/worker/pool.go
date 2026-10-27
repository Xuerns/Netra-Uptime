package worker

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"netra-uptime/internal/model"
	"netra-uptime/internal/notification"
	"netra-uptime/internal/websocket"
	"netra-uptime/internal/worker/checker"
	"netra-uptime/pkg/cache"
	"netra-uptime/pkg/email"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Job struct {
	Monitor model.Monitor
}

type WorkerPool struct {
	db               *gorm.DB
	rdb              *redis.Client
	jobQueue         chan Job
	numWorkers       int
	emailSender      email.EmailSender
	hub              *websocket.WsHandler
	dispatcher       *notification.Dispatcher
	notifySem        chan struct{}
	inFlightMonitors sync.Map
}

func NewWorkerPool(db *gorm.DB, rdb *redis.Client, numWorkers int, sender email.EmailSender, hub *websocket.WsHandler) *WorkerPool {
	return &WorkerPool{
		db:          db,
		rdb:         rdb,
		jobQueue:    make(chan Job, 1000),
		numWorkers:  numWorkers,
		emailSender: sender,
		hub:         hub,
		dispatcher:  notification.NewDispatcher(db, sender),
		notifySem:   make(chan struct{}, 50), // max 50 concurrent notification goroutines
	}
}

func (wp *WorkerPool) GetDB() *gorm.DB {
	return wp.db
}

func (wp *WorkerPool) safeGoNotify(fn func()) {
	go func() {
		wp.notifySem <- struct{}{}
		defer func() { <-wp.notifySem }()
		fn()
	}()
}

func (wp *WorkerPool) Start() {
	log.Printf("[WorkerPool] Started %d workers\n", wp.numWorkers)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.worker(i)
	}
	// Start daily SSL & Domain Expiry background scheduler
	go wp.StartSSLDomainScheduler()
}

func (wp *WorkerPool) StartSSLDomainScheduler() {
	log.Println("[SSL/Domain Scheduler] Started background daily checker")
	// Run initial check on startup after a brief delay
	time.AfterFunc(15*time.Second, func() {
		wp.runAllSSLDomainChecks()
	})

	// Ticker for daily check (every 24 hours)
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		wp.runAllSSLDomainChecks()
	}
}

func (wp *WorkerPool) runAllSSLDomainChecks() {
	var monitors []model.Monitor
	err := wp.db.Where("is_active = true AND type = 'http' AND enable_ssl_domain_check = true").Find(&monitors).Error
	if err != nil {
		log.Printf("[SSL/Domain Scheduler Error] Failed to fetch monitors: %v", err)
		return
	}

	log.Printf("[SSL/Domain Scheduler] Running daily check for %d monitors", len(monitors))
	for _, m := range monitors {
		mon := m
		go ProcessSSLDomainCheck(wp.db, wp.emailSender, wp.dispatcher, &mon)
		// Small jitter delay between domain queries to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}
}

func (wp *WorkerPool) QueueJob(job Job) {
	wp.jobQueue <- job
}

func (wp *WorkerPool) worker(_ int) {
	for job := range wp.jobQueue {
		wp.procces(&job.Monitor)
	}
}

func (wp *WorkerPool) ProcessImmediate(monitorID uuid.UUID) (*model.Check, error) {
	var mon model.Monitor
	if err := wp.db.First(&mon, "id = ?", monitorID).Error; err != nil {
		return nil, err
	}
	check := wp.procces(&mon)
	if check == nil {
		return nil, errors.New("failed to execute check")
	}
	return check, nil
}

func (wp *WorkerPool) procces(monitor *model.Monitor) *model.Check {
	monID := monitor.ID.String()
	if _, loaded := wp.inFlightMonitors.LoadOrStore(monID, struct{}{}); loaded {
		return nil
	}
	defer wp.inFlightMonitors.Delete(monID)

	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var statusCode int
	var responseTime time.Duration
	var checkErr error

	switch monitor.Type {
	case "http":
		statusCode, responseTime, checkErr = checker.CheckHTTP(monitor.Target, timeout, monitor.ExpectedStatusCode)
	case "tcp":
		responseTime, checkErr = checker.CheckTCP(monitor.Target, timeout)
	case "ping":
		responseTime, checkErr = checker.CheckPing(monitor.Target, timeout)
	}

	var dbMonitor model.Monitor
	if err := wp.db.Preload("AlertRules", "is_active = true").First(&dbMonitor, "id = ?", monitor.ID).Error; err != nil {
		log.Printf("Failed to fetch monitor for state evaluation: %v\n", err)
		return nil
	}

	var effectiveRules []model.AlertRule
	if dbMonitor.AlertTemplateID != nil {
		if err := wp.db.Where("template_id = ? AND is_active = true", *dbMonitor.AlertTemplateID).Find(&effectiveRules).Error; err != nil {
			log.Printf("[Worker] Failed to load template rules for monitor %s: %v", monitor.Name, err)
		}
	} else {
		// If monitor has no explicit template assigned, check if organization has a default template
		var defaultTemplate model.AlertRuleTemplate
		if err := wp.db.Where("organization_id = ? AND is_default = true", dbMonitor.OrganizationID).First(&defaultTemplate).Error; err == nil {
			_ = wp.db.Where("template_id = ? AND is_active = true", defaultTemplate.ID).Find(&effectiveRules).Error
		}
	}

	prevStatus := dbMonitor.Status
	prevFailureCount := dbMonitor.FailureCount

	eval := EvaluateRules(&dbMonitor, effectiveRules, statusCode, responseTime, checkErr)

	errMsg := ""
	if checkErr != nil {
		errMsg = checker.FormatErrorMessage(checkErr, timeout)
	} else if eval.IsBreached {
		errMsg = eval.Reason
	}

	statusCheck := "up"
	if eval.IsBreached {
		if eval.Severity == "critical" {
			statusCheck = "down"
		} else {
			statusCheck = "degraded"
		}
	}

	checkResult := model.Check{
		MonitorID:      monitor.ID,
		Status:         statusCheck,
		StatusCode:     statusCode,
		ResponseTimeMs: int(responseTime.Milliseconds()),
		ErrorMessage:   errMsg,
		CheckedAt:      time.Now(),
	}

	if err := wp.db.Create(&checkResult).Error; err != nil {
		log.Printf("Failed to create check for monitor %s: %v\n", monitor.Name, err)
	}

	if !eval.IsBreached {
		// Success / Healthy Scenario
		if dbMonitor.Status == "down" || dbMonitor.Status == "degraded" {
			prevStatus := dbMonitor.Status
			dbMonitor.Status = "up"
			dbMonitor.FailureCount = 0

			subject := fmt.Sprintf("[Netra Uptime] Resolved: %s is back UP", monitor.Name)
			message := fmt.Sprintf("<h3>Service Restored</h3><p>Your monitor <b>%s</b> (%s) has returned to healthy state (previously %s).</p>", html.EscapeString(dbMonitor.Name), html.EscapeString(dbMonitor.Target), html.EscapeString(prevStatus))

			wp.safeGoNotify(func() {
				wp.sendNotification(&dbMonitor, subject, message)
			})
			wp.safeGoNotify(func() {
				wp.dispatcher.Dispatch(&dbMonitor, nil, "monitor.up", subject, message)
			})

			now := time.Now()
			err := wp.db.Model(&model.Incident{}).Where("monitor_id = ? AND status = ?", monitor.ID, "ongoing").Updates(map[string]interface{}{
				"status":      "resolved",
				"resolved_at": now,
				"summary":     "Resolved automatically: all alert rules passed healthy state.",
			}).Error

			if err != nil {
				log.Printf("[Incident] Failed to resolve incident for monitor %s: %v\n", monitor.Name, err)
			} else {
				log.Printf("[Incident] Monitor Resolved: %s is back UP!\n", monitor.Name)
			}
		} else {
			dbMonitor.FailureCount = 0
		}
	} else {
		// Rule Breached Scenario
		switch eval.Severity {
		case "critical":
			dbMonitor.FailureCount++

			if dbMonitor.FailureCount >= eval.RequiredBreaches && dbMonitor.Status != "down" {
				dbMonitor.Status = "down"

				subject := fmt.Sprintf("[Netra Uptime] ALERT: %s is DOWN (%s)", monitor.Name, eval.RuleName)
				message := fmt.Sprintf("<h3>Service Down Alert</h3><p>Your monitor <b>%s</b> (%s) triggered alert rule <b>%s</b>.<br>Reason: <i>%s</i> (breached %d consecutive times)</p>", html.EscapeString(dbMonitor.Name), html.EscapeString(dbMonitor.Target), html.EscapeString(eval.RuleName), html.EscapeString(eval.Reason), eval.RequiredBreaches)

				go wp.sendNotification(&dbMonitor, subject, message)

				var alertRuleUUID *uuid.UUID
				if eval.RuleID != nil {
					if parsed, err := uuid.Parse(*eval.RuleID); err == nil {
						alertRuleUUID = &parsed
					}
				}

				newIncident := model.Incident{
					MonitorID:     monitor.ID,
					Status:        "ongoing",
					Severity:      ResolveIncidentSeverity(dbMonitor.Criticality, eval.Severity),
					StartedAt:     time.Now(),
					Summary:       fmt.Sprintf("%s: %s", eval.RuleName, eval.Reason),
					TriggeredRule: eval.RuleName,
					AlertRuleID:   alertRuleUUID,
				}

				if err := wp.db.Create(&newIncident).Error; err != nil {
					log.Printf("[Incident] Failed to create incident for monitor %s: %v\n", monitor.Name, err)
				} else {
					incidentIDStr := newIncident.ID.String()
					go wp.dispatcher.Dispatch(&dbMonitor, &incidentIDStr, "monitor.down", subject, message)
					log.Printf("[Incident] MONITOR DOWN TRIGGERED: %s is DOWN (%s). Error: %s\n", monitor.Name, eval.RuleName, eval.Reason)
				}
			} else {
				log.Printf("[Check] Monitor %s breached critical rule %s %d time(s) (waiting for %d)\n", monitor.Name, eval.RuleName, dbMonitor.FailureCount, eval.RequiredBreaches)
			}
		case "warning":
			// Degraded / Unresponsive Warning (e.g. timeout, high latency)
			dbMonitor.FailureCount++

			if dbMonitor.FailureCount >= eval.RequiredBreaches && dbMonitor.Status != "degraded" && dbMonitor.Status != "down" {
				dbMonitor.Status = "degraded"

				isTimeout := eval.RuleName == "Request Timeout" || strings.Contains(strings.ToLower(eval.RuleName), "timeout")

				var subject, message string
				if isTimeout {
					subject = fmt.Sprintf("[Netra Uptime] UNRESPONSIVE: %s is not responding in time (%s)", monitor.Name, eval.RuleName)
					message = fmt.Sprintf("<h3>Service Unresponsive</h3><p>Your monitor <b>%s</b> (%s) did not respond within the expected timeout limit.<br>Reason: <i>%s</i> (breached %d consecutive times)</p><p style=\"color: #6B7280; font-size: 13px;\">Note: The service may be experiencing high traffic, slow queries, or cold starts rather than a complete outage.</p>", html.EscapeString(dbMonitor.Name), html.EscapeString(dbMonitor.Target), html.EscapeString(eval.Reason), eval.RequiredBreaches)
				} else {
					subject = fmt.Sprintf("[Netra Uptime] WARNING: %s is DEGRADED (%s)", monitor.Name, eval.RuleName)
					message = fmt.Sprintf("<h3>Service Degraded Warning</h3><p>Your monitor <b>%s</b> (%s) is experiencing degraded performance.<br>Rule: <b>%s</b><br>Reason: <i>%s</i> (breached %d consecutive times)</p>", html.EscapeString(dbMonitor.Name), html.EscapeString(dbMonitor.Target), html.EscapeString(eval.RuleName), html.EscapeString(eval.Reason), eval.RequiredBreaches)
				}

				wp.safeGoNotify(func() {
					wp.sendNotification(&dbMonitor, subject, message)
				})

				var alertRuleUUID *uuid.UUID
				if eval.RuleID != nil {
					if parsed, err := uuid.Parse(*eval.RuleID); err == nil {
						alertRuleUUID = &parsed
					}
				}

				newIncident := model.Incident{
					MonitorID:     monitor.ID,
					Status:        "ongoing",
					Severity:      "medium",
					StartedAt:     time.Now(),
					Summary:       fmt.Sprintf("%s: %s", eval.RuleName, eval.Reason),
					TriggeredRule: eval.RuleName,
					AlertRuleID:   alertRuleUUID,
				}

				if err := wp.db.Create(&newIncident).Error; err != nil {
					log.Printf("[Incident] Failed to create incident for monitor %s: %v\n", monitor.Name, err)
				} else {
					incidentIDStr := newIncident.ID.String()
					wp.safeGoNotify(func() {
						wp.dispatcher.Dispatch(&dbMonitor, &incidentIDStr, "monitor.degraded", subject, message)
					})
					log.Printf("[Incident] MONITOR UNRESPONSIVE / DEGRADED: %s (%s). %s\n", monitor.Name, eval.RuleName, eval.Reason)
				}
			} else {
				log.Printf("[Check] Monitor %s breached warning rule %s %d time(s) (waiting for %d)\n", monitor.Name, eval.RuleName, dbMonitor.FailureCount, eval.RequiredBreaches)
			}
		}
	}

	if err := wp.db.Save(&dbMonitor).Error; err != nil {
		log.Printf("Failed to update monitor state in DB: %v\n", err)
	} else if dbMonitor.Status != prevStatus || dbMonitor.FailureCount != prevFailureCount {
		log.Printf("[Check] %s | State Changed: %s -> %s (Fail Count: %d)\n", dbMonitor.Name, prevStatus, dbMonitor.Status, dbMonitor.FailureCount)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if wp.rdb != nil {
		wp.rdb.Del(ctx, cache.MonitorDetailKey(dbMonitor.OrganizationID.String(), dbMonitor.ID.String()))
		if dbMonitor.Status != prevStatus {
			wp.rdb.Del(ctx, cache.MonitorListKey(dbMonitor.OrganizationID.String()))
		}
	}

	if wp.hub != nil {
		updatedData := map[string]interface{}{
			"id":            dbMonitor.ID,
			"status":        dbMonitor.Status,
			"response_time": checkResult.ResponseTimeMs,
			"last_checked":  checkResult.CheckedAt,
			"failure_count": dbMonitor.FailureCount,
		}

		wp.hub.Hub.BroadcastToOrg(
			dbMonitor.OrganizationID.String(),
			"MONITOR_UPDATED",
			updatedData,
		)
	}

	return &checkResult
}

func (wp *WorkerPool) sendNotification(monitor *model.Monitor, subject, message string) {
	if wp.emailSender == nil {
		return
	}

	monitorID := monitor.ID

	// 1. Collect assignee emails
	var pics []model.MonitorPIC
	if err := wp.db.Preload("User").Where("monitor_id = ?", monitorID).Find(&pics).Error; err != nil {
		log.Printf("[Notification] Failed to fetch assignees for monitor %s: %v\n", monitor.Name, err)
		return
	}

	var recipients []string
	for _, pic := range pics {
		if pic.User.Email != "" {
			recipients = append(recipients, pic.User.Email)
		}
	}

	// 2. If no assignees assigned or no valid emails found, skip direct email dispatch (assignee is optional)
	if len(recipients) == 0 {
		log.Printf("[Notification] No assignees assigned for monitor %s, skipping direct email\n", monitor.Name)
		return
	}

	// 3. Send email and log the result correctly
	logMsg := fmt.Sprintf("Monitor '%s' (%s): %s", monitor.Name, monitor.Target, subject)
	_, err := wp.emailSender.SendEmail(recipients, subject, message)
	if err != nil {
		log.Printf("[Notification] Failed to send email for monitor %s: %v\n", monitor.Name, err)

		errDetail := err.Error()
		wp.db.Create(&model.NotificationLog{
			MonitorID:   &monitorID,
			Channel:     "email_assignee",
			Message:     logMsg,
			Status:      "failed",
			ErrorDetail: errDetail,
			SentAt:      time.Now(),
		})
		return
	}

	log.Printf("[Notification] Email sent successfully for monitor %s to %v\n", monitor.Name, recipients)
	wp.db.Create(&model.NotificationLog{
		MonitorID: &monitorID,
		Channel:   "email_assignee",
		Message:   logMsg,
		Status:    "sent",
		SentAt:    time.Now(),
	})
}
