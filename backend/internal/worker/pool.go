package worker

import (
	"PulseWatch/internal/model"
	"PulseWatch/internal/websocket"
	"PulseWatch/internal/worker/checker"
	"PulseWatch/pkg/email"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
}

func NewWorkerPool(db *gorm.DB, numWorkers int, sender email.EmailSender, hub *websocket.WsHandler) *WorkerPool {
	return &WorkerPool{
		db:          db,
		jobQueue:    make(chan Job, 1000),
		numWorkers:  numWorkers,
		emailSender: sender,
		hub:         hub,
	}
}

func (wp *WorkerPool) Start() {
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
		go ProcessSSLDomainCheck(wp.db, wp.emailSender, &mon)
		// Small jitter delay between domain queries to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}
}

func (wp *WorkerPool) TriggerSingleSSLDomainCheck(monitor *model.Monitor) {
	go ProcessSSLDomainCheck(wp.db, wp.emailSender, monitor)
}

func (wp *WorkerPool) QueueJob(job Job) {
	wp.jobQueue <- job
}

func (wp *WorkerPool) worker(id int) {
	log.Printf("[Worker %d] Started", id)
	for job := range wp.jobQueue {
		wp.procces(&job.Monitor)
	}
}

func (wp *WorkerPool) procces(monitor *model.Monitor) {
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var isCheckSuccess bool = true
	var statusCode int
	var responseTime time.Duration
	var checkErr error

	switch monitor.Type {
	case "http":
		statusCode, responseTime, checkErr = checker.CheckHTTP(monitor.Target, timeout, monitor.ExpectedStatusCode)
		if checkErr != nil || (monitor.ExpectedStatusCode != 0 && statusCode != monitor.ExpectedStatusCode) {
			isCheckSuccess = false
		}
	case "tcp":
		responseTime, checkErr = checker.CheckTCP(monitor.Target, timeout)
		if checkErr != nil {
			isCheckSuccess = false
		}
	case "ping":
		responseTime, checkErr = checker.CheckPing(monitor.Target, timeout)
		if checkErr != nil {
			isCheckSuccess = false
		}
	}

	errMsg := ""
	if checkErr != nil {
		errMsg = checkErr.Error()
	} else if !isCheckSuccess && monitor.Type == "http" {
		errMsg = "status code mismatch"
	}

	statusCheck := "up"
	if !isCheckSuccess {
		statusCheck = "down"
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

	var dbMonitor model.Monitor
	if err := wp.db.First(&dbMonitor, "id = ?", monitor.ID).Error; err != nil {
		log.Printf("Failed to fetch monitor for state evaluation: %v\n", err)
		return
	}

	const threshold = 3

	if isCheckSuccess {
		// Success Scenario
		if dbMonitor.Status == "down" {
			dbMonitor.Status = "up"
			dbMonitor.FailureCount = 0

			subject := fmt.Sprintf("[PulseWatch] Resolved: %s is back UP", monitor.Name)
			message := fmt.Sprintf("<h3>Service restored</h3><p>Your monitor <b>%s</b> (%s) has returned to healty state.</p>", dbMonitor.Name, dbMonitor.Target)

			go wp.sendNotification(&dbMonitor, subject, message)

			now := time.Now()
			err := wp.db.Model(&model.Incident{}).Where("monitor_id = ? AND status = ?", monitor.ID, "ongoing").Updates(map[string]interface{}{
				"status":      "resolved",
				"resolved_at": now,
				"summary":     "Resolved automatically because service returned to healthy state.",
			}).Error

			if err != nil {
				log.Printf("[Incident] Failed to resolve incident for monitor %s: %v\n", monitor.Name, err)
			} else {
				log.Printf("[Incident] Monitor Resolved: %s is back UP!\n", monitor.Name)
			}
		} else {
			// Status already up, and also for reset count
			dbMonitor.FailureCount = 0
		}
	} else {
		// Failed Scenario
		dbMonitor.FailureCount++

		if dbMonitor.FailureCount >= threshold && dbMonitor.Status == "up" {
			dbMonitor.Status = "down"

			subject := fmt.Sprintf("[PulseWatch] ALERT %s is DOWN", monitor.Name)
			message := fmt.Sprintf("<h3>Service Down Alert</h3><p>Your monitor <b>%s</b> (%s) is DOWN after %d consecutive failures.<br>Error: <i>%s</i></p>", dbMonitor.Name, dbMonitor.Target, threshold, errMsg)

			go wp.sendNotification(&dbMonitor, subject, message)

			newIncident := model.Incident{
				MonitorID: monitor.ID,
				Status:    "ongoing",
				StartedAt: time.Now(),
				Summary:   "Monitor down: " + errMsg,
			}

			if err := wp.db.Create(&newIncident).Error; err != nil {
				log.Printf("[Incident] Failed to create incident for monitor %s: %v\n", monitor.Name, err)
			} else {
				log.Printf("[Incident] MONITOR DOWN TRIGGERED: %s is DOWN after %d failures. Error: %s\n", monitor.Name, threshold, errMsg)
			}

		} else {
			log.Printf("[Check] Monitor %s failed %d time(s) (waiting for threshold %d to trigger DOWN)\n", monitor.Name, dbMonitor.FailureCount, threshold)
		}
	}

	if err := wp.db.Save(&dbMonitor).Error; err != nil {
		log.Printf("Failed to update monitor state in DB: %v\n", err)
	} else {
		log.Printf("[Check] %s | Current State: %s (Fail Count: %d)\n", dbMonitor.Name, dbMonitor.Status, dbMonitor.FailureCount)
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
}

func (wp *WorkerPool) sendNotification(monitor *model.Monitor, subject, message string) {
	if wp.emailSender == nil {
		return
	}

	var members []model.MonitorPIC
	err := wp.db.Preload("User").Where("monitor_id = ?", monitor.ID).Find(&members).Error
	if err != nil {
		log.Printf("[Notification] Failed to find members\n")
		return
	}

	var recipients []string
	for _, member := range members {
		if member.User.Email != "" {
			recipients = append(recipients, member.User.Email)
		}
	}

	if len(recipients) == 0 {
		log.Printf("[Notification] No recipient found\n")
		return
	}

	// Send Email
	_, err = wp.emailSender.SendEmail(recipients, subject, message)
	if err != nil {
		log.Printf("[Notification] Failed to send email.\n Details %v:\n", err)

		wp.db.Create(&model.NotificationLog{
			ChannelID: uuid.Nil,
			Status:    "failed",
			SentAt:    time.Now(),
		})
	}

	log.Printf("[Notification] Email send successfully\n")
	wp.db.Create(&model.NotificationLog{
		Status: "success",
		SentAt: time.Now(),
	})

}
