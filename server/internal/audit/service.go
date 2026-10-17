package audit

import (
	"errors"
	"math"
	"netra-uptime/internal/model"
	"time"

	"github.com/google/uuid"
)

type AuditService interface {
	Record(params RecordAuditLogParams) error
	GetLogs(orgId string, query GetAuditLogsQuery) (*PaginatedAuditLogsResponse, error)
}

type AuditServiceImpl struct {
	repo AuditRepository
}

func NewAuditService(repo AuditRepository) AuditService {
	return &AuditServiceImpl{repo: repo}
}

func (s *AuditServiceImpl) Record(params RecordAuditLogParams) error {
	orgUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return errors.New("invalid organization ID")
	}

	userUUID, err := uuid.Parse(params.UserID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	var entityUUID *uuid.UUID
	if params.EntityID != nil && *params.EntityID != "" {
		parsed, err := uuid.Parse(*params.EntityID)
		if err == nil {
			entityUUID = &parsed
		}
	}

	logEntry := model.AuditLog{
		OrganizationID: orgUUID,
		UserID:         userUUID,
		Action:         params.Action,
		EntityType:     params.EntityType,
		EntityID:       entityUUID,
		Description:    params.Description,
		Details:        params.Details,
		IPAddress:      params.IPAddress,
	}

	return s.repo.Create(&logEntry)
}

func (s *AuditServiceImpl) GetLogs(orgId string, query GetAuditLogsQuery) (*PaginatedAuditLogsResponse, error) {
	logs, total, err := s.repo.GetByOrganization(orgId, query)
	if err != nil {
		return nil, errors.New("failed to retrieve audit logs")
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	var logResponses []AuditLogResponse
	for _, l := range logs {
		var entityIDStr *string
		if l.EntityID != nil {
			str := l.EntityID.String()
			entityIDStr = &str
		}

		userName := ""
		userEmail := ""
		userAvatarURL := ""
		if l.User.ID != uuid.Nil {
			userName = l.User.Name
			userEmail = l.User.Email
			userAvatarURL = l.User.AvatarURL
		}

		logResponses = append(logResponses, AuditLogResponse{
			ID:             l.ID.String(),
			OrganizationID: l.OrganizationID.String(),
			UserID:         l.UserID.String(),
			UserName:       userName,
			UserEmail:      userEmail,
			UserAvatarURL:  userAvatarURL,
			Action:         l.Action,
			EntityType:     l.EntityType,
			EntityID:       entityIDStr,
			Description:    l.Description,
			Details:        l.Details,
			IPAddress:      l.IPAddress,
			CreatedAt:      l.CreatedAt.Format(time.RFC3339),
		})
	}

	return &PaginatedAuditLogsResponse{
		Logs:       logResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
