package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"netra-uptime/config"
	"netra-uptime/internal/audit"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/cache"
	"netra-uptime/pkg/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type MonitorService interface {
	Create(orgId, userId, ipAddress string, req *CreateMonitorRequest) (*model.Monitor, error)
	GetById(orgId, monitorId string) (*MonitorResponse, error)
	GetAllMonitor(orgId string) ([]MonitorResponse, error)
	UpdateMonitor(orgId, userId, ipAddress, monitorId string, req UpdatedMonitorRequest) (*model.Monitor, error)
	DeleteMonitor(orgId, userId, ipAddress, monitorId string) error
	GetDailyStats(orgId, monitorId string, days int) ([]model.MonitorDailyStat, error)
	TriggerCheck(orgId, monitorId string) (*RecentCheck, error)
	GetFleetOverview(orgId string, windowDays int) (*FleetOverviewResponse, error)
}

type ImmediateChecker interface {
	ProcessImmediate(monitorID uuid.UUID) (*model.Check, error)
}

type MonitorServiceImpl struct {
	repo             MonitorRepository
	cache            *cache.MemoryCache
	auditService     audit.AuditService
	immediateChecker ImmediateChecker
	cfg              *config.Config
	db               *gorm.DB
	sfGroup          singleflight.Group
}

func NewMonitorService(repo MonitorRepository, cacheStore *cache.MemoryCache, auditService audit.AuditService, opts ...interface{}) MonitorService {
	var ic ImmediateChecker
	var cfg *config.Config
	var db *gorm.DB
	for _, opt := range opts {
		switch v := opt.(type) {
		case ImmediateChecker:
			ic = v
		case *config.Config:
			cfg = v
		case *gorm.DB:
			db = v
		}
	}
	if db == nil && ic != nil {
		if dbGetter, ok := ic.(interface{ GetDB() *gorm.DB }); ok {
			db = dbGetter.GetDB()
		}
	}
	if db == nil && repo != nil {
		if dbGetter, ok := repo.(interface{ GetDB() *gorm.DB }); ok {
			db = dbGetter.GetDB()
		}
	}
	return &MonitorServiceImpl{
		repo:             repo,
		cache:            cacheStore,
		auditService:     auditService,
		immediateChecker: ic,
		cfg:              cfg,
		db:               db,
	}
}

func (s *MonitorServiceImpl) getConfig() *config.Config {
	if s.cfg != nil {
		return s.cfg
	}
	return config.Load()
}

func (s *MonitorServiceImpl) Create(orgId, userId, ipAddress string, req *CreateMonitorRequest) (*model.Monitor, error) {
	parceUUID, err := uuid.Parse(orgId)

	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	if err := utils.ValidateTargetHost(req.Target); err != nil {
		return nil, fmt.Errorf("Invalid target: %w", err)
	}

	maxMonitors := 200
	maxAssignees := 50
	minInterval := 30
	maxTimeout := 30

	cfg := s.getConfig()
	if cfg != nil {
		if cfg.MaxMonitorsPerOrg > 0 {
			maxMonitors = cfg.MaxMonitorsPerOrg
		} else if cfg.SelfHostedMaxMonitors > 0 {
			maxMonitors = cfg.SelfHostedMaxMonitors
		}
		if cfg.MaxAssigneesPerMonitor > 0 {
			maxAssignees = cfg.MaxAssigneesPerMonitor
		} else if cfg.SelfHostedMaxPIC > 0 {
			maxAssignees = cfg.SelfHostedMaxPIC
		}
		if cfg.SelfHostedMinInterval > 0 {
			minInterval = cfg.SelfHostedMinInterval
		}
		if cfg.SelfHostedMaxTimeout > 0 {
			maxTimeout = cfg.SelfHostedMaxTimeout
		}
	}
	if minInterval < 30 {
		minInterval = 30
	}
	if maxTimeout > 30 {
		maxTimeout = 30
	}

	totalMonitorHave, err := s.repo.TotalOrganizationMonitor(orgId)
	if err != nil {
		return nil, err
	}

	if totalMonitorHave >= int64(maxMonitors) {
		return nil, errors.New("Maximum monitor limit reached for this organization")
	}

	assignees := req.Assignees
	if len(assignees) == 0 && len(req.MonitorPICs) > 0 {
		assignees = req.MonitorPICs
	}

	if len(assignees) > maxAssignees {
		return nil, errors.New("You have reach max assignee limit for this monitor")
	}

	if req.IntervalSeconds < 30 || req.IntervalSeconds < minInterval {
		return nil, errors.New("Check interval is below the minimum allowed frequency (minimum 30 seconds)")
	}

	if req.TimeoutSeconds < 1 {
		return nil, errors.New("Invalid timeout seconds")
	}

	if req.TimeoutSeconds > 30 || req.TimeoutSeconds > maxTimeout {
		return nil, errors.New("Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)")
	}

	if req.TimeoutSeconds >= req.IntervalSeconds {
		return nil, errors.New("monitor timeout must be strictly less than check interval to prevent overlapping checks")
	}

	rootDomain, err := utils.ExtractRootDomain(req.Target)
	if err != nil {
		return nil, fmt.Errorf("Invalid target domain: %w", err)
	}

	if s.db != nil {
		var count int64
		if err := s.db.Model(&model.Monitor{}).Where("organization_id = ? AND is_active = true AND target_domain = ?", parceUUID, rootDomain).Count(&count).Error; err != nil {
			return nil, err
		}
		if count >= 5 {
			return nil, fmt.Errorf("domain quota exceeded: maximum of 5 active monitors allowed for domain %s in this organization", rootDomain)
		}
	}

	enableCheck := true
	if req.EnableSSLDomainCheck != nil {
		enableCheck = *req.EnableSSLDomainCheck
	}

	var monitor = model.Monitor{
		Name:                 req.Name,
		Type:                 req.Type,
		Target:               req.Target,
		TargetDomain:         rootDomain,
		IntervalSeconds:      req.IntervalSeconds,
		TimeoutSeconds:       req.TimeoutSeconds,
		ExpectedStatusCode:   req.ExpectedStatusCode,
		Criticality:          req.Criticality,
		OrganizationID:       parceUUID,
		EnableSSLDomainCheck: enableCheck,
	}

	if req.AlertTemplateID != nil && *req.AlertTemplateID != "" {
		parsed, err := uuid.Parse(*req.AlertTemplateID)
		if err != nil {
			return nil, errors.New("invalid alert template UUID format")
		}
		if s.db != nil {
			var count int64
			if err := s.db.Model(&model.AlertRuleTemplate{}).Where("id = ? AND organization_id = ?", parsed, parceUUID).Count(&count).Error; err != nil {
				return nil, err
			}
			if count == 0 {
				return nil, errors.New("alert template not found or access denied")
			}
		}
		monitor.AlertTemplateID = &parsed
	}

	if len(assignees) > 0 {
		if !s.repo.ValidateUsersInOrg(orgId, assignees) {
			return nil, errors.New("Some of assignees are not member of this organization")
		}

		for _, a := range assignees {
			parsedUUID, err := uuid.Parse(a)
			if err != nil {
				return nil, errors.New("invalid assignee UUID format")
			}

			monitor.MonitorPICs = append(monitor.MonitorPICs, model.MonitorPIC{
				UserID: parsedUUID,
			})
		}
	}

	if err := s.repo.Create(&monitor); err != nil {
		return nil, errors.New("Failed to create new monitor")
	}

	s.invalidateMonitorCache(orgId, "")

	if s.auditService != nil {
		monitorIDStr := monitor.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         userId,
			Action:         "MONITOR_CREATE",
			EntityType:     "monitor",
			EntityID:       &monitorIDStr,
			Description:    fmt.Sprintf("Created monitor '%s' (%s)", monitor.Name, monitor.Type),
			IPAddress:      ipAddress,
		})
	}

	return &monitor, nil
}

func mapMonitorToResponse(m model.Monitor) MonitorResponse {
	status := m.Status
	if !m.IsActive {
		status = "paused"
	}

	resp := MonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Type:                 m.Type,
		Target:               m.Target,
		Status:               status,
		IntervalSeconds:      m.IntervalSeconds,
		TimeoutSeconds:       m.TimeoutSeconds,
		ExpectedStatusCode:   m.ExpectedStatusCode,
		Criticality:          m.Criticality,
		IsActive:             m.IsActive,
		OrganizationID:       m.OrganizationID.String(),
		EnableSSLDomainCheck: m.EnableSSLDomainCheck,
		SSLStatus:            "none",
		DomainStatus:         "none",
	}

	if m.Type == "http" && m.EnableSSLDomainCheck {
		resp.SSLDaysLeft = m.SSLDaysLeft
		resp.SSLStatus = m.SSLStatus
		resp.SSLError = m.SSLError
		resp.DomainDaysLeft = m.DomainDaysLeft
		resp.DomainStatus = m.DomainStatus
		resp.DomainError = m.DomainError

		if m.SSLExpiresAt != nil {
			tStr := m.SSLExpiresAt.Format(time.RFC3339)
			resp.SSLExpiresAt = &tStr
		}
		if m.DomainExpiresAt != nil {
			tStr := m.DomainExpiresAt.Format(time.RFC3339)
			resp.DomainExpiresAt = &tStr
		}
	}

	if m.Organization.ID != uuid.Nil {
		resp.Organization = &OrgInfo{
			ID:   m.Organization.ID.String(),
			Name: m.Organization.Name,
			Slug: m.Organization.Slug,
		}
	}

	for _, pic := range m.MonitorPICs {
		picInfo := MonitorPICInfo{
			UserID: pic.UserID.String(),
		}
		resp.MonitorPICs = append(resp.MonitorPICs, picInfo)
		resp.Assignees = append(resp.Assignees, picInfo)
	}

	if m.CalculatedUptime30D != nil {
		resp.UptimePercent = m.CalculatedUptime30D
	} else if len(m.RecentChecks) > 0 {
		upCount := 0
		for _, rc := range m.RecentChecks {
			if rc.Status == "up" {
				upCount++
			}
		}
		pct := (float64(upCount) / float64(len(m.RecentChecks))) * 100.0
		resp.UptimePercent = &pct
	} else {
		pct := 100.0
		resp.UptimePercent = &pct
	}

	if len(m.RecentChecks) > 0 {
		latest := m.RecentChecks[0]
		lastCheckedAt := latest.CheckedAt.Format(time.RFC3339)
		resp.LastCheckedAt = &lastCheckedAt
		resp.ResponseTimeMs = &latest.ResponseTimeMs
	}

	for _, rc := range m.RecentChecks {
		checkedAt := rc.CheckedAt.Format(time.RFC3339)
		var errMsg *string
		if rc.ErrorMessage != "" {
			errMsg = &rc.ErrorMessage
		}

		resp.RecentChecks = append(resp.RecentChecks, RecentCheck{
			ID:             rc.ID.String(),
			MonitorID:      rc.MonitorID.String(),
			Status:         rc.Status,
			StatusCode:     rc.StatusCode,
			ResponseTimeMs: rc.ResponseTimeMs,
			ErrorMessage:   errMsg,
			CheckedAt:      checkedAt,
		})
	}

	if m.AlertTemplateID != nil {
		tmplIDStr := m.AlertTemplateID.String()
		resp.AlertTemplateID = &tmplIDStr
	}

	return resp
}

func (s *MonitorServiceImpl) invalidateMonitorCache(orgId, monitorId string) {
	if s.cache == nil {
		return
	}
	if monitorId != "" {
		s.cache.Del(cache.MonitorDetailKey(orgId, monitorId), cache.MonitorListKey(orgId), cache.MonitorOverviewKey(orgId, 30))
	} else {
		s.cache.Del(cache.MonitorListKey(orgId), cache.MonitorOverviewKey(orgId, 30))
	}
}

func (s *MonitorServiceImpl) GetById(orgId, monitorId string) (*MonitorResponse, error) {
	cacheKey := cache.MonitorDetailKey(orgId, monitorId)

	if s.cache != nil {
		if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
			var monitorResponse MonitorResponse

			if err := json.Unmarshal([]byte(cacheData), &monitorResponse); err == nil {
				if monitorResponse.OrganizationID == orgId {
					return &monitorResponse, nil
				}
				return nil, errors.New("Failed to get Monitor")
			}
		}
	}

	res, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		if s.cache != nil {
			if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
				var monitorResponse MonitorResponse
				if err := json.Unmarshal([]byte(cacheData), &monitorResponse); err == nil {
					if monitorResponse.OrganizationID == orgId {
						return &monitorResponse, nil
					}
					return nil, errors.New("Failed to get Monitor")
				}
			}
		}

		monitor, err := s.repo.GetById(orgId, monitorId)
		if err != nil {
			return nil, errors.New("Failed to get Monitor")
		}

		response := mapMonitorToResponse(*monitor)

		if s.cache != nil {
			if jsonData, err := json.Marshal(response); err == nil {
				s.cache.Set(cacheKey, string(jsonData), 15*time.Second)
			}
		}

		return &response, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*MonitorResponse), nil
}

func (s *MonitorServiceImpl) GetAllMonitor(orgId string) ([]MonitorResponse, error) {
	cacheKey := cache.MonitorListKey(orgId)

	if s.cache != nil {
		if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
			var monitorsResponse []MonitorResponse

			if err := json.Unmarshal([]byte(cacheData), &monitorsResponse); err == nil {
				return monitorsResponse, nil
			}
		}
	}

	res, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		if s.cache != nil {
			if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
				var monitorsResponse []MonitorResponse
				if err := json.Unmarshal([]byte(cacheData), &monitorsResponse); err == nil {
					return monitorsResponse, nil
				}
			}
		}

		monitors, err := s.repo.GetAllByOrg(orgId)
		if err != nil {
			return nil, errors.New("Failed to get All monitor in organization")
		}

		var response []MonitorResponse
		for _, m := range monitors {
			response = append(response, mapMonitorToResponse(m))
		}

		if s.cache != nil {
			if jsonData, err := json.Marshal(response); err == nil {
				s.cache.Set(cacheKey, string(jsonData), 15*time.Second)
			}
		}

		return response, nil
	})

	if err != nil {
		return nil, err
	}

	return res.([]MonitorResponse), nil
}

func (s *MonitorServiceImpl) UpdateMonitor(orgId, userId, ipAddress, monitorId string, req UpdatedMonitorRequest) (*model.Monitor, error) {
	parceOrgUUID, err := uuid.Parse(orgId)
	if err != nil {
		return nil, errors.New("invalid organization UUID format")
	}
	parceUUID, err := uuid.Parse(monitorId)
	if err != nil {
		return nil, errors.New("invalid monitor UUID format")
	}

	maxAssignees := 50
	minInterval := 30
	maxTimeout := 30

	cfg := s.getConfig()
	if cfg != nil {
		if cfg.MaxAssigneesPerMonitor > 0 {
			maxAssignees = cfg.MaxAssigneesPerMonitor
		} else if cfg.SelfHostedMaxPIC > 0 {
			maxAssignees = cfg.SelfHostedMaxPIC
		}
		if cfg.SelfHostedMinInterval > 0 {
			minInterval = cfg.SelfHostedMinInterval
		}
		if cfg.SelfHostedMaxTimeout > 0 {
			maxTimeout = cfg.SelfHostedMaxTimeout
		}
	}
	if minInterval < 30 {
		minInterval = 30
	}
	if maxTimeout > 30 {
		maxTimeout = 30
	}

	updatedMonitor, err := s.repo.GetById(orgId, monitorId)

	if err != nil {
		return nil, errors.New("Failed to get monitor")
	}

	effectiveInterval := updatedMonitor.IntervalSeconds
	if req.IntervalSeconds != nil {
		effectiveInterval = *req.IntervalSeconds
	}
	effectiveTimeout := updatedMonitor.TimeoutSeconds
	if req.TimeoutSeconds != nil {
		effectiveTimeout = *req.TimeoutSeconds
	}

	if req.IntervalSeconds != nil {
		if *req.IntervalSeconds < 30 || *req.IntervalSeconds < minInterval {
			return nil, errors.New("Check interval is below the minimum allowed frequency (minimum 30 seconds)")
		}
		updatedMonitor.IntervalSeconds = *req.IntervalSeconds
	}

	if effectiveInterval < 30 || effectiveInterval < minInterval {
		return nil, errors.New("Check interval is below the minimum allowed frequency (minimum 30 seconds)")
	}

	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds < 1 {
			return nil, errors.New("Invalid timeout seconds")
		}
		if *req.TimeoutSeconds > 30 || *req.TimeoutSeconds > maxTimeout {
			return nil, errors.New("Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)")
		}
		updatedMonitor.TimeoutSeconds = *req.TimeoutSeconds
	}

	if effectiveTimeout < 1 {
		return nil, errors.New("Invalid timeout seconds")
	}

	if effectiveTimeout > 30 || effectiveTimeout > maxTimeout {
		return nil, errors.New("Request timeout exceeds the maximum allowed timeout (maximum 30 seconds)")
	}

	if effectiveTimeout >= effectiveInterval {
		return nil, errors.New("monitor timeout must be strictly less than check interval to prevent overlapping checks")
	}

	if req.Name != nil {
		updatedMonitor.Name = *req.Name
	}

	if req.Target != nil {
		if err := utils.ValidateTargetHost(*req.Target); err != nil {
			return nil, fmt.Errorf("Invalid target: %w", err)
		}
		updatedMonitor.Target = *req.Target
		rootDomain, err := utils.ExtractRootDomain(*req.Target)
		if err != nil {
			return nil, fmt.Errorf("Invalid target domain: %w", err)
		}
		updatedMonitor.TargetDomain = rootDomain
	} else if updatedMonitor.TargetDomain == "" && updatedMonitor.Target != "" {
		if rootDomain, err := utils.ExtractRootDomain(updatedMonitor.Target); err == nil {
			updatedMonitor.TargetDomain = rootDomain
		}
	}

	if req.ExpectedStatusCode != nil {
		updatedMonitor.ExpectedStatusCode = *req.ExpectedStatusCode
	}

	isActive := updatedMonitor.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
		updatedMonitor.IsActive = isActive
		if !isActive {
			updatedMonitor.Status = "paused"
		} else if updatedMonitor.Status == "paused" {
			updatedMonitor.Status = "up"
		}
	}

	if s.db != nil && isActive && updatedMonitor.TargetDomain != "" {
		var count int64
		if err := s.db.Model(&model.Monitor{}).Where("organization_id = ? AND is_active = true AND target_domain = ? AND id != ?", parceOrgUUID, updatedMonitor.TargetDomain, parceUUID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count >= 5 {
			return nil, fmt.Errorf("domain quota exceeded: maximum of 5 active monitors allowed for domain %s in this organization", updatedMonitor.TargetDomain)
		}
	}

	if req.Criticality != nil {
		updatedMonitor.Criticality = *req.Criticality
	}

	if req.EnableSSLDomainCheck != nil {
		updatedMonitor.EnableSSLDomainCheck = *req.EnableSSLDomainCheck
		if !*req.EnableSSLDomainCheck {
			updatedMonitor.SSLStatus = "none"
			updatedMonitor.DomainStatus = "none"
			updatedMonitor.SSLDaysLeft = nil
			updatedMonitor.DomainDaysLeft = nil
			updatedMonitor.SSLExpiresAt = nil
			updatedMonitor.DomainExpiresAt = nil
			updatedMonitor.SSLError = ""
			updatedMonitor.DomainError = ""
			updatedMonitor.LastSSLCheckedAt = nil
			updatedMonitor.LastDomainCheckedAt = nil
			updatedMonitor.LastSSLSentReminderDay = nil
			updatedMonitor.LastDomainSentReminderDay = nil
		}
	}

	if req.AlertTemplateID != nil {
		if *req.AlertTemplateID == "" {
			updatedMonitor.AlertTemplateID = nil
		} else {
			parsed, err := uuid.Parse(*req.AlertTemplateID)
			if err != nil {
				return nil, errors.New("invalid alert template UUID format")
			}
			if s.db != nil {
				var count int64
				if err := s.db.Model(&model.AlertRuleTemplate{}).Where("id = ? AND organization_id = ?", parsed, parceOrgUUID).Count(&count).Error; err != nil {
					return nil, err
				}
				if count == 0 {
					return nil, errors.New("alert template not found or access denied")
				}
			}
			updatedMonitor.AlertTemplateID = &parsed
		}
	}

	if err := s.repo.Update(updatedMonitor); err != nil {
		return nil, errors.New("Failed to update monitor")
	}

	assignees := req.Assignees
	if assignees == nil && req.MonitorPICs != nil {
		assignees = req.MonitorPICs
	}

	if assignees != nil {
		if len(*assignees) > maxAssignees {
			return nil, errors.New("You have reach max assignee limit for this monitor")
		}

		if len(*assignees) > 0 && !s.repo.ValidateUsersInOrg(orgId, *assignees) {
			return nil, errors.New("Some of assignees are not member of this organization")
		}

		var picIds []model.MonitorPIC
		for _, id := range *assignees {
			parsedUUID, err := uuid.Parse(id)

			if err != nil {
				return nil, errors.New("Some of assignee UUID is invalid")
			}
			picIds = append(picIds, model.MonitorPIC{
				UserID: parsedUUID,
			})
		}

		if err := s.repo.ReplaceMonitorPICs(updatedMonitor, picIds); err != nil {
			return nil, errors.New("Failed to replace assignees")
		}
	}

	s.invalidateMonitorCache(orgId, monitorId)

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         userId,
			Action:         "MONITOR_UPDATE",
			EntityType:     "monitor",
			EntityID:       &monitorId,
			Description:    fmt.Sprintf("Updated monitor '%s'", updatedMonitor.Name),
			IPAddress:      ipAddress,
		})
	}

	return updatedMonitor, nil
}

func (s *MonitorServiceImpl) DeleteMonitor(orgId, userId, ipAddress, monitorId string) error {
	if err := s.repo.Delete(orgId, monitorId); err != nil {
		return errors.New("Failed to delete monitor")
	}

	s.invalidateMonitorCache(orgId, monitorId)

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgId,
			UserID:         userId,
			Action:         "MONITOR_DELETE",
			EntityType:     "monitor",
			EntityID:       &monitorId,
			Description:    fmt.Sprintf("Deleted monitor ID %s", monitorId),
			IPAddress:      ipAddress,
		})
	}

	return nil
}

func (s *MonitorServiceImpl) GetDailyStats(orgId, monitorId string, days int) ([]model.MonitorDailyStat, error) {
	_, err := s.repo.GetById(orgId, monitorId)
	if err != nil {
		return nil, errors.New("Monitor not found or access denied")
	}

	if days <= 0 {
		days = 30
	}

	return s.repo.GetDailyStats(monitorId, days)
}

func (s *MonitorServiceImpl) TriggerCheck(orgId, monitorId string) (*RecentCheck, error) {
	mon, err := s.repo.GetById(orgId, monitorId)
	if err != nil {
		return nil, errors.New("Monitor not found or access denied")
	}

	if !mon.IsActive {
		return nil, errors.New("cannot trigger check on inactive monitor")
	}

	cooldownKey := fmt.Sprintf("check_cooldown:monitor:%s", monitorId)
	if s.cache != nil {
		if _, ok := s.cache.Get(cooldownKey); ok {
			return nil, errors.New("check cooldown active, please wait 10 seconds before triggering again")
		}
		s.cache.Set(cooldownKey, "1", 10*time.Second)
	}

	if s.immediateChecker == nil {
		return nil, errors.New("Checker engine is not available")
	}

	check, err := s.immediateChecker.ProcessImmediate(mon.ID)
	if err != nil {
		return nil, err
	}

	s.invalidateMonitorCache(orgId, monitorId)

	var errMsg *string
	if check.ErrorMessage != "" {
		errMsg = &check.ErrorMessage
	}

	return &RecentCheck{
		ID:             check.ID.String(),
		MonitorID:      check.MonitorID.String(),
		Status:         check.Status,
		StatusCode:     check.StatusCode,
		ResponseTimeMs: check.ResponseTimeMs,
		ErrorMessage:   errMsg,
		CheckedAt:      check.CheckedAt.Format(time.RFC3339),
	}, nil
}

func (s *MonitorServiceImpl) GetFleetOverview(orgId string, windowDays int) (*FleetOverviewResponse, error) {
	if windowDays <= 0 {
		windowDays = 30
	}

	cacheKey := cache.MonitorOverviewKey(orgId, windowDays)

	// Check cache first
	if s.cache != nil {
		if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
			var cachedOverview FleetOverviewResponse
			if err := json.Unmarshal([]byte(cacheData), &cachedOverview); err == nil {
				return &cachedOverview, nil
			}
		}
	}

	// Singleflight
	res, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		// Re-check cache inside singleflight
		if s.cache != nil {
			if cacheData, ok := s.cache.Get(cacheKey); ok && cacheData != "" {
				var cachedOverview FleetOverviewResponse
				if err := json.Unmarshal([]byte(cacheData), &cachedOverview); err == nil {
					return &cachedOverview, nil
				}
			}
		}

		overview, err := s.repo.GetFleetOverview(orgId, windowDays)
		if err != nil {
			return nil, err
		}

		// Calculate trend_delta
		if overview.Uptime.CurrentPercent != nil && overview.Uptime.PreviousPercent != nil {
			delta := math.Round((*overview.Uptime.CurrentPercent-*overview.Uptime.PreviousPercent)*100) / 100
			overview.Uptime.TrendDelta = &delta
		}

		if s.cache != nil {
			if jsonData, err := json.Marshal(overview); err == nil {
				s.cache.Set(cacheKey, string(jsonData), 30*time.Second)
			}
		}

		return overview, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*FleetOverviewResponse), nil
}
