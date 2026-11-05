package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"netra-uptime/config"
	"netra-uptime/internal/alertrule"
	"netra-uptime/internal/alerttemplate"
	"netra-uptime/internal/apikey"
	"netra-uptime/internal/audit"
	"netra-uptime/internal/auth"
	"netra-uptime/internal/health"
	"netra-uptime/internal/incident"
	"netra-uptime/internal/middleware"
	"netra-uptime/internal/model"
	"netra-uptime/internal/monitor"
	"netra-uptime/internal/notification"
	"netra-uptime/internal/organization"
	"netra-uptime/internal/publicapi"
	"netra-uptime/internal/router"
	"netra-uptime/internal/statuspage"
	"netra-uptime/internal/websocket"
	"netra-uptime/internal/worker"
	"netra-uptime/pkg/cache"
	"netra-uptime/pkg/database"
	"netra-uptime/pkg/email"
	"netra-uptime/pkg/utils"
)

type App struct {
	Config     *config.Config
	Server     *http.Server
	DB         *gorm.DB
	CacheStore *cache.MemoryCache
	WorkerPool *worker.WorkerPool
	Scheduler  *worker.Scheduler
	Retention  *worker.RetentionJob
	Router     *gin.Engine
}

// Setup initializes database, services, worker pool, scheduler, and gin engine.
func Setup(cfg *config.Config) (*App, error) {
	// Release mode by default unless GIN_MODE=debug or APP_ENV=development
	if os.Getenv("GIN_MODE") == "" && cfg.AppEnv != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{}, &model.Check{}, &model.Incident{}, &model.IncidentUpdate{},
		&model.Monitor{}, &model.NotificationChannel{}, &model.NotificationLog{},
		&model.Organization{}, &model.OrganizationMember{}, &model.RefreshToken{},
		&model.StatusPage{}, &model.StatusPageMonitor{}, &model.OrganizationInvitation{},
		&model.PasswordResetToken{}, &model.IncidentAssignee{}, &model.MonitorPIC{},
		&model.AuditLog{}, &model.APIKey{}, &model.MonitorDailyStat{}, &model.AlertRule{},
		&model.AlertRuleTemplate{}, &model.ScheduledMaintenance{}, &model.MaintenanceMonitor{},
		&model.ComponentGroup{},
	); err != nil {
		return nil, fmt.Errorf("database auto-migrate failed: %w", err)
	}

	sender, err := email.NewSMTPSender()
	if err != nil {
		log.Printf("Warning: Failed to initialize SMTP email sender: %v (email alerts may be disabled)\n", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, db)

	cacheStore := cache.NewMemoryCache()

	wp := worker.NewWorkerPool(db, cacheStore, 10, sender, wsHandler)
	wp.Start()

	monitorRepo := monitor.NewMonitorRepository(db, cfg)
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
			return nil, fmt.Errorf("failed to register 'nohtml' validator: %w", err)
		}
	}

	r.Use(middleware.SetupCORS())
	r.Use(middleware.SecurityHeaders())

	handlers := buildHandlers(db, cacheStore, wp, cfg, wsHub)
	router.SetupRouter(r, db, cacheStore, handlers, wsHandler)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	statusStr := "disabled"
	if cfg.AllowPrivateNetworkMonitoring {
		statusStr = "enabled"
	}
	log.Printf("[Netra Uptime] Private network monitoring: %s (%s)\n", statusStr, cfg.AllowPrivateNetworkMonitoringReason)

	return &App{
		Config:     cfg,
		Server:     srv,
		DB:         db,
		CacheStore: cacheStore,
		WorkerPool: wp,
		Scheduler:  sched,
		Retention:  retention,
		Router:     r,
	}, nil
}

// Shutdown gracefully terminates HTTP connections, workers, cache eviction, and closes the database.
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error

	// 1. Shutdown HTTP server gracefully
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			errs = append(errs, fmt.Errorf("http server shutdown error: %w", err))
		}
	}

	// 2. Stop scheduler and background retention
	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}
	if a.Retention != nil {
		a.Retention.Stop()
	}

	// 3. Stop cache eviction goroutine BEFORE closing DB (Requirement #7)
	if a.CacheStore != nil {
		a.CacheStore.Stop()
	}

	// 4. Flush SQLite WAL and close database
	if a.DB != nil {
		if a.Config.DBDriver == "sqlite" {
			// TRUNCATE checkpoint flushes WAL pages into the database file and resets WAL file size to zero
			_ = a.DB.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error
		}
		if sqlDB, err := a.DB.DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("database close error: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

func buildHandlers(db *gorm.DB, cacheStore *cache.MemoryCache, wp *worker.WorkerPool, cfg *config.Config, wsEvicter ...organization.WebSocketEvicter) router.Handlers {
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo, cacheStore, cfg)
	authHandler := auth.NewAuthHandler(authService)

	auditRepo := audit.NewAuditRepository(db)
	auditService := audit.NewAuditService(auditRepo)
	auditHandler := audit.NewAuditHandler(auditService)

	organizationRepo := organization.NewOrganizationRepository(db)
	organizationService := organization.NewOrganizationService(organizationRepo, auditService, cfg)
	if len(wsEvicter) > 0 && wsEvicter[0] != nil {
		organizationService.SetWebSocketEvicter(wsEvicter[0])
	}
	organizationHandler := organization.NewOrganizationHandler(organizationService)

	monitorRepo := monitor.NewMonitorRepository(db, cfg)
	monitorServise := monitor.NewMonitorService(monitorRepo, cacheStore, auditService, wp, cfg)
	monitorHandler := monitor.NewMonitorHandler(monitorServise)

	incidentRepo := incident.NewMonitorRepository(db)
	incidentService := incident.NewIncidentService(incidentRepo)
	incidentHandler := incident.NewIncidentHandler(incidentService)

	statuspageRepo := statuspage.NewStatusPageRepository(db, cfg)
	statuspageService := statuspage.NewStatusPageService(statuspageRepo, organizationRepo, cacheStore, auditService, cfg)
	statuspageHandler := statuspage.NewStatusPageHandler(statuspageService)

	apiKeyRepo := apikey.NewAPIKeyRepository(db)
	apiKeyService := apikey.NewAPIKeyService(apiKeyRepo, auditService)
	organizationService.SetAPIKeyRevoker(apiKeyService)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyService)

	publicAPIHandler := publicapi.NewPublicAPIHandler(monitorServise, incidentService, statuspageService)

	notificationRepo := notification.NewChannelRepository(db)
	notificationService := notification.NewChannelService(notificationRepo)
	notificationHandler := notification.NewChannelHandler(notificationService)

	alertRuleRepo := alertrule.NewRepository(db)
	alertRuleService := alertrule.NewService(alertRuleRepo)
	alertRuleHandler := alertrule.NewHandler(alertRuleService)

	alertTemplateRepo := alerttemplate.NewRepository(db)
	alertTemplateService := alerttemplate.NewService(alertTemplateRepo)
	alertTemplateHandler := alerttemplate.NewHandler(alertTemplateService)

	healthHandler := health.NewHealthHandler(db, cfg)

	return router.Handlers{
		AuthHandler:          authHandler,
		OrganizationHandler:  organizationHandler,
		MonitorHandler:       monitorHandler,
		IncidentHandler:      incidentHandler,
		StatusPageHandler:    statuspageHandler,
		AuditHandler:         auditHandler,
		APIKeyHandler:        apiKeyHandler,
		PublicAPIHandler:     publicAPIHandler,
		APIKeyService:        apiKeyService,
		NotificationHandler:  notificationHandler,
		AlertRuleHandler:     alertRuleHandler,
		AlertTemplateHandler: alertTemplateHandler,
		HealthHandler:        healthHandler,
	}
}
