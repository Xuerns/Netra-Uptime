package notification

import (
	"netra-uptime/internal/model"

	"gorm.io/gorm"
)


type ChannelRepository interface {
	Create(channel *model.NotificationChannel) error
	GetAll(orgID string) ([]model.NotificationChannel, error)
	GetByID(orgID, channelID string) (*model.NotificationChannel, error)
	Update(channel *model.NotificationChannel) error
	Delete(orgID, channelID string) error
	CountByOrg(orgID string) (int64, error)

	CreateLog(log *model.NotificationLog) error
	GetLogs(orgID string, limit, offset int) ([]model.NotificationLog, int64, error)
}


type channelRepositoryImpl struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepositoryImpl{db: db}
}

func (r *channelRepositoryImpl) Create(ch *model.NotificationChannel) error {
	return r.db.Create(ch).Error
}

func (r *channelRepositoryImpl) GetAll(orgID string) ([]model.NotificationChannel, error) {
	var channels []model.NotificationChannel
	if err := r.db.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *channelRepositoryImpl) GetByID(orgID, channelID string) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	if err := r.db.
		Where("organization_id = ? AND id = ?", orgID, channelID).
		First(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *channelRepositoryImpl) Update(ch *model.NotificationChannel) error {
	return r.db.Save(ch).Error
}

func (r *channelRepositoryImpl) Delete(orgID, channelID string) error {
	return r.db.
		Where("organization_id = ? AND id = ?", orgID, channelID).
		Delete(&model.NotificationChannel{}).Error
}

func (r *channelRepositoryImpl) CountByOrg(orgID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.NotificationChannel{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// CreateLog persists a single notification dispatch record.
func (r *channelRepositoryImpl) CreateLog(entry *model.NotificationLog) error {
	return r.db.Create(entry).Error
}

// GetLogs fetches paginated notification logs for monitors that belong to the org.
// We join through monitors to enforce tenant isolation without storing org_id on the log table.
func (r *channelRepositoryImpl) GetLogs(orgID string, limit, offset int) ([]model.NotificationLog, int64, error) {
	var logs []model.NotificationLog
	var total int64

	baseQuery := r.db.Model(&model.NotificationLog{}).
		Joins("LEFT JOIN monitors ON monitors.id = notification_logs.monitor_id").
		Where("monitors.organization_id = ? OR notification_logs.monitor_id IS NULL", orgID)

	// We scope by org only when monitor_id is present. Org-level channel logs
	// that have no monitor_id are excluded here intentionally; they'll be fetched
	// by a separate query when needed.
	scopedQuery := r.db.Model(&model.NotificationLog{}).
		Joins("LEFT JOIN monitors ON monitors.id = notification_logs.monitor_id").
		Where("monitors.organization_id = ?", orgID)

	_ = baseQuery // keep for future use

	if err := scopedQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := scopedQuery.
		Order("notification_logs.sent_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
