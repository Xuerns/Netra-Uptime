package statuspage

import (
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

type StatusPageService interface {
	Create(orgID, userID, ipAddress string, req *CreateStatusPageRequest) (*model.StatusPage, error)
	GetAllByOrganizationID(orgID, userID string) ([]StatusPageResponse, error)
	Delete(statusPageID, orgID, UserID, ipAddress string) error
	Update(statuspageID, orgID, userID, ipAddress string, req *UpdateStatusPageRequest) (*model.StatusPage, error)
	GetById(statusPageID, orgID, userID string) (*StatusPageResponse, error)
	GetBySubdomain(subdomain string) (*PublicStatusPageResponse, error)
	GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error)
	GetStatusPageByUserId(userId string) ([]StatusPageResponse, error)
}

type SubscriptionChecker interface {
	GetOrganizationPlan(orgId string) (*model.Subscription, error)
	GetMemberById(orgId, userId string) (*model.OrganizationMember, error)
}

type StatusPageServiceImpl struct {
	repo         StatusPageRepository
	Subchecker   SubscriptionChecker
	rdb          *redis.Client
	auditService audit.AuditService
}

func NewStatusPageService(repo StatusPageRepository, checker SubscriptionChecker, rdb *redis.Client, auditService audit.AuditService) StatusPageService {
	return &StatusPageServiceImpl{repo: repo, Subchecker: checker, rdb: rdb, auditService: auditService}
}

func (s *StatusPageServiceImpl) Create(orgID, userID, ipAddress string, req *CreateStatusPageRequest) (*model.StatusPage, error) {
	totalStatuspage, err := s.repo.TotalOrganizationStatusPage(orgID)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	subs, err := s.Subchecker.GetOrganizationPlan(orgID)
	if err != nil {
		return nil, errors.New("Failed to get organization subscription")
	}

	maxStatusPage := 1
	switch subs.Plan {
	case "pro":
		maxStatusPage = 3
	case "max":
		maxStatusPage = 5
	}

	if totalStatuspage >= int64(maxStatusPage) {
		return nil, errors.New("You have reach max status page limit for your plan")
	}

	member, err := s.Subchecker.GetMemberById(orgID, userID)

	if err == nil && member == nil {
		return nil, errors.New("You're not member of this organization")
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

	statusPage := &model.StatusPage{
		OrganizationID:     orgUUID,
		Title:              req.Title,
		Description:        req.Description,
		Subdomain:          req.Subdomain,
		IsPublic:           req.IsPublic,
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
		monitors = append(monitors, StatusPageMonitorInfo{
			ID:           m.ID.String(),
			StatusPageID: m.StatusPageID.String(),
			MonitorID:    m.MonitorID.String(),
			DisplayOrder: m.DisplayOrder,
		})
	}

	resp := StatusPageResponse{
		ID:                 sp.ID.String(),
		OrganizationID:     sp.OrganizationID.String(),
		Subdomain:          sp.Subdomain,
		Title:              sp.Title,
		Description:        sp.Description,
		IsPublic:           sp.IsPublic,
		CreatedAt:          sp.CreatedAt,
		StatusPageMonitors: monitors,
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
	member, err := s.Subchecker.GetMemberById(orgID, userID)

	if err == nil && member == nil {
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
	member, err := s.Subchecker.GetMemberById(orgID, userID)

	if err == nil && member == nil {
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
	member, err := s.Subchecker.GetMemberById(orgID, userID)

	if err == nil && member == nil {
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

	if req.MonitorID != nil {
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

func (s *StatusPageServiceImpl) GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error) {
	statusPageMonitors, err := s.repo.GetStatusPageMonitor(statusPageID)
	if err != nil {
		return nil, errors.New("Something wrong, please try again")
	}

	return statusPageMonitors, nil
}

func (s *StatusPageServiceImpl) GetById(statusPageID, orgID, userID string) (*StatusPageResponse, error) {
	member, err := s.Subchecker.GetMemberById(orgID, userID)

	if err == nil && member == nil {
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

	var monitorsResponse []PublicMonitorResponse
	for _, spm := range statusPage.StatusPageMonitors {
		if spm.Monitor.IsActive {
			var recentChecks []int
			for _, check := range spm.Monitor.RecentChecks {
				recentChecks = append(recentChecks, check.ResponseTimeMs)
			}

			monitorsResponse = append(monitorsResponse, PublicMonitorResponse{
				ID:           spm.Monitor.ID.String(),
				Name:         spm.Monitor.Name,
				Type:         spm.Monitor.Type,
				Status:       spm.Monitor.Status,
				RecentChecks: recentChecks,
			})
		}
	}

	response := &PublicStatusPageResponse{
		Title:       statusPage.Title,
		Description: statusPage.Description,
		Subdomain:   statusPage.Subdomain,
		Monitors:    monitorsResponse,
	}

	return response, nil
}

func (s *StatusPageServiceImpl) GetStatusPageByUserId(userId string) ([]StatusPageResponse, error) {
	ctx := context.Background()
	cacheKey := "statuspages:user:" + userId

	cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cacheData != "" {
		var statuspages []StatusPageResponse

		if err := json.Unmarshal([]byte(cacheData), &statuspages); err == nil {
			return statuspages, nil
		}
	}

	statusPages, err := s.repo.GetStatusPageByUserId(userId)
	if err != nil {
		return nil, errors.New("Failed to get user status pages")
	}

	var response []StatusPageResponse
	for _, sp := range statusPages {
		response = append(response, toStatusPageResponse(sp))
	}

	if jsonData, err := json.Marshal(response); err == nil {
		s.rdb.Set(ctx, cacheKey, jsonData, 24*time.Hour)
	}

	return response, nil
}
