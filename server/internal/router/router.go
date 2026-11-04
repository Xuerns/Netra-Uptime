package router

import (
	"netra-uptime/config"
	"netra-uptime/internal/alertrule"
	"netra-uptime/internal/alerttemplate"
	"netra-uptime/internal/apikey"
	"netra-uptime/internal/audit"
	"netra-uptime/internal/auth"
	"netra-uptime/internal/health"
	"netra-uptime/internal/incident"
	"netra-uptime/internal/middleware"
	"netra-uptime/internal/monitor"
	"netra-uptime/internal/notification"
	"netra-uptime/internal/organization"
	"netra-uptime/internal/publicapi"
	"netra-uptime/internal/static"
	"netra-uptime/internal/statuspage"
	"netra-uptime/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handlers struct {
	AuthHandler          auth.AuthHandler
	OrganizationHandler  organization.OrganizationHandler
	MonitorHandler       monitor.MonitorHandler
	IncidentHandler      incident.Incidenthandler
	StatusPageHandler    statuspage.StatusPageHandler
	AuditHandler         audit.AuditHandler
	APIKeyHandler        *apikey.APIKeyHandler
	PublicAPIHandler     *publicapi.PublicAPIHandler
	APIKeyService        apikey.APIKeyService
	NotificationHandler  *notification.ChannelHandler
	AlertRuleHandler     *alertrule.Handler
	AlertTemplateHandler *alerttemplate.Handler
	HealthHandler        *health.HealthHandler
}

func SetupRouter(r *gin.Engine, db *gorm.DB, redis *redis.Client, h Handlers, wsHandler *websocket.WsHandler) {
	// Register embedded client static routes & SPA fallback
	static.RegisterStaticRoutes(r)

	api := r.Group("api/v1")

	// System Health Check
	if h.HealthHandler != nil {
		r.GET("/healthz", h.HealthHandler.Check)
		api.GET("/health", h.HealthHandler.Check)
	}

	cfg := config.Load()
	authRateLimit := 60
	if cfg != nil && cfg.AuthRateLimit > 0 {
		authRateLimit = cfg.AuthRateLimit
	}

	authLimiter := middleware.RateLimiter(redis, authRateLimit, 1*time.Minute)
	refreshLimiter := middleware.RateLimiter(redis, 30, 1*time.Minute)
	resetLimiter := middleware.RateLimiter(redis, 3, 15*time.Minute)

	// Public API
	api.POST("/auth/register", authLimiter, h.AuthHandler.Register)
	api.POST("/auth/login", authLimiter, h.AuthHandler.Login)
	api.POST("/auth/logout", authLimiter, h.AuthHandler.Logout)
	api.POST("/auth/refresh", refreshLimiter, h.AuthHandler.RefreshAccessToken)
	api.POST("/auth/forgotpassword", resetLimiter, h.AuthHandler.ForgotPassword)
	api.PATCH("/auth/resetpassword", resetLimiter, h.AuthHandler.ResetPassword)

	api.GET("/public/statuspage/:subdomain", h.StatusPageHandler.GetPublicBySubdomain)
	api.GET("/public/statuspage/:subdomain/summary.json", h.StatusPageHandler.GetPublicSummaryJSON)
	api.GET("/public/statuspage/:subdomain/status.json", h.StatusPageHandler.GetPublicStatusJSON)
	api.GET("/public/statuspage/:subdomain/badge.svg", h.StatusPageHandler.GetStatusBadge)
	// Protected API -> Need Access token
	protected := r.Group("api/v1")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/ws", wsHandler.ServeWS)

	protected.GET("/auth/profile", h.AuthHandler.GetProfile)
	protected.PUT("/auth/profile", h.AuthHandler.UpdateProfile)
	protected.POST("/auth/password/change/init", authLimiter, h.AuthHandler.InitPasswordChange)
	protected.POST("/auth/password/change/verify", authLimiter, h.AuthHandler.VerifyPasswordChange)

	// Organization
	protected.GET("/organization/all", h.OrganizationHandler.GetAllOrganization)
	protected.POST("/organization/create", h.OrganizationHandler.CreateOrganization)
	protected.GET("/invitation", h.OrganizationHandler.GetInvitation)
	protected.PUT("/invitation/action", h.OrganizationHandler.InvitationAction)
	protected.GET("/search/persons", h.OrganizationHandler.SearchPersons)

	// Specific API -> Need Access token and Tenant ID
	orgRoutes := protected.Group("/organization/:orgId")
	orgRoutes.Use(middleware.TenantMiddleware(db))
	orgRoutes.GET("/members", h.OrganizationHandler.GetAllMember)
	orgRoutes.POST("/invite", middleware.RequireRole("owner", "admin"), h.OrganizationHandler.InviteMember)
	orgRoutes.DELETE("/kick", middleware.RequireRole("owner"), h.OrganizationHandler.KickMember)
	orgRoutes.DELETE("/disband", middleware.RequireRole("owner"), h.OrganizationHandler.Disband)
	orgRoutes.PUT("/update", middleware.RequireRole("owner"), h.OrganizationHandler.UpdateOrganization)
	orgRoutes.PUT("/transfer", middleware.RequireRole("owner"), h.OrganizationHandler.TransferOwnership)
	orgRoutes.PUT("/leave", h.OrganizationHandler.LeaveOrganization)
	orgRoutes.PUT("/promote", middleware.RequireRole("owner"), h.OrganizationHandler.PromoteToAdmin)
	orgRoutes.PUT("/demote", middleware.RequireRole("owner"), h.OrganizationHandler.DemoteToMember)

	// Monitor
	orgRoutes.POST("/monitors", middleware.RequireRole("owner"), h.MonitorHandler.CreateMonitor)
	orgRoutes.GET("/monitors", h.MonitorHandler.GetAllMonitor)
	orgRoutes.GET("/monitors/overview", h.MonitorHandler.GetFleetOverview)
	orgRoutes.GET("/monitors/:monitorId", h.MonitorHandler.GetById)
	orgRoutes.GET("/monitors/:monitorId/stats", h.MonitorHandler.GetDailyStats)
	checkLimiter := middleware.RateLimiter(redis, 10, 1*time.Minute)
	orgRoutes.POST("/monitors/:monitorId/check", checkLimiter, h.MonitorHandler.TriggerCheck)
	orgRoutes.PUT("/monitors/:monitorId", middleware.RequireRole("owner", "admin"), h.MonitorHandler.UpdateMonitor)
	orgRoutes.DELETE("/monitors/:monitorId", middleware.RequireRole("owner", "admin"), h.MonitorHandler.DeleteMonitor)

	// Alert Rules
	orgRoutes.GET("/monitors/:monitorId/alert-rules", h.AlertRuleHandler.GetByMonitor)
	orgRoutes.POST("/monitors/:monitorId/alert-rules", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Create)
	orgRoutes.GET("/alert-rules/:ruleId", h.AlertRuleHandler.GetByID)
	orgRoutes.PUT("/alert-rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Update)
	orgRoutes.DELETE("/alert-rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Delete)

	// Alert Rule Templates
	if h.AlertTemplateHandler != nil {
		orgRoutes.GET("/alert-templates", h.AlertTemplateHandler.List)
		orgRoutes.POST("/alert-templates", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.Create)
		orgRoutes.GET("/alert-templates/:templateId", h.AlertTemplateHandler.GetByID)
		orgRoutes.PUT("/alert-templates/:templateId", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.Update)
		orgRoutes.DELETE("/alert-templates/:templateId", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.Delete)
		orgRoutes.POST("/alert-templates/:templateId/rules", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.CreateRule)
		orgRoutes.PUT("/alert-templates/:templateId/rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.UpdateRule)
		orgRoutes.DELETE("/alert-templates/:templateId/rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.DeleteRule)
		orgRoutes.POST("/monitors/:monitorId/template", middleware.RequireRole("owner", "admin"), h.AlertTemplateHandler.AssignToMonitor)
	}

	// Incident
	orgRoutes.GET("/incident/:incidentId", h.IncidentHandler.GetIncidentDetail)
	orgRoutes.GET("/incident", h.IncidentHandler.GetAllIncident)
	orgRoutes.POST("/incident/:incidentId", h.IncidentHandler.AddIncidentUpdate)
	orgRoutes.POST("/incident/:incidentId/join", h.IncidentHandler.JoinAssignee)
	orgRoutes.GET("/incident/:incidentId/updates", h.IncidentHandler.GetMessageAndUser)

	// StatusPage
	orgRoutes.POST("/statuspage", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.CreateStatusPage)
	orgRoutes.GET("/statuspage", h.StatusPageHandler.GetAllStatusPage)
	orgRoutes.GET("/statuspage/:statusPageId", h.StatusPageHandler.GetById)
	orgRoutes.GET("/statuspage/:statusPageId/monitors", h.StatusPageHandler.GetStatusPageMonitor)
	orgRoutes.DELETE("/statuspage/:statusPageId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.DeleteStatusPage)
	orgRoutes.PUT("/statuspage/:statusPageId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.UpdateStatusPage)
	orgRoutes.POST("/statuspage/:statusPageId/maintenance", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.CreateMaintenance)
	orgRoutes.GET("/statuspage/:statusPageId/maintenance", h.StatusPageHandler.GetMaintenances)
	orgRoutes.PUT("/statuspage/:statusPageId/maintenance/:maintenanceId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.UpdateMaintenance)
	orgRoutes.DELETE("/statuspage/:statusPageId/maintenance/:maintenanceId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.DeleteMaintenance)
	orgRoutes.POST("/statuspage/:statusPageId/groups", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.CreateComponentGroup)
	orgRoutes.GET("/statuspage/:statusPageId/groups", h.StatusPageHandler.GetComponentGroups)
	orgRoutes.PUT("/statuspage/:statusPageId/groups/:groupId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.UpdateComponentGroup)
	orgRoutes.DELETE("/statuspage/:statusPageId/groups/:groupId", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.DeleteComponentGroup)
	orgRoutes.PUT("/statuspage/:statusPageId/monitors/:monitorId/group", middleware.RequireRole("owner", "admin"), h.StatusPageHandler.AssignMonitorGroup)

	// Audit Logs
	orgRoutes.GET("/audit-logs", h.AuditHandler.GetOrgAuditLogs)

	// Notification Channels
	orgRoutes.GET("/notifications/channels", h.NotificationHandler.GetAll)
	orgRoutes.GET("/notifications/channels/:channelId", h.NotificationHandler.GetByID)
	orgRoutes.POST("/notifications/channels", middleware.RequireRole("owner", "admin"), h.NotificationHandler.Create)
	orgRoutes.PUT("/notifications/channels/:channelId", middleware.RequireRole("owner", "admin"), h.NotificationHandler.Update)
	orgRoutes.DELETE("/notifications/channels/:channelId", middleware.RequireRole("owner", "admin"), h.NotificationHandler.Delete)
	orgRoutes.GET("/notifications/logs", h.NotificationHandler.GetLogs)

	// API Keys Management
	orgRoutes.GET("/api-keys", h.APIKeyHandler.List)
	orgRoutes.POST("/api-keys", middleware.RequireRole("owner", "admin"), h.APIKeyHandler.Create)
	orgRoutes.DELETE("/api-keys/:keyId", middleware.RequireRole("owner", "admin"), h.APIKeyHandler.Revoke)

	// Public REST API (protected by API Key auth + 60 req/min rate limit)
	publicGroup := r.Group("api/v1/public")
	publicGroup.Use(middleware.APIKeyAuthMiddleware(h.APIKeyService, redis))
	{
		publicGroup.GET("/monitors", h.PublicAPIHandler.ListMonitors)
		publicGroup.GET("/monitors/:id", h.PublicAPIHandler.GetMonitor)
		publicGroup.POST("/monitors", middleware.RequireAPIKeyPermission("write"), h.PublicAPIHandler.CreateMonitor)
		publicGroup.PUT("/monitors/:id", middleware.RequireAPIKeyPermission("write"), h.PublicAPIHandler.UpdateMonitor)
		publicGroup.DELETE("/monitors/:id", middleware.RequireAPIKeyPermission("write"), h.PublicAPIHandler.DeleteMonitor)

		publicGroup.GET("/incidents", h.PublicAPIHandler.ListIncidents)
		publicGroup.POST("/incidents", middleware.RequireAPIKeyPermission("write"), h.PublicAPIHandler.CreateIncidentNote)

		publicGroup.GET("/status-pages", h.PublicAPIHandler.ListStatusPages)
	}
}
