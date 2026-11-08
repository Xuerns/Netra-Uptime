package statuspage

import (
	"errors"
	"netra-uptime/config"
	"netra-uptime/internal/model"
	"time"

	"github.com/google/uuid"
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
	GetStatusPageMonitor(statusPageID, orgID string) ([]model.StatusPageMonitor, error)
	TotalOrganizationStatusPage(orgId string) (int64, error)
	GetActiveIncidentsByMonitorIDs(monitorIDs []string) ([]model.Incident, error)
	GetRecentResolvedIncidentsByMonitorIDs(monitorIDs []string, days int) ([]model.Incident, error)
	Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error)
	GetMonitorDailyStats(monitorIDs []string, days int) (map[string][]model.MonitorDailyStat, error)
	GetActiveMaintenances(statusPageID string) ([]model.ScheduledMaintenance, []model.ScheduledMaintenance, error)
	CreateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs []string) (*model.ScheduledMaintenance, error)
	GetMaintenancesByStatusPageID(statusPageID, orgID string) ([]model.ScheduledMaintenance, error)
	GetMaintenanceByID(maintenanceID, orgID string) (*model.ScheduledMaintenance, error)
	UpdateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs *[]string) (*model.ScheduledMaintenance, error)
	DeleteMaintenance(maintenanceID, orgID string) error
	GetIncidentHistory(monitorIDs []string, days int) ([]model.Incident, error)
	CreateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error)
	GetComponentGroupsByStatusPageID(statusPageID string) ([]model.ComponentGroup, error)
	GetComponentGroupByID(groupID, statusPageID string) (*model.ComponentGroup, error)
	UpdateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error)
	DeleteComponentGroup(groupID, statusPageID string) error
	AssignMonitorToGroup(statusPageID, monitorID string, groupID *string) error
	ValidateMonitorsBelongToOrg(orgID string, monitorIDs []string) error
}

type StatusPageRepositoryImpl struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewStatusPageRepository(db *gorm.DB, cfg ...*config.Config) StatusPageRepository {
	var c *config.Config
	if len(cfg) > 0 && cfg[0] != nil {
		c = cfg[0]
	}
	return &StatusPageRepositoryImpl{db: db, cfg: c}
}

func (r *StatusPageRepositoryImpl) getRecentChecksLimit() int {
	if r.cfg != nil && r.cfg.RecentChecksLimit > 0 {
		return r.cfg.RecentChecksLimit
	}
	return config.Load().RecentChecksLimit
}

func (r *StatusPageRepositoryImpl) Create(statuspage *model.StatusPage) (*model.StatusPage, error) {

	if err := r.db.Create(&statuspage).Error; err != nil {
		return nil, err
	}

	return statuspage, nil
}

func (r *StatusPageRepositoryImpl) GetAllByOrganizationID(orgID string) ([]model.StatusPage, error) {
	var StatusPages []model.StatusPage

	if err := r.db.Preload("Organization").
		Preload("StatusPageMonitors").
		Preload("ComponentGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_groups.display_order ASC, component_groups.name ASC")
		}).
		Where("organization_id = ?", orgID).Find(&StatusPages).Error; err != nil {
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

	err := r.db.Model(&updatedStatusPage).
		Clauses(clause.Returning{}).
		Where("id = ? AND organization_id = ?", statuspageID, orgID).
		Select("Subdomain", "Title", "Description", "IsPublic", "LogoURL", "FaviconURL", "ThemeColor", "CustomCSS").
		Updates(statuspage).Error

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
	if err := r.db.Preload("Organization").
		Preload("StatusPageMonitors").
		Preload("StatusPageMonitors.Monitor").
		Preload("ComponentGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_groups.display_order ASC, component_groups.name ASC")
		}).
		Where("id = ? AND organization_id = ?", statuspageID, orgID).First(&statuspage).Error; err != nil {
		return nil, err
	}
	return &statuspage, nil
}

func (r *StatusPageRepositoryImpl) GetStatusPageMonitor(statusPageID, orgID string) ([]model.StatusPageMonitor, error) {
	var count int64
	if err := r.db.Model(&model.StatusPage{}).Where("id = ? AND organization_id = ?", statusPageID, orgID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("status page not found or unauthorized")
	}

	var statusPageMonitors []model.StatusPageMonitor
	if err := r.db.Preload("Monitor").Where("status_page_id = ?", statusPageID).Order("display_order").Find(&statusPageMonitors).Error; err != nil {
		return nil, err
	}
	return statusPageMonitors, nil
}

func (r *StatusPageRepositoryImpl) GetBySubdomain(subdomain string) (*model.StatusPage, error) {
	var statuspage model.StatusPage
	if err := r.db.Preload("StatusPageMonitors").
		Preload("StatusPageMonitors.Monitor").
		Preload("StatusPageMonitors.ComponentGroup").
		Preload("ComponentGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_groups.display_order ASC, component_groups.name ASC")
		}).
		Where("subdomain = ? AND is_public = ?", subdomain, true).First(&statuspage).Error; err != nil {
		return nil, err
	}

	var monitorIDs []string
	for _, spMonitor := range statuspage.StatusPageMonitors {
		monitorIDs = append(monitorIDs, spMonitor.MonitorID.String())
	}

	if len(monitorIDs) > 0 {
		limit := r.getRecentChecksLimit()
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

func (r *StatusPageRepositoryImpl) TotalOrganizationStatusPage(orgId string) (int64, error) {
	var totalStatusPage int64
	if err := r.db.Model(&model.StatusPage{}).Where("organization_id = ?", orgId).Count(&totalStatusPage).Error; err != nil {
		return 0, err
	}

	return totalStatusPage, nil
}

func (r *StatusPageRepositoryImpl) GetActiveIncidentsByMonitorIDs(monitorIDs []string) ([]model.Incident, error) {
	var incidents []model.Incident
	if len(monitorIDs) == 0 {
		return incidents, nil
	}
	err := r.db.Preload("Monitor").
		Preload("IncidentUpdates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("monitor_id IN ? AND status = ?", monitorIDs, "ongoing").
		Order("started_at DESC").Find(&incidents).Error
	return incidents, err
}

func (r *StatusPageRepositoryImpl) GetRecentResolvedIncidentsByMonitorIDs(monitorIDs []string, days int) ([]model.Incident, error) {
	var incidents []model.Incident
	if len(monitorIDs) == 0 {
		return incidents, nil
	}
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.Preload("Monitor").
		Preload("IncidentUpdates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("monitor_id IN ? AND status = ? AND resolved_at >= ?", monitorIDs, "resolved", since).
		Order("resolved_at DESC").Limit(20).Find(&incidents).Error
	return incidents, err
}

func (r *StatusPageRepositoryImpl) Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error) {
	result := make(map[string]float64)
	if len(monitorIDs) == 0 {
		return result, nil
	}

	sinceDate := time.Now().AddDate(0, 0, -30)

	type uptimeAgg struct {
		MonitorID   string `gorm:"column:monitor_id"`
		TotalChecks int64  `gorm:"column:total_checks"`
		UpChecks    int64  `gorm:"column:up_checks"`
	}

	var dailyAggs []uptimeAgg
	r.db.Table("monitor_daily_stats").
		Select("monitor_id, SUM(total_checks) as total_checks, SUM(up_checks) as up_checks").
		Where("monitor_id IN ? AND date >= ?", monitorIDs, sinceDate).
		Group("monitor_id").
		Scan(&dailyAggs)

	var checkAggs []uptimeAgg
	r.db.Table("checks").
		Select(`
			monitor_id,
			COUNT(*) as total_checks,
			COUNT(CASE WHEN status = 'up' OR (status = 'degraded' AND ((status_code > 0 AND status_code < 500) OR (status_code = 0 AND response_time_ms > 0 AND error_message LIKE 'Response time%'))) THEN 1 END) as up_checks
		`).
		Where("monitor_id IN ? AND checked_at >= ? AND status != 'maintenance'", monitorIDs, sinceDate).
		Group("monitor_id").
		Scan(&checkAggs)

	totals := make(map[string]int64)
	ups := make(map[string]int64)

	for _, d := range dailyAggs {
		totals[d.MonitorID] += d.TotalChecks
		ups[d.MonitorID] += d.UpChecks
	}

	for _, c := range checkAggs {
		totals[c.MonitorID] += c.TotalChecks
		ups[c.MonitorID] += c.UpChecks
	}

	for _, id := range monitorIDs {
		tot := totals[id]
		u := ups[id]
		if tot > 0 {
			result[id] = (float64(u) / float64(tot)) * 100.0
		} else {
			result[id] = 100.0
		}
	}

	return result, nil
}

func (r *StatusPageRepositoryImpl) GetMonitorDailyStats(monitorIDs []string, days int) (map[string][]model.MonitorDailyStat, error) {
	result := make(map[string][]model.MonitorDailyStat)
	if len(monitorIDs) == 0 {
		return result, nil
	}

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	var stats []model.MonitorDailyStat
	if err := r.db.Where("monitor_id IN ? AND date >= ?", monitorIDs, startDate).
		Order("date ASC").Find(&stats).Error; err != nil {
		return nil, err
	}

	for _, s := range stats {
		mID := s.MonitorID.String()
		result[mID] = append(result[mID], s)
	}

	// Also aggregate today's checks from checks table
	type rawDailyAgg struct {
		MonitorID   string `gorm:"column:monitor_id"`
		Date        string `gorm:"column:date"`
		TotalChecks int    `gorm:"column:total_checks"`
		UpChecks    int    `gorm:"column:up_checks"`
		DownChecks  int    `gorm:"column:down_checks"`
	}

	var rawAggs []rawDailyAgg
	r.db.Table("checks").
		Select(`
			monitor_id,
			DATE(checked_at) AS date,
			COUNT(*) AS total_checks,
			COUNT(CASE WHEN status = 'up' OR (status = 'degraded' AND ((status_code > 0 AND status_code < 500) OR (status_code = 0 AND response_time_ms > 0 AND error_message LIKE 'Response time%'))) THEN 1 END) AS up_checks,
			COUNT(CASE WHEN status = 'down' OR (status = 'degraded' AND NOT ((status_code > 0 AND status_code < 500) OR (status_code = 0 AND response_time_ms > 0 AND error_message LIKE 'Response time%'))) THEN 1 END) AS down_checks
		`).
		Where("monitor_id IN ? AND checked_at >= ? AND status != 'maintenance'", monitorIDs, startDate).
		Group("monitor_id, DATE(checked_at)").
		Scan(&rawAggs)

	if len(rawAggs) > 0 {
		for _, raw := range rawAggs {
			if len(raw.Date) < 10 {
				continue
			}
			parsedDate, err := time.Parse("2006-01-02", raw.Date[:10])
			if err != nil {
				continue
			}
			mID := raw.MonitorID
			dateStr := parsedDate.Format("2006-01-02")
			existingList := result[mID]
			found := false
			for idx, s := range existingList {
				if s.Date.Format("2006-01-02") == dateStr {
					tot := s.TotalChecks + raw.TotalChecks
					up := s.UpChecks + raw.UpChecks
					down := s.DownChecks + raw.DownChecks
					pct := 100.0
					if tot > 0 {
						pct = (float64(up) / float64(tot)) * 100.0
					}
					existingList[idx].TotalChecks = tot
					existingList[idx].UpChecks = up
					existingList[idx].DownChecks = down
					existingList[idx].UptimePercentage = pct
					found = true
					break
				}
			}
			if !found {
				pct := 100.0
				if raw.TotalChecks > 0 {
					pct = (float64(raw.UpChecks) / float64(raw.TotalChecks)) * 100.0
				}
				mUUID, _ := uuid.Parse(mID)
				existingList = append(existingList, model.MonitorDailyStat{
					MonitorID:        mUUID,
					Date:             parsedDate,
					TotalChecks:      raw.TotalChecks,
					UpChecks:         raw.UpChecks,
					DownChecks:       raw.DownChecks,
					UptimePercentage: pct,
				})
			}
			result[mID] = existingList
		}
	}

	return result, nil
}

func (r *StatusPageRepositoryImpl) GetActiveMaintenances(statusPageID string) ([]model.ScheduledMaintenance, []model.ScheduledMaintenance, error) {
	now := time.Now()
	var ongoing []model.ScheduledMaintenance
	var upcoming []model.ScheduledMaintenance

	err := r.db.Preload("AffectedMonitors.Monitor").
		Where("status_page_id = ? AND (status = 'in_progress' OR (status = 'scheduled' AND scheduled_start <= ? AND scheduled_end >= ?))", statusPageID, now, now).
		Order("scheduled_start ASC").
		Find(&ongoing).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.Preload("AffectedMonitors.Monitor").
		Where("status_page_id = ? AND status = 'scheduled' AND scheduled_start > ?", statusPageID, now).
		Order("scheduled_start ASC").
		Limit(5).
		Find(&upcoming).Error
	if err != nil {
		return nil, nil, err
	}

	return ongoing, upcoming, nil
}

func (r *StatusPageRepositoryImpl) CreateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs []string) (*model.ScheduledMaintenance, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(maintenance).Error; err != nil {
			return err
		}

		for _, mID := range monitorIDs {
			parsedID, err := uuid.Parse(mID)
			if err != nil {
				continue
			}
			mm := model.MaintenanceMonitor{
				MaintenanceID: maintenance.ID,
				MonitorID:     parsedID,
			}
			if err := tx.Create(&mm).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var loaded model.ScheduledMaintenance
	if err := r.db.Preload("AffectedMonitors.Monitor").Where("id = ?", maintenance.ID).First(&loaded).Error; err != nil {
		return maintenance, nil
	}
	return &loaded, nil
}

func (r *StatusPageRepositoryImpl) GetMaintenancesByStatusPageID(statusPageID, orgID string) ([]model.ScheduledMaintenance, error) {
	var list []model.ScheduledMaintenance
	err := r.db.Preload("AffectedMonitors.Monitor").
		Where("status_page_id = ? AND organization_id = ?", statusPageID, orgID).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *StatusPageRepositoryImpl) GetMaintenanceByID(maintenanceID, orgID string) (*model.ScheduledMaintenance, error) {
	var m model.ScheduledMaintenance
	err := r.db.Preload("AffectedMonitors.Monitor").
		Where("id = ? AND organization_id = ?", maintenanceID, orgID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *StatusPageRepositoryImpl) UpdateMaintenance(maintenance *model.ScheduledMaintenance, monitorIDs *[]string) (*model.ScheduledMaintenance, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(maintenance).Error; err != nil {
			return err
		}

		if monitorIDs != nil {
			if err := tx.Where("maintenance_id = ?", maintenance.ID).Delete(&model.MaintenanceMonitor{}).Error; err != nil {
				return err
			}
			for _, mID := range *monitorIDs {
				parsedID, err := uuid.Parse(mID)
				if err != nil {
					continue
				}
				mm := model.MaintenanceMonitor{
					MaintenanceID: maintenance.ID,
					MonitorID:     parsedID,
				}
				if err := tx.Create(&mm).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var loaded model.ScheduledMaintenance
	if err := r.db.Preload("AffectedMonitors.Monitor").Where("id = ?", maintenance.ID).First(&loaded).Error; err != nil {
		return maintenance, nil
	}
	return &loaded, nil
}

func (r *StatusPageRepositoryImpl) DeleteMaintenance(maintenanceID, orgID string) error {
	return r.db.Where("id = ? AND organization_id = ?", maintenanceID, orgID).Delete(&model.ScheduledMaintenance{}).Error
}

func (r *StatusPageRepositoryImpl) GetIncidentHistory(monitorIDs []string, days int) ([]model.Incident, error) {
	if len(monitorIDs) == 0 {
		return []model.Incident{}, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var incidents []model.Incident
	err := r.db.Preload("Monitor").
		Preload("IncidentUpdates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("monitor_id IN ? AND status = ? AND started_at >= ?", monitorIDs, "resolved", cutoff).
		Order("started_at DESC").
		Find(&incidents).Error
	return incidents, err
}

func (r *StatusPageRepositoryImpl) CreateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error) {
	if err := r.db.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (r *StatusPageRepositoryImpl) GetComponentGroupsByStatusPageID(statusPageID string) ([]model.ComponentGroup, error) {
	var groups []model.ComponentGroup
	err := r.db.Where("status_page_id = ?", statusPageID).
		Order("display_order ASC, name ASC").
		Find(&groups).Error
	return groups, err
}

func (r *StatusPageRepositoryImpl) GetComponentGroupByID(groupID, statusPageID string) (*model.ComponentGroup, error) {
	var group model.ComponentGroup
	err := r.db.Where("id = ? AND status_page_id = ?", groupID, statusPageID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *StatusPageRepositoryImpl) UpdateComponentGroup(group *model.ComponentGroup) (*model.ComponentGroup, error) {
	if err := r.db.Save(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (r *StatusPageRepositoryImpl) DeleteComponentGroup(groupID, statusPageID string) error {
	_ = r.db.Model(&model.StatusPageMonitor{}).
		Where("status_page_id = ? AND group_id = ?", statusPageID, groupID).
		Update("group_id", nil).Error

	return r.db.Where("id = ? AND status_page_id = ?", groupID, statusPageID).
		Delete(&model.ComponentGroup{}).Error
}

func (r *StatusPageRepositoryImpl) AssignMonitorToGroup(statusPageID, monitorID string, groupID *string) error {
	var parsedGroupID *uuid.UUID
	if groupID != nil && *groupID != "" {
		gUUID, err := uuid.Parse(*groupID)
		if err == nil {
			parsedGroupID = &gUUID
		}
	}
	return r.db.Model(&model.StatusPageMonitor{}).
		Where("status_page_id = ? AND monitor_id = ?", statusPageID, monitorID).
		Update("group_id", parsedGroupID).Error
}

func (r *StatusPageRepositoryImpl) ValidateMonitorsBelongToOrg(orgID string, monitorIDs []string) error {
	if len(monitorIDs) == 0 {
		return nil
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return errors.New("invalid organization id")
	}

	uniqueMap := make(map[uuid.UUID]struct{})
	for _, idStr := range monitorIDs {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			return errors.New("one or more monitors do not belong to this organization")
		}
		uniqueMap[parsed] = struct{}{}
	}

	if len(uniqueMap) == 0 {
		return nil
	}

	uniqueList := make([]uuid.UUID, 0, len(uniqueMap))
	for id := range uniqueMap {
		uniqueList = append(uniqueList, id)
	}

	var count int64
	err = r.db.Model(&model.Monitor{}).
		Where("organization_id = ? AND id IN ?", orgUUID, uniqueList).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count != int64(len(uniqueMap)) {
		return errors.New("one or more monitors do not belong to this organization")
	}

	return nil
}
