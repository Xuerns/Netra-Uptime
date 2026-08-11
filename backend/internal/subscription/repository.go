package subscription

import (
	"PulseWatch/internal/model"

	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	CreateSubscription(subscription *model.Subscription) error
	UpdateSubscription(subscription *model.Subscription) error
	FindSubscriptionByOrgId(orgId string) (*model.Subscription, error)
	CreateTransaction(transaction *model.Transaction) error
	FindPendingTransaction(OrganizationId, plan string) (*model.Transaction, error)
	UpdateTransaction(transaction *model.Transaction) error
	FincTransactionByOrderId(orderId string) (*model.Transaction, error)
	GetAllSubsByUserId(userId string) ([]model.Subscription, error)
}

type SubscriptionRepositoryImpl struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &SubscriptionRepositoryImpl{db: db}
}

func (r *SubscriptionRepositoryImpl) CreateSubscription(subscription *model.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *SubscriptionRepositoryImpl) UpdateSubscription(subscription *model.Subscription) error {
	return r.db.Save(subscription).Error
}

func (r *SubscriptionRepositoryImpl) FindSubscriptionByOrgId(orgId string) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := r.db.Where("organization_id = ?", orgId).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepositoryImpl) CreateTransaction(transaction *model.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *SubscriptionRepositoryImpl) FindPendingTransaction(OrganizationId, plan string) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.Where("organization_id = ? AND status = ? AND plan_type = ?", OrganizationId, "pending", plan).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *SubscriptionRepositoryImpl) UpdateTransaction(transaction *model.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *SubscriptionRepositoryImpl) FincTransactionByOrderId(orderId string) (*model.Transaction, error) {
	var transaction model.Transaction
	if err := r.db.Where("order_id = ?", orderId).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *SubscriptionRepositoryImpl) GetAllSubsByUserId(userId string) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	if err := r.db.Preload("Organization").
		Joins("JOIN organization_members ON subscriptions.organization_id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userId).
		Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}
