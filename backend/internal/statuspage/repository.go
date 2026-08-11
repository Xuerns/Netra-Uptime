package statuspage

import (
	"PulseWatch/config"
	"PulseWatch/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatusPageRepository interface {
	Create(statuspage *model.StatusPage) (*model.StatusPage, error)
	GetAllByOrganizationID(orgID string) ([]model.StatusPage, error)
	Delete(statusPageID, orgID string) error
	Update(statuspageID, orgID string, statuspage *model.StatusPage) (*model.StatusPage, error)
	GetById(statuspageID, orgID string) (*model.StatusPage, error)
	GetBySubdomain(subdomain string) (*model.StatusPage, error)
	GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error)
	GetStatusPageByUserId(userId string) ([]model.StatusPage, error)
	TotalOrganizationStatusPage(orgId string) (int64, error)
}

type StatusPageRepositoryImpl struct {
	db *gorm.DB
}

func NewStatusPageRepository(db *gorm.DB) StatusPageRepository {
	return &StatusPageRepositoryImpl{db: db}
}

func (r *StatusPageRepositoryImpl) Create(statuspage *model.StatusPage) (*model.StatusPage, error) {

	if err := r.db.Create(&statuspage).Error; err != nil {
		return nil, err
	}

	return statuspage, nil
}

func (r *StatusPageRepositoryImpl) GetAllByOrganizationID(orgID string) ([]model.StatusPage, error) {
	var StatusPages []model.StatusPage

	if err := r.db.Where("organization_id = ?", orgID).Find(&StatusPages).Error; err != nil {
		return nil, err
	}

	return StatusPages, nil
}

func (r *StatusPageRepositoryImpl) Delete(statusPageID, orgID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var statuspage model.StatusPage
		if err := tx.Where("id = ? AND organization_id = ?", statusPageID, orgID).First(&statuspage).Error; err != nil {
			return err
		}

		if err := tx.Where("status_page_id = ?", statusPageID).Delete(&model.StatusPageMonitor{}).Error; err != nil {
			return err
		}

		return tx.Delete(&statuspage).Error
	})
}

func (r *StatusPageRepositoryImpl) Update(statuspageID, orgID string, statuspage *model.StatusPage) (*model.StatusPage, error) {
	var updatedStatusPage model.StatusPage

	err := r.db.Model(&updatedStatusPage).Clauses(clause.Returning{}).Where("id = ? AND organization_id = ?", statuspageID, orgID).Updates(statuspage).Error

	if err != nil {
		return nil, err
	}

	if statuspage.StatusPageMonitors != nil {
		if err := r.db.Where("status_page_id = ?", statuspageID).Delete(&model.StatusPageMonitor{}).Error; err != nil {
			return nil, err
		}

		if len(statuspage.StatusPageMonitors) > 0 {
			if err := r.db.Model(&updatedStatusPage).Association("StatusPageMonitors").Append(statuspage.StatusPageMonitors); err != nil {
				return nil, err
			}
		}
	}

	return &updatedStatusPage, nil
}

func (r *StatusPageRepositoryImpl) GetById(statuspageID, orgID string) (*model.StatusPage, error) {
	var statuspage model.StatusPage
	if err := r.db.Where("id = ? AND organization_id = ?", statuspageID, orgID).First(&statuspage).Error; err != nil {
		return nil, err
	}
	return &statuspage, nil
}

func (r *StatusPageRepositoryImpl) GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error) {
	var statusPageMonitors []model.StatusPageMonitor
	if err := r.db.Preload("Monitor").Where("status_page_id = ?", statusPageID).Order("display_order").Find(&statusPageMonitors).Error; err != nil {
		return nil, err
	}
	return statusPageMonitors, nil
}

func (r *StatusPageRepositoryImpl) GetBySubdomain(subdomain string) (*model.StatusPage, error) {
	var statuspage model.StatusPage
	if err := r.db.Preload("StatusPageMonitors").Preload("StatusPageMonitors.Monitor").Where("subdomain = ? AND is_public = ?", subdomain, true).First(&statuspage).Error; err != nil {
		return nil, err
	}

	var monitorIDs []string
	for _, spMonitor := range statuspage.StatusPageMonitors {
		monitorIDs = append(monitorIDs, spMonitor.MonitorID.String())
	}

	if len(monitorIDs) > 0 {
		limit := config.Load().RecentChecksLimit
		var recentChecks []model.Check
		r.db.Raw(`
			SELECT * FROM (
				SELECT *, ROW_NUMBER() OVER(PARTITION BY monitor_id ORDER BY checked_at DESC) as rn
				FROM checks
				WHERE monitor_id IN ?
			) as subquery
			WHERE rn <= ?
		`, monitorIDs, limit).Scan(&recentChecks)

		checkMap := make(map[string][]model.Check)
		for _, check := range recentChecks {
			monitorIDStr := check.MonitorID.String()
			checkMap[monitorIDStr] = append(checkMap[monitorIDStr], check)
		}

		for i, spMonitor := range statuspage.StatusPageMonitors {
			statuspage.StatusPageMonitors[i].Monitor.RecentChecks = checkMap[spMonitor.MonitorID.String()]
		}
	}

	return &statuspage, nil
}

func (r *StatusPageRepositoryImpl) GetStatusPageByUserId(userId string) ([]model.StatusPage, error) {
	var statuspages []model.StatusPage
	err := r.db.Preload("Organization").
		Preload("StatusPageMonitors").
		Joins("JOIN organization_members ON status_pages.organization_id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userId).
		Find(&statuspages).Error

	if err != nil {
		return nil, err
	}

	return statuspages, nil
}

func (r *StatusPageRepositoryImpl) TotalOrganizationStatusPage(orgId string) (int64, error) {
	var totalStatusPage int64
	if err := r.db.Model(&model.StatusPage{}).Where("organization_id = ?", orgId).Count(&totalStatusPage).Error; err != nil {
		return 0, err
	}

	return totalStatusPage, nil
}