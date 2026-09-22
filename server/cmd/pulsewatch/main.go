package main

import (
	"PulseWatch/config"
	"PulseWatch/internal/middleware"
	"PulseWatch/internal/model"
	"PulseWatch/internal/monitor"
	"PulseWatch/internal/router"
	"PulseWatch/internal/websocket"
	"PulseWatch/internal/worker"
	"PulseWatch/pkg/cache"
	"PulseWatch/pkg/database"
	"PulseWatch/pkg/email"
	"PulseWatch/pkg/utils"
	"log"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func main() {

	config := config.Load()
	db, err := database.NewPostgres(config)

	if err != nil {
		log.Fatalf("Failed to connect database: %v\n", err)
	}

	_ = db.AutoMigrate(&model.User{}, &model.Check{}, &model.Incident{}, &model.IncidentUpdate{}, &model.Monitor{}, &model.NotificationChannel{},
		&model.NotificationLog{}, &model.Organization{}, &model.OrganizationMember{}, &model.RefreshToken{}, &model.StatusPage{}, &model.StatusPageMonitor{},
		&model.Subscription{}, &model.OrganizationInvitation{}, &model.PasswordResetToken{}, &model.IncidentAssignee{}, &model.MonitorPIC{}, &model.AuditLog{}, &model.APIKey{}, &model.MonitorDailyStat{}, &model.AlertRule{},
		&model.ScheduledMaintenance{}, &model.MaintenanceMonitor{}, &model.ComponentGroup{})

	sender, err := email.NewSMTPSender()
	if err != nil {
		log.Printf("Warning: Failed to initialize SMTP email sender: %v (email alerts may be disabled)\n", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, db)

	wp := worker.NewWorkerPool(db, 10, sender, wsHandler)
	wp.Start()

	monitorRepo := monitor.NewMonitorRepository(db)
	sched := worker.NewScheduler(monitorRepo, wp)
	sched.Start()

	retention := worker.NewRetentionJob(db)
	retention.Start()

	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}

	r.Use(middleware.SetupCORS())

	rdb := cache.NewRedisClient(config)

	handlers := BuildHandler(db, rdb)
	router.SetupRouter(r, db, rdb, handlers, wsHandler)

	r.Run()
}
