package worker

import (
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
)

type Job struct {
	Monitor model.Monitor
}

type WorkerPool struct {
	db               *gorm.DB
	cache            *cache.MemoryCache
	jobQueue         chan Job
	numWorkers       int
	emailSender      email.EmailSender
	hub              *websocket.WsHandler
	dispatcher       *notification.Dispatcher
	notifySem        chan struct{}
	notifyWg         sync.WaitGroup
	inFlightMonitors sync.Map
	retryDelay       time.Duration
	disableRetry     bool
}

func NewWorkerPool(db *gorm.DB, cacheStore *cache.MemoryCache, numWorkers int, sender email.EmailSender, hub *websocket.WsHandler) *WorkerPool {
	return &WorkerPool{
		db:          db,
		cache:       cacheStore,
		jobQueue:    make(chan Job, 1000),
		numWorkers:  numWorkers,
		emailSender: sender,
		hub:         hub,
		dispatcher:  notification.NewDispatcher(db, sender),
		notifySem:   make(chan struct{}, 50), // max 50 concurrent notification goroutines
		retryDelay:  2 * time.Second,
	}
}

func (wp *WorkerPool) SetRetryDelay(d time.Duration) {
	if d < 0 {
		wp.disableRetry = true
		wp.retryDelay = 0
		return
	}
	wp.disableRetry = false
	wp.retryDelay = d
}

func (wp *WorkerPool) GetDB() *gorm.DB {
	return wp.db
}

func (wp *WorkerPool) safeGoNotify(fn func()) {
	wp.notifyWg.Add(1)
	go func() {
		defer wp.notifyWg.Done()
		wp.notifySem <- struct{}{}
		defer func() { <-wp.notifySem }()
		fn()
	}()
}

func (wp *WorkerPool) WaitNotifications() {
	wp.notifyWg.Wait()
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
	// Clean up stale SSL/Domain status for monitors where check is disabled or non-HTTP
	wp.db.Model(&model.Monitor{}).
		Where("(enable_ssl_domain_check = false OR type != 'http') AND (ssl_status != 'none' OR domain_status != 'none')").
		Updates(map[string]interface{}{
			"ssl_status":                    "none",
			"domain_status":                 "none",
			"ssl_days_left":                 nil,
			"domain_days_left":              nil,
			"ssl_expires_at":                nil,
			"domain_expires_at":             nil,
			"ssl_error":                     "",
			"domain_error":                  "",
			"last_ssl_checked_at":           nil,
			"last_domain_checked_at":        nil,
			"last_ssl_sent_reminder_day":    nil,
			"last_domain_sent_reminder_day": nil,
		})

	var monitors []model.Monitor
	err := wp.db.Where("is_active = true AND type = 'http' AND enable_ssl_domain_check = true").Find(&monitors).Error
	if err != nil {
		log.Printf("[SSL/Domain Scheduler Error] Failed to fetch monitors: %v", err)
		return
	}

	log.Printf("[SSL/Domain Scheduler] Running daily check for %d monitors", len(monitors))
	for _, m := range monitors {
		mon := m
		go func(m model.Monitor) {
			ProcessSSLDomainCheck(wp.db, wp.emailSender, wp.dispatcher, &m)
			if wp.cache != nil {
				orgID := m.OrganizationID.String()
				wp.cache.Del(cache.MonitorListKey(orgID), cache.MonitorDetailKey(orgID, m.ID.String()), cache.MonitorOverviewKey(orgID, 30))
			}
		}(mon)
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

func shouldRetryCheck(monitor *model.Monitor, statusCode int, checkErr error) bool {
	if checkErr != nil {
		msg := strings.ToLower(checkErr.Error())
		if strings.Contains(msg, "ssrf protection") ||
			strings.Contains(msg, "prohibited") ||
			strings.Contains(msg, "invalid target") ||
			strings.Contains(msg, "missing hostname") ||
			strings.Contains(msg, "cannot be empty") {
			return false
		}
		return true
	}

	if monitor.Type == "http" {
		if statusCode >= 500 || (monitor.ExpectedStatusCode > 0 && statusCode != monitor.ExpectedStatusCode) {
			return true
		}
	}

	return false
}

func (wp *WorkerPool) executeSingleCheck(monitor *model.Monitor, timeout time.Duration) (statusCode int, responseTime time.Duration, responseSnippet string, checkErr error) {
	switch monitor.Type {
	case "http":
		statusCode, responseTime, responseSnippet, checkErr = checker.CheckHTTPWithOptions(monitor.Target, timeout, checker.HTTPCheckOptions{
			Method:         monitor.HTTPMethod,
			HeadersJSON:    monitor.HTTPHeaders,
			Body:           monitor.HTTPBody,
			ExpectedStatus: monitor.ExpectedStatusCode,
		})
	case "tcp":
		responseTime, checkErr = checker.CheckTCP(monitor.Target, timeout)
	case "ping":
		responseTime, checkErr = checker.CheckPing(monitor.Target, timeout)
	}
	return statusCode, responseTime, responseSnippet, checkErr
}

func (wp *WorkerPool) syncAndCheckMaintenance(monitorID uuid.UUID, now time.Time) (*model.ScheduledMaintenance, bool) {
	if wp.db == nil {
		return nil, false
	}

	// 1. Auto-transition scheduled -> in_progress
	wp.db.Model(&model.ScheduledMaintenance{}).
		Where("status = ? AND scheduled_start <= ? AND scheduled_end >= ?", "scheduled", now, now).
		Updates(map[string]interface{}{
			"status":       "in_progress",
			"actual_start": now,
		})

	// 2. Auto-transition scheduled/in_progress -> completed
	wp.db.Model(&model.ScheduledMaintenance{}).
		Where("status IN ? AND scheduled_end < ?", []string{"scheduled", "in_progress"}, now).
		Updates(map[string]interface{}{
			"status":     "completed",
			"actual_end": now,
		})

	// 3. Query active maintenance for this monitor
	var activeMaint model.ScheduledMaintenance
	err := wp.db.Table("scheduled_maintenances").
		Select("scheduled_maintenances.*").
		Joins("INNER JOIN maintenance_monitors ON maintenance_monitors.maintenance_id = scheduled_maintenances.id").
		Where(
			"maintenance_monitors.monitor_id = ? AND scheduled_maintenances.status IN ? AND scheduled_maintenances.scheduled_start <= ? AND scheduled_maintenances.scheduled_end >= ?",
			monitorID, []string{"scheduled", "in_progress"}, now, now,
		).
		Order("scheduled_maintenances.scheduled_start ASC").
		Limit(1).
		Find(&activeMaint).Error
	if err == nil && activeMaint.ID != uuid.Nil {
		return &activeMaint, true
	}

	return nil, false
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

	statusCode, responseTime, responseSnippet, checkErr := wp.executeSingleCheck(monitor, timeout)
	if !wp.disableRetry && shouldRetryCheck(monitor, statusCode, checkErr) {
		if wp.retryDelay > 0 {
			time.Sleep(wp.retryDelay)
		}
		log.Printf("[Worker] Immediate retry triggered for monitor %s (%s) after initial failure\n", monitor.Name, monitor.Target)
		statusCode, responseTime, responseSnippet, checkErr = wp.executeSingleCheck(monitor, timeout)
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
		if err := wp.db.Where("organization_id = ? AND is_default = true", dbMonitor.OrganizationID).Limit(1).Find(&defaultTemplate).Error; err == nil && defaultTemplate.ID != uuid.Nil {
			_ = wp.db.Where("template_id = ? AND is_active = true", defaultTemplate.ID).Find(&effectiveRules).Error
		}
	}

	prevStatus := dbMonitor.Status
	prevFailureCount := dbMonitor.FailureCount

	activeMaint, inMaintenance := wp.syncAndCheckMaintenance(monitor.ID, time.Now().UTC())

	eval := EvaluateRules(&dbMonitor, effectiveRules, statusCode, responseTime, checkErr)

	errMsg := ""
	if checkErr != nil {
		errMsg = checker.FormatErrorMessage(checkErr, timeout)
	} else if eval.IsBreached {
		errMsg = eval.Reason
	}

	if responseSnippet != "" && statusCode >= 400 && errMsg != "" {
		errMsg = fmt.Sprintf("%s — Response: %s", errMsg, responseSnippet)
	}

	statusCheck := "up"
	if eval.IsBreached {
		if inMaintenance {
			statusCheck = "maintenance"
			errMsg = fmt.Sprintf("[Maintenance: %s] %s", activeMaint.Title, errMsg)
		} else if eval.Severity == "critical" {
			statusCheck = "down"
		} else {
			statusCheck = "degraded"
		}
	}

	checkResult := model.Check{
		MonitorID:       monitor.ID,
		Status:          statusCheck,
		StatusCode:      statusCode,
		ResponseTimeMs:  int(responseTime.Milliseconds()),
		ErrorMessage:    errMsg,
		ResponseSnippet: responseSnippet,
		CheckedAt:       time.Now(),
	}

	if err := wp.db.Create(&checkResult).Error; err != nil {
		log.Printf("Failed to create check for monitor %s: %v\n", monitor.Name, err)
	}

	if inMaintenance && eval.IsBreached {
		log.Printf("[Maintenance] Suppressed alert/incident for monitor %s during maintenance '%s'\n", dbMonitor.Name, activeMaint.Title)
	} else if !eval.IsBreached {
		// Success / Healthy Scenario
		if dbMonitor.Status == "down" || dbMonitor.Status == "degraded" {
			prevStatus := dbMonitor.Status
			dbMonitor.Status = "up"
			dbMonitor.FailureCount = 0

			subject := fmt.Sprintf("[Netra Uptime] Resolved: %s is back UP", monitor.Name)
			message := fmt.Sprintf("<h3>Service Restored</h3><p>Your monitor <b>%s</b> (%s) has returned to healthy state (previously %s).</p>", html.EscapeString(dbMonitor.Name), html.EscapeString(dbMonitor.Target), html.EscapeString(prevStatus))

			monCopy := dbMonitor
			wp.safeGoNotify(func() {
				wp.sendNotification(&monCopy, subject, message)
			})
			wp.safeGoNotify(func() {
				wp.dispatcher.Dispatch(&monCopy, nil, "monitor.up", subject, message)
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

				monCopy := dbMonitor
				wp.safeGoNotify(func() {
					wp.sendNotification(&monCopy, subject, message)
				})

				var alertRuleUUID *uuid.UUID
				if eval.RuleID != nil {
					if parsed, err := uuid.Parse(*eval.RuleID); err == nil {
						alertRuleUUID = &parsed
					}
				}

				newIncident := model.Incident{
					MonitorID:       monitor.ID,
					Status:          "ongoing",
					Severity:        ResolveIncidentSeverity(dbMonitor.Criticality, eval.Severity),
					StartedAt:       time.Now(),
					Summary:         fmt.Sprintf("%s: %s", eval.RuleName, eval.Reason),
					TriggeredRule:   eval.RuleName,
					ResponseSnippet: responseSnippet,
					AlertRuleID:     alertRuleUUID,
				}

				if err := wp.db.Create(&newIncident).Error; err != nil {
					log.Printf("[Incident] Failed to create incident for monitor %s: %v\n", monitor.Name, err)
				} else {
					incidentIDStr := newIncident.ID.String()
					wp.safeGoNotify(func() {
						wp.dispatcher.Dispatch(&monCopy, &incidentIDStr, "monitor.down", subject, message)
					})
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

				monCopy := dbMonitor
				wp.safeGoNotify(func() {
					wp.sendNotification(&monCopy, subject, message)
				})

				var alertRuleUUID *uuid.UUID
				if eval.RuleID != nil {
					if parsed, err := uuid.Parse(*eval.RuleID); err == nil {
						alertRuleUUID = &parsed
					}
				}

				newIncident := model.Incident{
					MonitorID:       monitor.ID,
					Status:          "ongoing",
					Severity:        "medium",
					StartedAt:       time.Now(),
					Summary:         fmt.Sprintf("%s: %s", eval.RuleName, eval.Reason),
					TriggeredRule:   eval.RuleName,
					ResponseSnippet: responseSnippet,
					AlertRuleID:     alertRuleUUID,
				}

				if err := wp.db.Create(&newIncident).Error; err != nil {
					log.Printf("[Incident] Failed to create incident for monitor %s: %v\n", monitor.Name, err)
				} else {
					incidentIDStr := newIncident.ID.String()
					wp.safeGoNotify(func() {
						wp.dispatcher.Dispatch(&monCopy, &incidentIDStr, "monitor.degraded", subject, message)
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

	if wp.cache != nil {
		wp.cache.Del(cache.MonitorDetailKey(dbMonitor.OrganizationID.String(), dbMonitor.ID.String()))
		if dbMonitor.Status != prevStatus {
			wp.cache.Del(cache.MonitorListKey(dbMonitor.OrganizationID.String()), cache.MonitorOverviewKey(dbMonitor.OrganizationID.String(), 30))
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
