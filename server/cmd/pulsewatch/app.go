package main

import (
	"PulseWatch/config"
	"PulseWatch/internal/alertrule"
	"PulseWatch/internal/apikey"
	"PulseWatch/internal/audit"
	"PulseWatch/internal/auth"
	"PulseWatch/internal/health"
	"PulseWatch/internal/incident"
	"PulseWatch/internal/monitor"
	"PulseWatch/internal/notification"
	"PulseWatch/internal/organization"
	"PulseWatch/internal/publicapi"
	"PulseWatch/internal/router"
	"PulseWatch/internal/statuspage"
	"PulseWatch/internal/worker"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func BuildHandler(db *gorm.DB, rdb *redis.Client, wp *worker.WorkerPool, cfg *config.Config) router.Handlers {
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo, rdb)
	authHandler := auth.NewAuthHandler(authService)

	auditRepo := audit.NewAuditRepository(db)
	auditService := audit.NewAuditService(auditRepo)
	auditHandler := audit.NewAuditHandler(auditService)

	organizationRepo := organization.NewOrganizationRepository(db)
	organizationService := organization.NewOrganizationService(organizationRepo, auditService)
	organizationHandler := organization.NewOrganizationHandler(organizationService)

	monitorRepo := monitor.NewMonitorRepository(db)
	monitorServise := monitor.NewMonitorService(monitorRepo, rdb, organizationService, auditService, wp)
	monitorHandler := monitor.NewMonitorHandler(monitorServise)

	incidentRepo := incident.NewMonitorRepository(db)
	incidentService := incident.NewIncidentService(incidentRepo)
	incidentHandler := incident.NewIncidentHandler(incidentService)

	statuspageRepo := statuspage.NewStatusPageRepository(db)
	statuspageService := statuspage.NewStatusPageService(statuspageRepo, organizationRepo, rdb, auditService)
	statuspageHandler := statuspage.NewStatusPageHandler(statuspageService)

	apiKeyRepo := apikey.NewAPIKeyRepository(db)
	apiKeyService := apikey.NewAPIKeyService(apiKeyRepo, auditService)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyService)

	publicAPIHandler := publicapi.NewPublicAPIHandler(monitorServise, incidentService, statuspageService)

	notificationRepo := notification.NewChannelRepository(db)
	notificationService := notification.NewChannelService(notificationRepo)
	notificationHandler := notification.NewChannelHandler(notificationService)

	alertRuleRepo := alertrule.NewRepository(db)
	alertRuleService := alertrule.NewService(alertRuleRepo)
	alertRuleHandler := alertrule.NewHandler(alertRuleService)

	healthHandler := health.NewHealthHandler(db, rdb, cfg)

	return router.Handlers{
		AuthHandler:         authHandler,
		OrganizationHandler: organizationHandler,
		MonitorHandler:      monitorHandler,
		IncidentHandler:     incidentHandler,
		StatusPageHandler:   statuspageHandler,
		AuditHandler:        auditHandler,
		APIKeyHandler:       apiKeyHandler,
		PublicAPIHandler:    publicAPIHandler,
		APIKeyService:       apiKeyService,
		NotificationHandler: notificationHandler,
		AlertRuleHandler:    alertRuleHandler,
		HealthHandler:       healthHandler,
	}
}
