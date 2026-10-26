package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"netra-uptime/config"
	"netra-uptime/internal/audit"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MonitorService interface {
	Create(orgId, userId, ipAddress string, req *CreateMonitorRequest) (*model.Monitor, error)
	GetById(orgId, monitorId string) (*MonitorResponse, error)
	GetAllMonitor(orgId string) ([]MonitorResponse, error)
	UpdateMonitor(orgId, userId, ipAddress, monitorId string, req UpdatedMonitorRequest) (*model.Monitor, error)
	DeleteMonitor(orgId, userId, ipAddress, monitorId string) error
	GetDailyStats(orgId, monitorId string, days int) ([]model.MonitorDailyStat, error)
	TriggerCheck(orgId, monitorId string) (*RecentCheck, error)
}

type SubscriptionChecker interface {
	GetOrganizationPlan(orgId string) (*model.Subscription, error)
}

type ImmediateChecker interface {
	ProcessImmediate(monitorID uuid.UUID) (*model.Check, error)
}

type MonitorServiceImpl struct {
	repo             MonitorRepository
	rdb              *redis.Client
	subsChecker      SubscriptionChecker
	auditService     audit.AuditService
	immediateChecker ImmediateChecker
	cfg              *config.Config
}

func NewMonitorService(repo MonitorRepository, rdb *redis.Client, checker SubscriptionChecker, auditService audit.AuditService, opts ...interface{}) MonitorService {
	var ic ImmediateChecker
	var cfg *config.Config
	for _, opt := range opts {
		switch v := opt.(type) {
		case ImmediateChecker:
			ic = v
		case *config.Config:
			cfg = v
		}
	}
	return &MonitorServiceImpl{
		repo:             repo,
		rdb:              rdb,
		subsChecker:      checker,
		auditService:     auditService,
		immediateChecker: ic,
		cfg:              cfg,
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

	subs, err := s.subsChecker.GetOrganizationPlan(orgId)
	if err != nil {
		return nil, errors.New("Failed to get organization subscription")
	}

	maxMonitors := 5
	maxAssignees := 5
	maxInterval := 300
	maxTimeout := 10

	cfg := s.getConfig()
	if cfg.AppMode == "selfhosted" {
		if cfg.SelfHostedMaxMonitors > 0 {
			maxMonitors = cfg.SelfHostedMaxMonitors
		} else {
			maxMonitors = 999999
		}
		if cfg.SelfHostedMaxPIC > 0 {
			maxAssignees = cfg.SelfHostedMaxPIC
		}
		maxInterval = cfg.SelfHostedMinInterval
		if maxInterval < 10 {
			maxInterval = 10
		}
		maxTimeout = cfg.SelfHostedMaxTimeout
	} else if subs.Plan != "free" && subs.Plan != "hobby" {
		maxMonitors = 20
		maxAssignees = 10
		maxInterval = 60
		maxTimeout = 20
	}

	totalMonitorHave, err := s.repo.TotalOrganizationMonitor(orgId)
	if err != nil {
		return nil, err
	}

	if totalMonitorHave >= int64(maxMonitors) {
		return nil, errors.New("You have reach max monitor limit for your plan")
	}

	assignees := req.Assignees
	if len(assignees) == 0 && len(req.MonitorPICs) > 0 {
		assignees = req.MonitorPICs
	}

	if len(assignees) > maxAssignees {
		return nil, errors.New("You have reach max assignee limit for this monitor")
	}

	if req.IntervalSeconds < maxInterval {
		return nil, errors.New("You have reach max interval limit for your plan")
	}

	if req.TimeoutSeconds > maxTimeout {
		return nil, errors.New("You have reach max timeout limit for your plan")
	}

	if req.TimeoutSeconds < 10 {
		return nil, errors.New("Invalid timeout seconds")
	}

	enableCheck := true
	if req.EnableSSLDomainCheck != nil {
		enableCheck = *req.EnableSSLDomainCheck
	}

	var monitor = model.Monitor{
		Name:                 req.Name,
		Type:                 req.Type,
		Target:               req.Target,
		IntervalSeconds:      req.IntervalSeconds,
		TimeoutSeconds:       req.TimeoutSeconds,
		ExpectedStatusCode:   req.ExpectedStatusCode,
		Criticality:          req.Criticality,
		OrganizationID:       parceUUID,
		EnableSSLDomainCheck: enableCheck,
	}

	if req.AlertTemplateID != nil && *req.AlertTemplateID != "" {
		if parsed, err := uuid.Parse(*req.AlertTemplateID); err == nil {
			monitor.AlertTemplateID = &parsed
		}
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

	if s.rdb != nil {
		s.rdb.Del(context.Background(), "monitors:org:"+orgId)
	}

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
		SSLDaysLeft:          m.SSLDaysLeft,
		SSLStatus:            m.SSLStatus,
		SSLError:             m.SSLError,
		DomainDaysLeft:       m.DomainDaysLeft,
		DomainStatus:         m.DomainStatus,
		DomainError:          m.DomainError,
	}

	if m.SSLExpiresAt != nil {
		tStr := m.SSLExpiresAt.Format(time.RFC3339)
		resp.SSLExpiresAt = &tStr
	}
	if m.DomainExpiresAt != nil {
		tStr := m.DomainExpiresAt.Format(time.RFC3339)
		resp.DomainExpiresAt = &tStr
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

func (s *MonitorServiceImpl) GetById(orgId, monitorId string) (*MonitorResponse, error) {
	ctx := context.Background()
	cacheKey := "monitor:org:" + orgId + ":" + monitorId

	if s.rdb != nil {
		cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil && cacheData != "" {
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

	if s.rdb != nil {
		if jsonData, err := json.Marshal(response); err == nil {
			s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Second)
		}
	}

	return &response, nil
}

func (s *MonitorServiceImpl) GetAllMonitor(orgId string) ([]MonitorResponse, error) {
	ctx := context.Background()
	cacheKey := "monitors:org:" + orgId

	if s.rdb != nil {
		cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil && cacheData != "" {
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

	if s.rdb != nil {
		if jsonData, err := json.Marshal(response); err == nil {
			s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Second)
		}
	}

	return response, nil
}

func (s *MonitorServiceImpl) UpdateMonitor(orgId, userId, ipAddress, monitorId string, req UpdatedMonitorRequest) (*model.Monitor, error) {
	subs, err := s.subsChecker.GetOrganizationPlan(orgId)
	if err != nil {
		return nil, errors.New("Failed to get organization subscription")
	}

	maxAssignees := 5
	maxInterval := 300
	maxTimeout := 10

	cfg := s.getConfig()
	if cfg.AppMode == "selfhosted" {
		if cfg.SelfHostedMaxPIC > 0 {
			maxAssignees = cfg.SelfHostedMaxPIC
		}
		maxInterval = cfg.SelfHostedMinInterval
		maxTimeout = cfg.SelfHostedMaxTimeout
	} else if subs.Plan != "free" && subs.Plan != "hobby" {
		maxAssignees = 10
		maxInterval = 60
		maxTimeout = 20
	}

	updatedMonitor, err := s.repo.GetById(orgId, monitorId)

	if err != nil {
		return nil, errors.New("Failed to get monitor")
	}

	if req.Name != nil {
		updatedMonitor.Name = *req.Name
	}

	if req.Target != nil {
		if err := utils.ValidateTargetHost(*req.Target); err != nil {
			return nil, fmt.Errorf("Invalid target: %w", err)
		}
		updatedMonitor.Target = *req.Target
	}

	if req.IntervalSeconds != nil {
		if *req.IntervalSeconds < maxInterval {
			return nil, errors.New("You have reach max interval limit for your plan")
		}
		updatedMonitor.IntervalSeconds = *req.IntervalSeconds
	}

	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds > maxTimeout {
			return nil, errors.New("You have reach max timeout limit for your plan")
		}
		if *req.TimeoutSeconds < 1 {
			return nil, errors.New("Invalid timeout seconds")
		}
		updatedMonitor.TimeoutSeconds = *req.TimeoutSeconds
	}

	if req.ExpectedStatusCode != nil {
		updatedMonitor.ExpectedStatusCode = *req.ExpectedStatusCode
	}

	if req.IsActive != nil {
		updatedMonitor.IsActive = *req.IsActive
		if !*req.IsActive {
			updatedMonitor.Status = "paused"
		} else if updatedMonitor.Status == "paused" {
			updatedMonitor.Status = "up"
		}
	}

	if req.Criticality != nil {
		updatedMonitor.Criticality = *req.Criticality
	}

	if req.EnableSSLDomainCheck != nil {
		updatedMonitor.EnableSSLDomainCheck = *req.EnableSSLDomainCheck
	}

	if req.AlertTemplateID != nil {
		if *req.AlertTemplateID == "" {
			updatedMonitor.AlertTemplateID = nil
		} else if parsed, err := uuid.Parse(*req.AlertTemplateID); err == nil {
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

	if s.rdb != nil {
		s.rdb.Del(context.Background(), "monitors:org:"+orgId)
		s.rdb.Del(context.Background(), "monitor:org:"+orgId+":"+monitorId)
	}

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

	if s.rdb != nil {
		s.rdb.Del(context.Background(), "monitors:org:"+orgId)
		s.rdb.Del(context.Background(), "monitor:org:"+orgId+":"+monitorId)
	}

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

	if s.immediateChecker == nil {
		return nil, errors.New("Checker engine is not available")
	}

	check, err := s.immediateChecker.ProcessImmediate(mon.ID)
	if err != nil {
		return nil, err
	}

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
