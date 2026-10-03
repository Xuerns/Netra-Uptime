package worker

import (
	"PulseWatch/internal/model"
	"PulseWatch/internal/notification"
	"PulseWatch/internal/websocket"
	"PulseWatch/internal/worker/checker"
	"PulseWatch/pkg/email"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type Job struct {
	Monitor model.Monitor
}

type WorkerPool struct {
	db          *gorm.DB
	jobQueue    chan Job
	numWorkers  int
	emailSender email.EmailSender
	hub         *websocket.WsHandler
	dispatcher  *notification.Dispatcher
}

func NewWorkerPool(db *gorm.DB, numWorkers int, sender email.EmailSender, hub *websocket.WsHandler) *WorkerPool {
	return &WorkerPool{
		db:          db,
		jobQueue:    make(chan Job, 1000),
		numWorkers:  numWorkers,
		emailSender: sender,
		hub:         hub,
		dispatcher:  notification.NewDispatcher(db, sender),
	}
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

func (wp *WorkerPool) worker(id int) {
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

	prevStatus := dbMonitor.Status
	prevFailureCount := dbMonitor.FailureCount

	eval := EvaluateRules(&dbMonitor, dbMonitor.AlertRules, statusCode, responseTime, checkErr)

	errMsg := ""
	if checkErr != nil {
		errMsg = checkErr.Error()
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

			subject := fmt.Sprintf("[PulseWatch] Resolved: %s is back UP", monitor.Name)
			message := fmt.Sprintf("<h3>Service Restored</h3><p>Your monitor <b>%s</b> (%s) has returned to healthy state (previously %s).</p>", dbMonitor.Name, dbMonitor.Target, prevStatus)

			go wp.sendNotification(&dbMonitor, subject, message)
			go wp.dispatcher.Dispatch(&dbMonitor, nil, "monitor.up", subject, message)

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
		if eval.Severity == "critical" {
			dbMonitor.FailureCount++

			if dbMonitor.FailureCount >= eval.RequiredBreaches && dbMonitor.Status != "down" {
				dbMonitor.Status = "down"

				subject := fmt.Sprintf("[PulseWatch] ALERT: %s is DOWN (%s)", monitor.Name, eval.RuleName)
				message := fmt.Sprintf("<h3>Service Down Alert</h3><p>Your monitor <b>%s</b> (%s) triggered alert rule <b>%s</b>.<br>Reason: <i>%s</i> (breached %d consecutive times)</p>", dbMonitor.Name, dbMonitor.Target, eval.RuleName, eval.Reason, eval.RequiredBreaches)

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
		} else if eval.Severity == "warning" {
			// Degraded Warning (e.g. high latency)
			if dbMonitor.Status == "up" {
				dbMonitor.Status = "degraded"
				subject := fmt.Sprintf("[PulseWatch] WARNING: %s is DEGRADED (%s)", monitor.Name, eval.RuleName)
				message := fmt.Sprintf("<h3>Service Degraded Warning</h3><p>Your monitor <b>%s</b> (%s) is experiencing degraded performance.<br>Rule: <b>%s</b><br>Reason: <i>%s</i></p>", dbMonitor.Name, dbMonitor.Target, eval.RuleName, eval.Reason)

				go wp.sendNotification(&dbMonitor, subject, message)
				go wp.dispatcher.Dispatch(&dbMonitor, nil, "monitor.degraded", subject, message)
				log.Printf("[Check] MONITOR DEGRADED: %s (%s). %s\n", monitor.Name, eval.RuleName, eval.Reason)
			}
		}
	}

	if err := wp.db.Save(&dbMonitor).Error; err != nil {
		log.Printf("Failed to update monitor state in DB: %v\n", err)
	} else if dbMonitor.Status != prevStatus || dbMonitor.FailureCount != prevFailureCount {
		log.Printf("[Check] %s | State Changed: %s -> %s (Fail Count: %d)\n", dbMonitor.Name, prevStatus, dbMonitor.Status, dbMonitor.FailureCount)
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

