package statuspage

import (
	"errors"
	"fmt"
	"netra-uptime/config"
	"netra-uptime/internal/audit"
	"netra-uptime/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type StatusPageService interface {
	Create(orgID, userID, ipAddress string, req *CreateStatusPageRequest) (*model.StatusPage, error)
	GetAllByOrganizationID(orgID, userID string) ([]StatusPageResponse, error)
	Delete(statusPageID, orgID, UserID, ipAddress string) error
	Update(statuspageID, orgID, userID, ipAddress string, req *UpdateStatusPageRequest) (*model.StatusPage, error)
	GetById(statusPageID, orgID, userID string) (*StatusPageResponse, error)
	GetBySubdomain(subdomain string) (*PublicStatusPageResponse, error)
	GetStatusPageMonitor(statusPageID, orgID string) ([]model.StatusPageMonitor, error)
	CreateMaintenance(orgID, statusPageID, userID, ipAddress string, req *CreateMaintenanceRequest) (*model.ScheduledMaintenance, error)
	GetMaintenances(orgID, statusPageID, userID string) ([]MaintenanceResponse, error)
	UpdateMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string, req *UpdateMaintenanceRequest) (*model.ScheduledMaintenance, error)
	DeleteMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string) error
	CreateComponentGroup(orgID, statusPageID, userID, ipAddress string, req *CreateComponentGroupRequest) (*model.ComponentGroup, error)
	GetComponentGroups(orgID, statusPageID, userID string) ([]ComponentGroupResponse, error)
	UpdateComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string, req *UpdateComponentGroupRequest) (*model.ComponentGroup, error)
	DeleteComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string) error
	AssignMonitorToGroup(orgID, statusPageID, monitorID, userID, ipAddress string, req *AssignMonitorGroupRequest) error
}

type OrganizationMemberChecker interface {
	GetMemberById(orgId, userId string) (*model.OrganizationMember, error)
}

type StatusPageServiceImpl struct {
	repo          StatusPageRepository
	memberChecker OrganizationMemberChecker
	rdb           *redis.Client
	auditService  audit.AuditService
	cfg           *config.Config
}

func NewStatusPageService(repo StatusPageRepository, checker OrganizationMemberChecker, rdb *redis.Client, auditService audit.AuditService, cfg ...*config.Config) StatusPageService {
	var c *config.Config
	if len(cfg) > 0 && cfg[0] != nil {
		c = cfg[0]
	}
	return &StatusPageServiceImpl{repo: repo, memberChecker: checker, rdb: rdb, auditService: auditService, cfg: c}
}

func (s *StatusPageServiceImpl) getConfig() *config.Config {
	if s.cfg != nil {
		return s.cfg
	}
	return config.Load()
}

func (s *StatusPageServiceImpl) Create(orgID, userID, ipAddress string, req *CreateStatusPageRequest) (*model.StatusPage, error) {
	totalStatuspage, err := s.repo.TotalOrganizationStatusPage(orgID)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	maxStatusPage := 10
	cfg := s.getConfig()
	if cfg != nil && cfg.MaxStatusPagesPerOrg > 0 {
		maxStatusPage = cfg.MaxStatusPagesPerOrg
	}

	if totalStatuspage >= int64(maxStatusPage) {
		return nil, errors.New("Maximum status page limit reached for this organization")
	}

	member, err := s.memberChecker.GetMemberById(orgID, userID)

	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	if len(req.MonitorID) > 0 {
		if err := s.repo.ValidateMonitorsBelongToOrg(orgID, req.MonitorID); err != nil {
			return nil, errors.New("one or more monitors do not belong to this organization")
		}
	}

	var statusPageMonitor []model.StatusPageMonitor
	for _, monitorID := range req.MonitorID {

		muid, err := uuid.Parse(monitorID)

		if err != nil {
			return nil, errors.New("Invalid monitor ID")
		}

		statusPageMonitor = append(statusPageMonitor, model.StatusPageMonitor{
			MonitorID: muid,
		})
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("Invalid organization ID")
	}

	themeColor := req.ThemeColor
	if themeColor == "" {
		themeColor = "#4edea3"
	}

	statusPage := &model.StatusPage{
		OrganizationID:     orgUUID,
		Title:              req.Title,
		Description:        req.Description,
		Subdomain:          req.Subdomain,
		IsPublic:           req.IsPublic,
		LogoURL:            req.LogoURL,
		FaviconURL:         req.FaviconURL,
		ThemeColor:         themeColor,
		CustomCSS:          req.CustomCSS,
		StatusPageMonitors: statusPageMonitor,
	}

	createdStatusPage, err := s.repo.Create(statusPage)

	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	if s.auditService != nil {
		spIDStr := createdStatusPage.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "STATUSPAGE_CREATE",
			EntityType:     "status_page",
			EntityID:       &spIDStr,
			Description:    fmt.Sprintf("Created status page '%s' (%s)", createdStatusPage.Title, createdStatusPage.Subdomain),
			IPAddress:      ipAddress,
		})
	}

	return createdStatusPage, nil
}

func toStatusPageResponse(sp model.StatusPage) StatusPageResponse {
	var monitors []StatusPageMonitorInfo
	for _, m := range sp.StatusPageMonitors {
		var gIDStr *string
		if m.GroupID != nil {
			s := m.GroupID.String()
			gIDStr = &s
		}
		monitors = append(monitors, StatusPageMonitorInfo{
			ID:           m.ID.String(),
			StatusPageID: m.StatusPageID.String(),
			MonitorID:    m.MonitorID.String(),
			DisplayOrder: m.DisplayOrder,
			GroupID:      gIDStr,
		})
	}

	compGroups := make([]ComponentGroupResponse, 0)
	for _, cg := range sp.ComponentGroups {
		mIDs := make([]string, 0)
		for _, m := range sp.StatusPageMonitors {
			if m.GroupID != nil && *m.GroupID == cg.ID {
				mIDs = append(mIDs, m.MonitorID.String())
			}
		}
		compGroups = append(compGroups, ComponentGroupResponse{
			ID:           cg.ID.String(),
			StatusPageID: cg.StatusPageID.String(),
			Name:         cg.Name,
			DisplayOrder: cg.DisplayOrder,
			IsExpanded:   cg.IsExpanded,
			MonitorIDs:   mIDs,
		})
	}

	themeColor := sp.ThemeColor
	if themeColor == "" {
		themeColor = "#4edea3"
	}

	resp := StatusPageResponse{
		ID:                 sp.ID.String(),
		OrganizationID:     sp.OrganizationID.String(),
		Subdomain:          sp.Subdomain,
		Title:              sp.Title,
		Description:        sp.Description,
		IsPublic:           sp.IsPublic,
		LogoURL:            sp.LogoURL,
		FaviconURL:         sp.FaviconURL,
		ThemeColor:         themeColor,
		CustomCSS:          sp.CustomCSS,
		CreatedAt:          sp.CreatedAt,
		StatusPageMonitors: monitors,
		ComponentGroups:    compGroups,
	}

	if sp.Organization.ID != uuid.Nil {
		resp.Organization = &OrgInfo{
			ID:   sp.Organization.ID.String(),
			Name: sp.Organization.Name,
			Slug: sp.Organization.Slug,
		}
	}

	return resp
}

func (s *StatusPageServiceImpl) GetAllByOrganizationID(orgID, userID string) ([]StatusPageResponse, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)

	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	statusPages, err := s.repo.GetAllByOrganizationID(orgID)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	var response []StatusPageResponse
	for _, sp := range statusPages {
		response = append(response, toStatusPageResponse(sp))
	}

	return response, nil
}

func (s *StatusPageServiceImpl) Delete(statusPageID, orgID, userID, ipAddress string) error {
	member, err := s.memberChecker.GetMemberById(orgID, userID)

	if err != nil || member == nil {
		return errors.New("You're not member of this organization")
	}

	if err := s.repo.Delete(statusPageID, orgID); err != nil {
		return errors.New("Something wrong, please try again")
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "STATUSPAGE_DELETE",
			EntityType:     "status_page",
			EntityID:       &statusPageID,
			Description:    fmt.Sprintf("Deleted status page ID %s", statusPageID),
			IPAddress:      ipAddress,
		})
	}

	return nil
}

func (s *StatusPageServiceImpl) Update(statuspageID, orgID, userID, ipAddress string, req *UpdateStatusPageRequest) (*model.StatusPage, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)

	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	statusPage, err := s.repo.GetById(statuspageID, orgID)
	if err != nil {
		return nil, errors.New("Not Found")
	}

	if req.Title != nil {
		statusPage.Title = *req.Title
	}

	if req.Description != nil {
		statusPage.Description = req.Description
	}

	if req.Subdomain != nil {
		statusPage.Subdomain = *req.Subdomain
	}

	if req.IsPublic != nil {
		statusPage.IsPublic = *req.IsPublic
	}

	if req.LogoURL != nil {
		statusPage.LogoURL = *req.LogoURL
	}

	if req.FaviconURL != nil {
		statusPage.FaviconURL = *req.FaviconURL
	}

	if req.ThemeColor != nil {
		statusPage.ThemeColor = *req.ThemeColor
	}

	if req.CustomCSS != nil {
		statusPage.CustomCSS = *req.CustomCSS
	}

	if req.MonitorID != nil {
		if len(*req.MonitorID) > 0 {
			if err := s.repo.ValidateMonitorsBelongToOrg(orgID, *req.MonitorID); err != nil {
				return nil, errors.New("one or more monitors do not belong to this organization")
			}
		}

		var statusPageMonitor []model.StatusPageMonitor
		for _, monitorID := range *req.MonitorID {

			muid, err := uuid.Parse(monitorID)

			if err != nil {
				return nil, errors.New("Invalid monitor ID")
			}

			statusPageMonitor = append(statusPageMonitor, model.StatusPageMonitor{
				MonitorID: muid,
			})
		}
		statusPage.StatusPageMonitors = statusPageMonitor
	}

	updatedStatusPage, err := s.repo.Update(statuspageID, orgID, statusPage)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "STATUSPAGE_UPDATE",
			EntityType:     "status_page",
			EntityID:       &statuspageID,
			Description:    fmt.Sprintf("Updated status page '%s'", updatedStatusPage.Title),
			IPAddress:      ipAddress,
		})
	}

	return updatedStatusPage, nil
}

func (s *StatusPageServiceImpl) GetStatusPageMonitor(statusPageID, orgID string) ([]model.StatusPageMonitor, error) {
	statusPageMonitors, err := s.repo.GetStatusPageMonitor(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	return statusPageMonitors, nil
}

func (s *StatusPageServiceImpl) GetById(statusPageID, orgID, userID string) (*StatusPageResponse, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)

	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	statusPage, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Not Found")
	}

	resp := toStatusPageResponse(*statusPage)
	return &resp, nil
}

func (s *StatusPageServiceImpl) GetBySubdomain(subdomain string) (*PublicStatusPageResponse, error) {
	statusPage, err := s.repo.GetBySubdomain(subdomain)
	if err != nil {
		return nil, errors.New("Status page not found")
	}

	var monitorIDs []string
	for _, spm := range statusPage.StatusPageMonitors {
		if spm.Monitor.IsActive {
			monitorIDs = append(monitorIDs, spm.Monitor.ID.String())
		}
	}

	var uptimeMap map[string]float64
	var dailyStatsMap map[string][]model.MonitorDailyStat
	if len(monitorIDs) > 0 {
		uptimeMap, _ = s.repo.Get30DayUptimeMap(monitorIDs)
		dailyStatsMap, _ = s.repo.GetMonitorDailyStats(monitorIDs, 90)
	}

	monitorsResponse := make([]PublicMonitorResponse, 0)
	for _, spm := range statusPage.StatusPageMonitors {
		if spm.Monitor.IsActive {
			recentChecks := make([]int, 0)
			for _, check := range spm.Monitor.RecentChecks {
				recentChecks = append(recentChecks, check.ResponseTimeMs)
			}

			var uptimePct *float64
			if uptimeMap != nil {
				if val, ok := uptimeMap[spm.Monitor.ID.String()]; ok {
					uptimePct = &val
				}
			}

			var monitorDailyStats []model.MonitorDailyStat
			if dailyStatsMap != nil {
				monitorDailyStats = dailyStatsMap[spm.Monitor.ID.String()]
			}
			dailyStats := generate90DayStats(monitorDailyStats)

			var groupIDStr *string
			var groupNameStr *string
			if spm.GroupID != nil {
				s := spm.GroupID.String()
				groupIDStr = &s
				if spm.ComponentGroup != nil && spm.ComponentGroup.Name != "" {
					groupNameStr = &spm.ComponentGroup.Name
				}
			}

			monitorsResponse = append(monitorsResponse, PublicMonitorResponse{
				ID:            spm.Monitor.ID.String(),
				Name:          spm.Monitor.Name,
				Type:          spm.Monitor.Type,
				Status:        spm.Monitor.Status,
				GroupID:       groupIDStr,
				GroupName:     groupNameStr,
				UptimePercent: uptimePct,
				RecentChecks:  recentChecks,
				DailyStats:    dailyStats,
			})
		}
	}

	var activeIncidents []model.Incident
	var recentIncidents []model.Incident
	var historyIncidents []model.Incident
	if len(monitorIDs) > 0 {
		activeIncidents, _ = s.repo.GetActiveIncidentsByMonitorIDs(monitorIDs)
		recentIncidents, _ = s.repo.GetRecentResolvedIncidentsByMonitorIDs(monitorIDs, 7)
		historyIncidents, _ = s.repo.GetIncidentHistory(monitorIDs, 90)
	}

	ongoingMaintenances, upcomingMaintenances, _ := s.repo.GetActiveMaintenances(statusPage.ID.String())

	overallStatus := calculateOverallStatus(monitorsResponse, activeIncidents)

	componentGroupsResponse := make([]PublicComponentGroupResponse, 0)
	for _, cg := range statusPage.ComponentGroups {
		componentGroupsResponse = append(componentGroupsResponse, PublicComponentGroupResponse{
			ID:           cg.ID.String(),
			Name:         cg.Name,
			DisplayOrder: cg.DisplayOrder,
			IsExpanded:   cg.IsExpanded,
		})
	}

	incidentHistory := groupIncidentsByDay(historyIncidents, 0)

	themeColor := statusPage.ThemeColor
	if themeColor == "" {
		themeColor = "#4edea3"
	}

	response := &PublicStatusPageResponse{
		Title:                statusPage.Title,
		Description:          statusPage.Description,
		Subdomain:            statusPage.Subdomain,
		OverallStatus:        overallStatus,
		LogoURL:              statusPage.LogoURL,
		FaviconURL:           statusPage.FaviconURL,
		ThemeColor:           themeColor,
		CustomCSS:            statusPage.CustomCSS,
		Monitors:             monitorsResponse,
		ComponentGroups:      componentGroupsResponse,
		ActiveIncidents:      mapIncidentsToResponse(activeIncidents),
		RecentIncidents:      mapIncidentsToResponse(recentIncidents),
		IncidentHistory:      incidentHistory,
		UpcomingMaintenances: mapMaintenancesToPublicResponse(upcomingMaintenances),
		OngoingMaintenances:  mapMaintenancesToPublicResponse(ongoingMaintenances),
	}

	return response, nil
}

func calculateOverallStatus(monitors []PublicMonitorResponse, activeIncidents []model.Incident) string {
	if len(monitors) == 0 {
		return "operational"
	}

	downCount := 0
	degradedCount := 0
	for _, m := range monitors {
		switch m.Status {
		case "down":
			downCount++
		case "degraded":
			degradedCount++
		}
	}

	if downCount == len(monitors) {
		return "major_outage"
	}
	if downCount > 0 {
		return "partial_outage"
	}
	if degradedCount > 0 {
		return "degraded"
	}
	for _, inc := range activeIncidents {
		if inc.Severity == "critical" || inc.Severity == "high" {
			return "degraded"
		}
	}

	return "operational"
}

func mapIncidentsToResponse(incidents []model.Incident) []PublicIncidentResponse {
	resp := make([]PublicIncidentResponse, 0)
	for _, inc := range incidents {
		var resolvedAtStr *string
		if inc.ResolvedAt != nil {
			s := inc.ResolvedAt.Format(time.RFC3339)
			resolvedAtStr = &s
		}

		updates := make([]PublicIncidentUpdateResponse, 0)
		for _, u := range inc.IncidentUpdates {
			updates = append(updates, PublicIncidentUpdateResponse{
				Message:   u.Message,
				CreatedAt: u.CreatedAt.Format(time.RFC3339),
			})
		}

		monitorName := ""
		if inc.Monitor.Name != "" {
			monitorName = inc.Monitor.Name
		}

		resp = append(resp, PublicIncidentResponse{
			ID:          inc.ID.String(),
			Status:      inc.Status,
			Severity:    inc.Severity,
			Summary:     inc.Summary,
			StartedAt:   inc.StartedAt.Format(time.RFC3339),
			ResolvedAt:  resolvedAtStr,
			MonitorName: monitorName,
			Updates:     updates,
		})
	}
	return resp
}

func generate90DayStats(stats []model.MonitorDailyStat) []PublicUptimeDayResponse {
	statMap := make(map[string]model.MonitorDailyStat)
	for _, s := range stats {
		statMap[s.Date.Format("2006-01-02")] = s
	}

	result := make([]PublicUptimeDayResponse, 90)
	now := time.Now()

	for i := 0; i < 90; i++ {
		day := now.AddDate(0, 0, -(89 - i))
		dateStr := day.Format("2006-01-02")
		if s, ok := statMap[dateStr]; ok && s.TotalChecks > 0 {
			result[i] = PublicUptimeDayResponse{
				Date:             dateStr,
				UptimePercentage: s.UptimePercentage,
				DownChecks:       s.DownChecks,
				TotalChecks:      s.TotalChecks,
				HasData:          true,
			}
		} else {
			result[i] = PublicUptimeDayResponse{
				Date:             dateStr,
				UptimePercentage: 100.0,
				DownChecks:       0,
				TotalChecks:      0,
				HasData:          false,
			}
		}
	}
	return result
}

func mapMaintenancesToPublicResponse(maintenances []model.ScheduledMaintenance) []PublicMaintenanceResponse {
	resp := make([]PublicMaintenanceResponse, 0)
	for _, m := range maintenances {
		monNames := make([]string, 0)
		for _, am := range m.AffectedMonitors {
			if am.Monitor.Name != "" {
				monNames = append(monNames, am.Monitor.Name)
			}
		}
		resp = append(resp, PublicMaintenanceResponse{
			ID:               m.ID.String(),
			Title:            m.Title,
			Description:      m.Description,
			Status:           m.Status,
			ScheduledStart:   m.ScheduledStart.Format(time.RFC3339),
			ScheduledEnd:     m.ScheduledEnd.Format(time.RFC3339),
			AffectedMonitors: monNames,
		})
	}
	return resp
}

func (s *StatusPageServiceImpl) CreateMaintenance(orgID, statusPageID, userID, ipAddress string, req *CreateMaintenanceRequest) (*model.ScheduledMaintenance, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	sp, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Status page not found")
	}

	if len(req.MonitorIDs) > 0 {
		if err := s.repo.ValidateMonitorsBelongToOrg(orgID, req.MonitorIDs); err != nil {
			return nil, errors.New("one or more monitors do not belong to this organization")
		}
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("Invalid organization ID")
	}

	maintenance := &model.ScheduledMaintenance{
		OrganizationID: orgUUID,
		StatusPageID:   sp.ID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         "scheduled",
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   req.ScheduledEnd,
	}

	created, err := s.repo.CreateMaintenance(maintenance, req.MonitorIDs)
	if err != nil {
		return nil, errors.New("Failed to create scheduled maintenance")
	}

	if s.auditService != nil {
		mIDStr := created.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "MAINTENANCE_CREATE",
			EntityType:     "scheduled_maintenance",
			EntityID:       &mIDStr,
			Description:    fmt.Sprintf("Created scheduled maintenance '%s'", created.Title),
			IPAddress:      ipAddress,
		})
	}

	return created, nil
}

func (s *StatusPageServiceImpl) GetMaintenances(orgID, statusPageID, userID string) ([]MaintenanceResponse, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	maintenances, err := s.repo.GetMaintenancesByStatusPageID(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Failed to retrieve maintenances")
	}

	resp := make([]MaintenanceResponse, 0)
	for _, m := range maintenances {
		monNames := make([]string, 0)
		monIDs := make([]string, 0)
		for _, am := range m.AffectedMonitors {
			if am.Monitor.Name != "" {
				monNames = append(monNames, am.Monitor.Name)
			}
			monIDs = append(monIDs, am.MonitorID.String())
		}
		resp = append(resp, MaintenanceResponse{
			ID:                 m.ID.String(),
			StatusPageID:       m.StatusPageID.String(),
			Title:              m.Title,
			Description:        m.Description,
			Status:             m.Status,
			ScheduledStart:     m.ScheduledStart,
			ScheduledEnd:       m.ScheduledEnd,
			ActualStart:        m.ActualStart,
			ActualEnd:          m.ActualEnd,
			CreatedAt:          m.CreatedAt,
			AffectedMonitors:   monNames,
			AffectedMonitorIDs: monIDs,
		})
	}

	return resp, nil
}

func (s *StatusPageServiceImpl) UpdateMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string, req *UpdateMaintenanceRequest) (*model.ScheduledMaintenance, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	existing, err := s.repo.GetMaintenanceByID(maintenanceID, orgID)
	if err != nil {
		return nil, errors.New("Maintenance not found")
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		existing.Status = *req.Status
		if *req.Status == "in_progress" && existing.ActualStart == nil {
			now := time.Now()
			existing.ActualStart = &now
		} else if *req.Status == "completed" && existing.ActualEnd == nil {
			now := time.Now()
			existing.ActualEnd = &now
		}
	}
	if req.ScheduledStart != nil {
		existing.ScheduledStart = *req.ScheduledStart
	}
	if req.ScheduledEnd != nil {
		existing.ScheduledEnd = *req.ScheduledEnd
	}

	if req.MonitorIDs != nil && len(*req.MonitorIDs) > 0 {
		if err := s.repo.ValidateMonitorsBelongToOrg(orgID, *req.MonitorIDs); err != nil {
			return nil, errors.New("one or more monitors do not belong to this organization")
		}
	}

	updated, err := s.repo.UpdateMaintenance(existing, req.MonitorIDs)
	if err != nil {
		return nil, errors.New("Failed to update scheduled maintenance")
	}

	if s.auditService != nil {
		mIDStr := updated.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "MAINTENANCE_UPDATE",
			EntityType:     "scheduled_maintenance",
			EntityID:       &mIDStr,
			Description:    fmt.Sprintf("Updated scheduled maintenance '%s'", updated.Title),
			IPAddress:      ipAddress,
		})
	}

	return updated, nil
}

func (s *StatusPageServiceImpl) DeleteMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string) error {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return errors.New("You're not member of this organization")
	}

	if err := s.repo.DeleteMaintenance(maintenanceID, orgID); err != nil {
		return errors.New("Failed to delete scheduled maintenance")
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "MAINTENANCE_DELETE",
			EntityType:     "scheduled_maintenance",
			EntityID:       &maintenanceID,
			Description:    fmt.Sprintf("Deleted scheduled maintenance ID %s", maintenanceID),
			IPAddress:      ipAddress,
		})
	}

	return nil
}

func groupIncidentsByDay(incidents []model.Incident, continuousDays int) []PublicIncidentDayGroup {
	incMap := make(map[string][]model.Incident)
	for _, inc := range incidents {
		dateStr := inc.StartedAt.Format("2006-01-02")
		incMap[dateStr] = append(incMap[dateStr], inc)
	}

	result := make([]PublicIncidentDayGroup, 0)
	now := time.Now()

	seenDates := make(map[string]bool)
	for i := 0; i < continuousDays; i++ {
		day := now.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		seenDates[dateStr] = true

		dayIncidents := make([]PublicIncidentResponse, 0)
		if incs, ok := incMap[dateStr]; ok {
			dayIncidents = mapIncidentsToResponse(incs)
		}
		result = append(result, PublicIncidentDayGroup{
			Date:      dateStr,
			Incidents: dayIncidents,
		})
	}

	for _, inc := range incidents {
		dateStr := inc.StartedAt.Format("2006-01-02")
		if !seenDates[dateStr] {
			seenDates[dateStr] = true
			result = append(result, PublicIncidentDayGroup{
				Date:      dateStr,
				Incidents: mapIncidentsToResponse(incMap[dateStr]),
			})
		}
	}

	return result
}

func (s *StatusPageServiceImpl) CreateComponentGroup(orgID, statusPageID, userID, ipAddress string, req *CreateComponentGroupRequest) (*model.ComponentGroup, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	sp, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Status page not found")
	}

	isExpanded := true
	if req.IsExpanded != nil {
		isExpanded = *req.IsExpanded
	}

	group := &model.ComponentGroup{
		StatusPageID: sp.ID,
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
		IsExpanded:   isExpanded,
	}

	created, err := s.repo.CreateComponentGroup(group)
	if err != nil {
		return nil, errors.New("Failed to create component group")
	}

	if s.auditService != nil {
		gIDStr := created.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "COMPONENT_GROUP_CREATE",
			EntityType:     "component_group",
			EntityID:       &gIDStr,
			Description:    fmt.Sprintf("Created component group '%s'", created.Name),
			IPAddress:      ipAddress,
		})
	}

	return created, nil
}

func (s *StatusPageServiceImpl) GetComponentGroups(orgID, statusPageID, userID string) ([]ComponentGroupResponse, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	sp, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Status page not found")
	}

	groups, err := s.repo.GetComponentGroupsByStatusPageID(statusPageID)
	if err != nil {
		return nil, errors.New("Failed to retrieve component groups")
	}

	resp := make([]ComponentGroupResponse, 0)
	for _, g := range groups {
		mIDs := make([]string, 0)
		for _, spm := range sp.StatusPageMonitors {
			if spm.GroupID != nil && *spm.GroupID == g.ID {
				mIDs = append(mIDs, spm.MonitorID.String())
			}
		}
		resp = append(resp, ComponentGroupResponse{
			ID:           g.ID.String(),
			StatusPageID: g.StatusPageID.String(),
			Name:         g.Name,
			DisplayOrder: g.DisplayOrder,
			IsExpanded:   g.IsExpanded,
			MonitorIDs:   mIDs,
		})
	}

	return resp, nil
}

func (s *StatusPageServiceImpl) UpdateComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string, req *UpdateComponentGroupRequest) (*model.ComponentGroup, error) {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return nil, errors.New("You're not member of this organization")
	}

	sp, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return nil, errors.New("Status page not found")
	}
	_ = sp

	group, err := s.repo.GetComponentGroupByID(groupID, statusPageID)
	if err != nil {
		return nil, errors.New("Component group not found")
	}

	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.DisplayOrder != nil {
		group.DisplayOrder = *req.DisplayOrder
	}
	if req.IsExpanded != nil {
		group.IsExpanded = *req.IsExpanded
	}

	updated, err := s.repo.UpdateComponentGroup(group)
	if err != nil {
		return nil, errors.New("Failed to update component group")
	}

	if s.auditService != nil {
		gIDStr := updated.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "COMPONENT_GROUP_UPDATE",
			EntityType:     "component_group",
			EntityID:       &gIDStr,
			Description:    fmt.Sprintf("Updated component group '%s'", updated.Name),
			IPAddress:      ipAddress,
		})
	}

	return updated, nil
}

func (s *StatusPageServiceImpl) DeleteComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string) error {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return errors.New("You're not member of this organization")
	}

	sp, err := s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return errors.New("Status page not found")
	}
	_ = sp

	group, err := s.repo.GetComponentGroupByID(groupID, statusPageID)
	if err != nil {
		return errors.New("Component group not found")
	}

	if err := s.repo.DeleteComponentGroup(groupID, statusPageID); err != nil {
		return errors.New("Failed to delete component group")
	}

	if s.auditService != nil {
		gIDStr := group.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "COMPONENT_GROUP_DELETE",
			EntityType:     "component_group",
			EntityID:       &gIDStr,
			Description:    fmt.Sprintf("Deleted component group '%s'", group.Name),
			IPAddress:      ipAddress,
		})
	}

	return nil
}

func (s *StatusPageServiceImpl) AssignMonitorToGroup(orgID, statusPageID, monitorID, userID, ipAddress string, req *AssignMonitorGroupRequest) error {
	member, err := s.memberChecker.GetMemberById(orgID, userID)
	if err != nil || member == nil {
		return errors.New("You're not member of this organization")
	}

	_, err = s.repo.GetById(statusPageID, orgID)
	if err != nil {
		return errors.New("Status page not found")
	}

	if err := s.repo.ValidateMonitorsBelongToOrg(orgID, []string{monitorID}); err != nil {
		return errors.New("one or more monitors do not belong to this organization")
	}

	if req.GroupID != nil && *req.GroupID != "" {
		_, err := s.repo.GetComponentGroupByID(*req.GroupID, statusPageID)
		if err != nil {
			return errors.New("Component group not found on this status page")
		}
	}

	if err := s.repo.AssignMonitorToGroup(statusPageID, monitorID, req.GroupID); err != nil {
		return errors.New("Failed to assign monitor to group")
	}

	return nil
}
