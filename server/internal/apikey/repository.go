package apikey

import (
	"netra-uptime/internal/model"
	"time"

	"gorm.io/gorm"
)

type APIKeyRepository interface {
	Create(key *model.APIKey) error
	GetAllByOrgID(orgID string) ([]model.APIKey, error)
	GetByHash(keyHash string) (*model.APIKey, error)
	Delete(orgID, keyID string) error
	UpdateLastUsed(keyID string, ip string) error
	GetCreatorRole(orgID, userID string) (string, error)
	RevokeKeysByOrgAndUser(orgID, userID string) error
}

type APIKeyRepositoryImpl struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &APIKeyRepositoryImpl{db: db}
}

func (r *APIKeyRepositoryImpl) Create(key *model.APIKey) error {
	return r.db.Create(key).Error
}

func (r *APIKeyRepositoryImpl) GetAllByOrgID(orgID string) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.Preload("CreatedBy").Where("organization_id = ?", orgID).Order("created_at desc").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepositoryImpl) GetByHash(keyHash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Where("key_hash = ?", keyHash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepositoryImpl) Delete(orgID, keyID string) error {
	return r.db.Where("id = ? AND organization_id = ?", keyID, orgID).Delete(&model.APIKey{}).Error
}

func (r *APIKeyRepositoryImpl) UpdateLastUsed(keyID string, ip string) error {
	now := time.Now()
	return r.db.Model(&model.APIKey{}).Where("id = ?", keyID).Updates(map[string]interface{}{
		"last_used_at": now,
		"last_used_ip": ip,
	}).Error
}

func (r *APIKeyRepositoryImpl) GetCreatorRole(orgID, userID string) (string, error) {
	var member model.OrganizationMember
	err := r.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *APIKeyRepositoryImpl) RevokeKeysByOrgAndUser(orgID, userID string) error {
	return r.db.Where("organization_id = ? AND created_by_id = ?", orgID, userID).Delete(&model.APIKey{}).Error
}
