package main

import (
	"log"
	"netra-uptime/config"
	"netra-uptime/internal/middleware"
	"netra-uptime/internal/model"
	"netra-uptime/internal/monitor"
	"netra-uptime/internal/router"
	"netra-uptime/internal/websocket"
	"netra-uptime/internal/worker"
	"netra-uptime/pkg/cache"
	"netra-uptime/pkg/database"
	"netra-uptime/pkg/email"
	"netra-uptime/pkg/utils"
	"os"
	"strings"
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
		&model.AlertRuleTemplate{}, &model.ScheduledMaintenance{}, &model.MaintenanceMonitor{}, &model.ComponentGroup{})

	sender, err := email.NewSMTPSender()
	if err != nil {
		log.Printf("Warning: Failed to initialize SMTP email sender: %v (email alerts may be disabled)\n", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, db)

	wp := worker.NewWorkerPool(db, 10, sender, wsHandler)
	wp.Start()

	monitorRepo := monitor.NewMonitorRepository(db, config)
	sched := worker.NewScheduler(monitorRepo, wp)
	sched.Start()

	retention := worker.NewRetentionJob(db)
	retention.Start()

	r := gin.Default()

	// Disable untrusted proxies by default to prevent IP spoofing / rate limit bypass (VULN-21)
	if trustedProxies := os.Getenv("TRUSTED_PROXIES"); trustedProxies != "" {
		proxies := strings.Split(trustedProxies, ",")
		for i := range proxies {
			proxies[i] = strings.TrimSpace(proxies[i])
		}
		_ = r.SetTrustedProxies(proxies)
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("nohtml", utils.NoHTMLValidator); err != nil {
			log.Fatalf("Failed to register 'nohtml' validator: %v", err)
		}
	}

	r.Use(middleware.SetupCORS())
	r.Use(middleware.SecurityHeaders())

	rdb := cache.NewRedisClient(config)

	handlers := BuildHandler(db, rdb, wp, config, wsHub)
	router.SetupRouter(r, db, rdb, handlers, wsHandler)

	addr := ":27346"
	if config.Port != "" {
		addr = ":" + config.Port
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run on %s: %v", addr, err)
	}
}
