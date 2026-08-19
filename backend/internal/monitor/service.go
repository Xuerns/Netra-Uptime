package monitor

import (
	"PulseWatch/config"
	"PulseWatch/internal/audit"
	"PulseWatch/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	GetAllMonitorByUserId(userId string) ([]MonitorResponse, error)
}

type SubscriptionChecker interface {
	GetOrganizationPlan(orgId string) (*model.Subscription, error)
}

type MonitorServiceImpl struct {
	repo         MonitorRepository
	rdb          *redis.Client
	subsChecker  SubscriptionChecker
	auditService audit.AuditService
}

func NewMonitorService(repo MonitorRepository, rdb *redis.Client, checker SubscriptionChecker, auditService audit.AuditService) MonitorService {
	return &MonitorServiceImpl{repo: repo, rdb: rdb, subsChecker: checker, auditService: auditService}
}

func (s *MonitorServiceImpl) Create(orgId, userId, ipAddress string, req *CreateMonitorRequest) (*model.Monitor, error) {
	parceUUID, err := uuid.Parse(orgId)

	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	subs, err := s.subsChecker.GetOrganizationPlan(orgId)
	if err != nil {
		return nil, errors.New("Failed to get organization subscription")
	}

	maxMonitors := 5
	maxPIC := 2
	maxInterval := 300
	maxTimeout := 10

	cfg := config.Load()
	if cfg.AppMode == "selfhosted" {
		if cfg.SelfHostedMaxMonitors > 0 {
			maxMonitors = cfg.SelfHostedMaxMonitors
		} else {
			maxMonitors = 999999
		}
		maxPIC = cfg.SelfHostedMaxPIC
		maxInterval = cfg.SelfHostedMinInterval
		maxTimeout = cfg.SelfHostedMaxTimeout
	} else {
		switch subs.Plan {
		case "pro":
			maxMonitors = 10
			maxPIC = 4
			maxInterval = 180
			maxTimeout = 15
		case "max":
			maxMonitors = 20
			maxPIC = 6
			maxInterval = 60
			maxTimeout = 20
		}
	}

	totalMonitorHave, err := s.repo.TotalOrganizationMonitor(orgId)
	if err != nil {
		return nil, err
	}

	if totalMonitorHave >= int64(maxMonitors) {
		return nil, errors.New("You have reach max monitor limit for your plan")
	}

	if len(req.MonitorPICs) > maxPIC {
		return nil, errors.New("You have reach max PIC limit for this monitor")
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

	if len(req.MonitorPICs) > 0 {

		if !s.repo.ValidateUsersInOrg(orgId, req.MonitorPICs) {
			return nil, errors.New("Some of PIC is not member of this organization")
		}

		for _, pic := range req.MonitorPICs {
			parceUUID, err := uuid.Parse(pic)
			if err != nil {
				return nil, errors.New("invalid PIC UUID format")
			}

			monitor.MonitorPICs = append(monitor.MonitorPICs, model.MonitorPIC{
				UserID: parceUUID,
			})
		}
	}

	if err := s.repo.Create(&monitor); err != nil {
		return nil, errors.New("Failed to create new monitor")
	}

	s.rdb.Del(context.Background(), "monitors:org:"+orgId)

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
	resp := MonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Type:                 m.Type,
		Target:               m.Target,
		Status:               m.Status,
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
		resp.MonitorPICs = append(resp.MonitorPICs, MonitorPICInfo{
			UserID: pic.UserID.String(),
		})
	}

	for _, rc := range m.RecentChecks {
		lastCheckedAt := rc.CheckedAt.Format(time.RFC3339)
		resp.LastCheckedAt = &lastCheckedAt
		resp.ResponseTimeMs = &rc.ResponseTimeMs

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
			CheckedAt:      lastCheckedAt,
		})
	}
	return resp
}

func (s *MonitorServiceImpl) GetById(orgId, monitorId string) (*MonitorResponse, error) {
	ctx := context.Background()
	cacheKey := "monitor:" + monitorId

	cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cacheData != "" {
		var monitorResponse MonitorResponse

		if err := json.Unmarshal([]byte(cacheData), &monitorResponse); err == nil {
			return &monitorResponse, nil
		}
	}

	monitor, err := s.repo.GetById(orgId, monitorId)

	if err != nil {
		return nil, errors.New("Failed to get Monitor")
	}

	response := mapMonitorToResponse(*monitor)

	if jsonData, err := json.Marshal(response); err == nil {
		s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Second)
	}

	return &response, nil
}

func (s *MonitorServiceImpl) GetAllMonitor(orgId string) ([]MonitorResponse, error) {
	ctx := context.Background()
	cacheKey := "monitors:org:" + orgId

	cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cacheData != "" {
		var monitorsResponse []MonitorResponse

		if err := json.Unmarshal([]byte(cacheData), &monitorsResponse); err == nil {
			return monitorsResponse, nil
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

	if jsonData, err := json.Marshal(response); err == nil {
		s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Second)
	}

	return response, nil
}

func (s *MonitorServiceImpl) UpdateMonitor(orgId, userId, ipAddress, monitorId string, req UpdatedMonitorRequest) (*model.Monitor, error) {
	subs, err := s.subsChecker.GetOrganizationPlan(orgId)
	if err != nil {
		return nil, errors.New("Failed to get organization subscription")
	}

	maxPIC := 2
	maxInterval := 300
	maxTimeout := 10

	cfg := config.Load()
	if cfg.AppMode == "selfhosted" {
		maxPIC = cfg.SelfHostedMaxPIC
		maxInterval = cfg.SelfHostedMinInterval
		maxTimeout = cfg.SelfHostedMaxTimeout
	} else {
		switch subs.Plan {
		case "pro":
			maxPIC = 4
			maxInterval = 180
			maxTimeout = 15
		case "max":
			maxPIC = 6
			maxInterval = 60
			maxTimeout = 20
		}
	}

	updatedMonitor, err := s.repo.GetById(orgId, monitorId)

	if err != nil {
		return nil, errors.New("Failed to get monitor")
	}

	if req.Name != nil {
		updatedMonitor.Name = *req.Name
	}

	if req.Target != nil {
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
	}

	if req.Criticality != nil {
		updatedMonitor.Criticality = *req.Criticality
	}

	if req.EnableSSLDomainCheck != nil {
		updatedMonitor.EnableSSLDomainCheck = *req.EnableSSLDomainCheck
	}

	if err := s.repo.Update(updatedMonitor); err != nil {
		return nil, errors.New("Failed to update monitor")
	}

	if req.MonitorPICs != nil {
		if len(*req.MonitorPICs) > maxPIC {
			return nil, errors.New("You have reach max PIC limit for this monitor")
		}

		if !s.repo.ValidateUsersInOrg(orgId, *req.MonitorPICs) {
			return nil, errors.New("Some of PIC is not member of this organization")
		}

		var picIds []model.MonitorPIC
		for _, id := range *req.MonitorPICs {
			parceUUID, err := uuid.Parse(id)

			if err != nil {
				return nil, errors.New("Some of PIC is invalid")
			}
			picIds = append(picIds, model.MonitorPIC{
				UserID: parceUUID,
			})
		}

		if err := s.repo.ReplaceMonitorPICs(updatedMonitor, picIds); err != nil {
			return nil, errors.New("Failed to replace PIC")
		}
	}

	s.rdb.Del(context.Background(), "monitors:org:"+orgId)
	s.rdb.Del(context.Background(), "monitor:"+monitorId)

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

	s.rdb.Del(context.Background(), "monitors:org:"+orgId)
	s.rdb.Del(context.Background(), "monitor:"+monitorId)

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

func (s *MonitorServiceImpl) GetAllMonitorByUserId(userId string) ([]MonitorResponse, error) {
	ctx := context.Background()
	cacheKey := "monitors:user:" + userId

	cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cacheData != "" {
		var monitorsResponse []MonitorResponse

		if err := json.Unmarshal([]byte(cacheData), &monitorsResponse); err == nil {
			return monitorsResponse, nil
		}
	}

	monitors, err := s.repo.GetAllMonitorByUserId(userId)
	if err != nil {
		return nil, errors.New("Failed to get user monitors")
	}

	var response []MonitorResponse
	for _, m := range monitors {
		response = append(response, mapMonitorToResponse(m))
	}

	if jsonData, err := json.Marshal(response); err == nil {
		s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Second)
	}

	return response, nil
}
