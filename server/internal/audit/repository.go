package audit

import (
	"strings"

	"netra-uptime/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(log *model.AuditLog) error
	GetByOrganization(orgId string, query GetAuditLogsQuery) ([]model.AuditLog, int64, error)
}

type AuditRepositoryImpl struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &AuditRepositoryImpl{db: db}
}

func (r *AuditRepositoryImpl) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepositoryImpl) GetByOrganization(orgId string, query GetAuditLogsQuery) ([]model.AuditLog, int64, error) {
	orgUUID, err := uuid.Parse(orgId)
	if err != nil {
		return nil, 0, err
	}

	dbQuery := r.db.Model(&model.AuditLog{}).
		Preload("User").
		Where("organization_id = ?", orgUUID)

	if query.Action != "" {
		if strings.Contains(query.Action, ",") {
			rawActions := strings.Split(query.Action, ",")
			var actions []string
			for _, a := range rawActions {
				if trimmed := strings.TrimSpace(a); trimmed != "" {
					actions = append(actions, trimmed)
				}
			}
			if len(actions) > 0 {
				dbQuery = dbQuery.Where("action IN ?", actions)
			}
		} else {
			dbQuery = dbQuery.Where("action = ?", query.Action)
		}
	}

	if query.EntityType != "" {
		if strings.Contains(query.EntityType, ",") {
			rawTypes := strings.Split(query.EntityType, ",")
			var types []string
			for _, t := range rawTypes {
				if trimmed := strings.TrimSpace(t); trimmed != "" {
					types = append(types, trimmed)
				}
			}
			if len(types) > 0 {
				dbQuery = dbQuery.Where("entity_type IN ?", types)
			}
		} else {
			dbQuery = dbQuery.Where("entity_type = ?", query.EntityType)
		}
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("LOWER(description) LIKE LOWER(?) OR LOWER(action) LIKE LOWER(?)", searchPattern, searchPattern)
	}

	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var logs []model.AuditLog
	err = dbQuery.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, nil
}
