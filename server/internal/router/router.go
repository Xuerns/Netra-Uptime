package router

import (
	"PulseWatch/internal/alertrule"
	"PulseWatch/internal/apikey"
	"PulseWatch/internal/audit"
	"PulseWatch/internal/auth"
	"PulseWatch/internal/incident"
	"PulseWatch/internal/middleware"
	"PulseWatch/internal/monitor"
	"PulseWatch/internal/notification"
	"PulseWatch/internal/organization"
	"PulseWatch/internal/publicapi"
	"PulseWatch/internal/static"
	"PulseWatch/internal/statuspage"
	"PulseWatch/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handlers struct {
	AuthHandler         auth.AuthHandler
	OrganizationHandler organization.OrganizationHandler
	MonitorHandler      monitor.MonitorHandler
	IncidentHandler     incident.Incidenthandler
	StatusPageHandler   statuspage.StatusPageHandler
	AuditHandler        audit.AuditHandler
	APIKeyHandler       *apikey.APIKeyHandler
	PublicAPIHandler    *publicapi.PublicAPIHandler
	APIKeyService       apikey.APIKeyService
	NotificationHandler *notification.ChannelHandler
	AlertRuleHandler    *alertrule.Handler
}

func SetupRouter(r *gin.Engine, db *gorm.DB, redis *redis.Client, h Handlers, wsHandler *websocket.WsHandler) {
	// Register embedded client static routes & SPA fallback
	static.RegisterStaticRoutes(r)

	api := r.Group("api/v1")

	authLimiter := middleware.RateLimiter(redis, 5, 1*time.Minute)
	resetLimiter := middleware.RateLimiter(redis, 3, 15*time.Minute)



	// Public API
	api.POST("/auth/register", authLimiter, h.AuthHandler.Register)
	api.POST("/auth/login", authLimiter, h.AuthHandler.Login)
	api.POST("/auth/logout", authLimiter, h.AuthHandler.Logout)
	api.POST("/auth/refresh", authLimiter, h.AuthHandler.RefreshAccessToken)
	api.POST("/auth/forgotpassword", resetLimiter, h.AuthHandler.ForgotPassword)
	api.PATCH("/auth/resetpassword", resetLimiter, h.AuthHandler.ResetPassword)
	api.POST("/auth", authLimiter, h.AuthHandler.OauthLoginAndRegister)

	api.GET("/public/statuspage/:subdomain", h.StatusPageHandler.GetPublicBySubdomain)
	api.GET("/public/statuspage/:subdomain/summary.json", h.StatusPageHandler.GetPublicSummaryJSON)
	api.GET("/public/statuspage/:subdomain/status.json", h.StatusPageHandler.GetPublicStatusJSON)
	api.GET("/public/statuspage/:subdomain/badge.svg", h.StatusPageHandler.GetStatusBadge)
	// Protected API -> Need Access token
	protected := r.Group("api/v1")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/ws", wsHandler.ServeWS)

	protected.GET("/auth/profile", h.AuthHandler.GetProfile)
	protected.POST("/auth/password/change/init", h.AuthHandler.InitPasswordChange)
	protected.POST("/auth/password/change/verify", h.AuthHandler.VerifyPasswordChange)

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
	orgRoutes.POST("/invite", middleware.RequireRole("owner"), h.OrganizationHandler.InviteMember)
	orgRoutes.DELETE("/kick", middleware.RequireRole("owner"), h.OrganizationHandler.KickMember)
	orgRoutes.DELETE("/disband", middleware.RequireRole("owner"), h.OrganizationHandler.Disband)
	orgRoutes.PUT("/update", middleware.RequireRole("owner"), h.OrganizationHandler.UpdateOrganization)
	orgRoutes.PUT("/transfer", middleware.RequireRole("owner"), h.OrganizationHandler.TransferOwnership)
	orgRoutes.PUT("/leave", h.OrganizationHandler.LeaveOrganization)
	orgRoutes.PUT("/promote", middleware.RequireRole("owner", "admin"), h.OrganizationHandler.PromoteToAdmin)
	orgRoutes.PUT("/demote", middleware.RequireRole("owner"), h.OrganizationHandler.DemoteToMember)

	// Monitor
	orgRoutes.POST("/monitors", middleware.RequireRole("owner"), h.MonitorHandler.CreateMonitor)
	orgRoutes.GET("/monitors", h.MonitorHandler.GetAllMonitor)
	orgRoutes.GET("/monitors/:monitorId", h.MonitorHandler.GetById)
	orgRoutes.GET("/monitors/:monitorId/stats", h.MonitorHandler.GetDailyStats)
	orgRoutes.PUT("/monitors/:monitorId", h.MonitorHandler.UpdateMonitor)
	orgRoutes.DELETE("/monitors/:monitorId", h.MonitorHandler.DeleteMonitor)

	// Alert Rules
	orgRoutes.GET("/monitors/:monitorId/alert-rules", h.AlertRuleHandler.GetByMonitor)
	orgRoutes.POST("/monitors/:monitorId/alert-rules", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Create)
	orgRoutes.GET("/alert-rules/:ruleId", h.AlertRuleHandler.GetByID)
	orgRoutes.PUT("/alert-rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Update)
	orgRoutes.DELETE("/alert-rules/:ruleId", middleware.RequireRole("owner", "admin"), h.AlertRuleHandler.Delete)

	protected.GET("/monitors", h.MonitorHandler.GetAllMonitorByUserId)

	// Incident
	orgRoutes.GET("/incident/:incidentId", h.IncidentHandler.GetIncidentDetail)
	orgRoutes.GET("/incident", h.IncidentHandler.GetAllIncident)
	orgRoutes.POST("/incident/:incidentId", h.IncidentHandler.AddIncidentUpdate)
	orgRoutes.POST("/incident/:incidentId/join", h.IncidentHandler.JoinAssignee)
	orgRoutes.GET("/incident/:incidentId/updates", h.IncidentHandler.GetMessageAndUser)

	// StatusPage
	orgRoutes.POST("/statuspage", h.StatusPageHandler.CreateStatusPage)
	orgRoutes.GET("/statuspage", h.StatusPageHandler.GetAllStatusPage)
	protected.GET("/statuspages", h.StatusPageHandler.GetAllStatusPageByUserId)
	orgRoutes.GET("/statuspage/:statusPageId", h.StatusPageHandler.GetById)
	orgRoutes.GET("/statuspage/:statusPageId/monitors", h.StatusPageHandler.GetStatusPageMonitor)
	orgRoutes.DELETE("/statuspage/:statusPageId", h.StatusPageHandler.DeleteStatusPage)
	orgRoutes.PUT("/statuspage/:statusPageId", h.StatusPageHandler.UpdateStatusPage)
	orgRoutes.POST("/statuspage/:statusPageId/maintenance", h.StatusPageHandler.CreateMaintenance)
	orgRoutes.GET("/statuspage/:statusPageId/maintenance", h.StatusPageHandler.GetMaintenances)
	orgRoutes.PUT("/statuspage/:statusPageId/maintenance/:maintenanceId", h.StatusPageHandler.UpdateMaintenance)
	orgRoutes.DELETE("/statuspage/:statusPageId/maintenance/:maintenanceId", h.StatusPageHandler.DeleteMaintenance)
	orgRoutes.POST("/statuspage/:statusPageId/groups", h.StatusPageHandler.CreateComponentGroup)
	orgRoutes.GET("/statuspage/:statusPageId/groups", h.StatusPageHandler.GetComponentGroups)
	orgRoutes.PUT("/statuspage/:statusPageId/groups/:groupId", h.StatusPageHandler.UpdateComponentGroup)
	orgRoutes.DELETE("/statuspage/:statusPageId/groups/:groupId", h.StatusPageHandler.DeleteComponentGroup)
	orgRoutes.PUT("/statuspage/:statusPageId/monitors/:monitorId/group", h.StatusPageHandler.AssignMonitorGroup)

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
