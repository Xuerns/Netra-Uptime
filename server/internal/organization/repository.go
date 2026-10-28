package organization

import (
	"errors"
	"netra-uptime/internal/model"
	"time"

	"gorm.io/gorm"
)

type OrganizationRepository interface {
	GetAllSlugOrganization(slug string) (*[]model.Organization, error)
	GetMemberById(orgId, memberId string) (*model.OrganizationMember, error)
	GetOrganizationMember(orgId, memberId string) error
	GetOrganizationAllMember(orgId string) (*[]model.OrganizationMember, error)
	GetAllOrganizationByUserId(userId string) (*[]model.Organization, error)
	CreateOrganization(org *model.Organization, ownerMember *model.OrganizationMember, invitations *[]model.OrganizationInvitation) error
	AcceptInvitation(invitation *model.OrganizationInvitation, member *model.OrganizationMember) error
	RejectInvitation(invitation *model.OrganizationInvitation) error
	KickMember(orgId, memberId string) error
	InviteMember(invitations *[]model.OrganizationInvitation) error
	GetInvitation(userId string) (*[]model.OrganizationInvitation, error)
	GetInvitationById(invitationId string) (*model.OrganizationInvitation, error)
	Disband(orgId, ownerId string) error
	UpdateOrganization(updateData *map[string]interface{}, orgId string) error
	SearchPersons(name string, excludeUserId string, limit int, offset int, orgId string) ([]model.User, error)
	PromoteToAdmin(userId, orgId string) error
	DemotedToMember(userId, orgId string) error
	TransferOwnership(ownerId, newOwnerId, orgId string) error
	LeaveOrganization(userId, orgId string) error
	TotalOrganizationUser(useId string) (int64, error)
	TotalOrganizationMember(orgId string) (int64, error)
	RevokeAPIKeysByMember(orgId, memberId string) error
}

type OrganizationRepositoryImpl struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &OrganizationRepositoryImpl{db: db}
}

func (r *OrganizationRepositoryImpl) GetAllSlugOrganization(slug string) (*[]model.Organization, error) {
	var organizations []model.Organization
	err := r.db.Where("slug = ?", slug).Find(&organizations).Error
	if err != nil {
		return nil, err
	}
	if len(organizations) == 0 {
		return nil, nil
	}
	return &organizations, nil
}

func (r *OrganizationRepositoryImpl) GetMemberById(orgId, memberId string) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.Where("organization_id = ? AND user_id = ?", orgId, memberId).First(&member).Error

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *OrganizationRepositoryImpl) GetOrganizationMember(orgId, memberId string) error {
	var member model.OrganizationMember
	err := r.db.Where("organization_id = ? AND user_id = ?", orgId, memberId).First(&member).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepositoryImpl) GetOrganizationAllMember(orgId string) (*[]model.OrganizationMember, error) {
	var member []model.OrganizationMember
	err := r.db.Select("id", "organization_id", "user_id", "role").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "avatar_url")
	}).Where("organization_id = ?", orgId).Find(&member).Error

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *OrganizationRepositoryImpl) GetAllOrganizationByUserId(userId string) (*[]model.Organization, error) {
	var organizations []model.Organization
	err := r.db.
		Preload("OrganizationMembers", func(db *gorm.DB) *gorm.DB {
			return db.Select("organization_id")
		}).
		Preload("Monitors", func(db *gorm.DB) *gorm.DB {
			return db.Select("organization_id, status")
		}).
		Joins("JOIN organization_members on organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userId).
		Find(&organizations).Error
	if err != nil {
		return nil, err
	}
	return &organizations, nil
}

func (r *OrganizationRepositoryImpl) CreateOrganization(org *model.Organization, ownerMember *model.OrganizationMember, invitations *[]model.OrganizationInvitation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(org).Error; err != nil {
			return err
		}

		ownerMember.OrganizationID = org.ID
		if err := tx.Create(ownerMember).Error; err != nil {
			return err
		}

		if invitations != nil && len(*invitations) > 0 {
			for i := range *invitations {
				(*invitations)[i].OrganizationID = org.ID
				(*invitations)[i].Status = "pending"
				(*invitations)[i].InviterID = ownerMember.ID
				(*invitations)[i].ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
			}

			if err := tx.Create(invitations).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *OrganizationRepositoryImpl) AcceptInvitation(invitation *model.OrganizationInvitation, member *model.OrganizationMember) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if invitation.Status == "expired" {
			_ = tx.Save(&invitation).Error
			return errors.New("Invitation Expired")
		}

		if err := tx.Save(&invitation).Error; err != nil {
			return err
		}

		member.OrganizationID = invitation.OrganizationID
		member.UserID = invitation.InviteeID
		member.Role = invitation.Role

		if err := tx.Create(member).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *OrganizationRepositoryImpl) RejectInvitation(invitation *model.OrganizationInvitation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if invitation.Status == "expired" {
			_ = tx.Save(&invitation).Error
			return errors.New("Invitation Expired")
		}

		if err := tx.Save(&invitation).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *OrganizationRepositoryImpl) KickMember(orgId, memberId string) error {
	return r.db.Where("organization_id = ? AND user_id = ?", orgId, memberId).Delete(&model.OrganizationMember{}).Error
}

func (r *OrganizationRepositoryImpl) InviteMember(invitations *[]model.OrganizationInvitation) error {
	return r.db.Create(invitations).Error
}

func (r *OrganizationRepositoryImpl) GetInvitation(userId string) (*[]model.OrganizationInvitation, error) {
	var invitations []model.OrganizationInvitation

	if err := r.db.Preload("Organization").Preload("Inviter").Where("invitee_id = ? AND status = ? AND expires_at > ?", userId, "pending", time.Now()).Find(&invitations).Error; err != nil {
		return nil, err
	}
	return &invitations, nil
}

func (r *OrganizationRepositoryImpl) GetInvitationById(invitationId string) (*model.OrganizationInvitation, error) {
	var invitation model.OrganizationInvitation
	if err := r.db.Where("id = ?", invitationId).First(&invitation).Error; err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *OrganizationRepositoryImpl) Disband(orgId, ownerId string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("organization_id = ?", orgId).Delete(&model.OrganizationMember{}).Error; err != nil {
			return err
		}

		if err := tx.Where("organization_id = ?", orgId).Delete(&model.OrganizationInvitation{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", orgId).Delete(&model.Organization{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *OrganizationRepositoryImpl) UpdateOrganization(updateData *map[string]interface{}, orgId string) error {
	if err := r.db.Model(&model.Organization{}).Where("id = ?", orgId).Updates(*updateData).Error; err != nil {
		return err
	}
	return nil
}

func (r *OrganizationRepositoryImpl) SearchPersons(name string, excludeUserId string, limit int, offset int, orgId string) ([]model.User, error) {
	var users []model.User
	query := r.db.Select("id", "name", "avatar_url").
		Where("name ILIKE ?", "%"+name+"%").
		Where("id != ?", excludeUserId)

	if orgId != "" {
		query = query.Where("id NOT IN (SELECT user_id FROM organization_members WHERE organization_id = ?)", orgId).
			Where("id NOT IN (SELECT invitee_id FROM organization_invitations WHERE organization_id = ? AND status = 'pending')", orgId)
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users, nil
}

func (r *OrganizationRepositoryImpl) PromoteToAdmin(userId, orgId string) error {
	return r.db.Model(&model.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", orgId, userId).Update("role", "admin").Error
}

func (r *OrganizationRepositoryImpl) DemotedToMember(userId, orgId string) error {
	return r.db.Model(&model.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", orgId, userId).Update("role", "member").Error
}

func (r *OrganizationRepositoryImpl) TransferOwnership(ownerId, newOwnerId, orgId string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Promote new user to owner
		if err := tx.Model(&model.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", orgId, newOwnerId).Update("role", "owner").Error; err != nil {
			return err
		}

		// 2. Demote current owner to admin
		if err := tx.Model(&model.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", orgId, ownerId).Update("role", "admin").Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *OrganizationRepositoryImpl) LeaveOrganization(userId, orgId string) error {
	return r.db.Where("organization_id = ? AND user_id = ?", orgId, userId).Delete(&model.OrganizationMember{}).Error
}

func (r *OrganizationRepositoryImpl) TotalOrganizationUser(useId string) (int64, error) {
	var totalOrganizationUser int64
	if err := r.db.Model(&model.OrganizationMember{}).Where("user_id = ?", useId).Count(&totalOrganizationUser).Error; err != nil {
		return 0, err
	}
	return totalOrganizationUser, nil
}

func (r *OrganizationRepositoryImpl) TotalOrganizationMember(orgId string) (int64, error) {
	var totalOrganizationMember int64
	if err := r.db.Model(&model.OrganizationMember{}).Where("organization_id = ?", orgId).Count(&totalOrganizationMember).Error; err != nil {
		return 0, err
	}
	return totalOrganizationMember, nil
}

func (r *OrganizationRepositoryImpl) RevokeAPIKeysByMember(orgId, memberId string) error {
	return r.db.Where("organization_id = ? AND created_by_id = ?", orgId, memberId).Delete(&model.APIKey{}).Error
}
