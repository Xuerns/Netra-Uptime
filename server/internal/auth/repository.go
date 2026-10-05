package auth

import (
	"PulseWatch/internal/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	FindUserByEmail(email string) (*model.User, error)
	CreateRefreshToken(refreshToken *model.RefreshToken) error
	RevokeRefreshToken(hashedToken string) error
	FindRefreshTokenByToken(hashedToken string) (*model.RefreshToken, error)
	CreateUserWithDefaultOrg(user *model.User, org *model.Organization, orgMember *model.OrganizationMember, subscription *model.Subscription) error
	GetUserById(userId string) (*model.User, error)
	UpdateProfile(userId string, name string) (*model.User, error)
	CreateResetToken(token *model.PasswordResetToken) error
	CreateNewPassword(userId, password string) error
	DeleteResetToken(userId, token string) (*model.PasswordResetToken, error)
	GetUserByResetToken(token string) (string, error)
}

type AuthRepositoryImpl struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

func (r *AuthRepositoryImpl) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepositoryImpl) CreateRefreshToken(refreshToken *model.RefreshToken) error {
	return r.db.Create(refreshToken).Error
}

func (r *AuthRepositoryImpl) RevokeRefreshToken(hashedToken string) error {
	return r.db.Where("token = ?", hashedToken).Delete(&model.RefreshToken{}).Error
}

func (r *AuthRepositoryImpl) FindRefreshTokenByToken(hashedToken string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := r.db.Where("token = ?", hashedToken).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *AuthRepositoryImpl) CreateUserWithDefaultOrg(user *model.User, org *model.Organization, orgMember *model.OrganizationMember, subscription *model.Subscription) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		orgMember.UserID = user.ID

		if err := tx.Create(org).Error; err != nil {
			return err
		}

		orgMember.OrganizationID = org.ID

		if err := tx.Create(orgMember).Error; err != nil {
			return err
		}

		subscription.OrganizationID = org.ID
		if err := tx.Create(subscription).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *AuthRepositoryImpl) GetUserById(userId string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepositoryImpl) UpdateProfile(userId string, name string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}
	user.Name = name
	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepositoryImpl) GetUserByResetToken(token string) (string, error) {
	var passwordResetToken model.PasswordResetToken
	if err := r.db.Where("token = ?", token).First(&passwordResetToken).Error; err != nil {
		return "", err
	}
	return passwordResetToken.UserID.String(), nil
}

func (r *AuthRepositoryImpl) CreateResetToken(token *model.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *AuthRepositoryImpl) CreateNewPassword(userId, hashPassword string) error {
	var user model.User
	if err := r.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return err
	}

	user.PasswordHash = &hashPassword
	if err := r.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (r *AuthRepositoryImpl) DeleteResetToken(userId, token string) (*model.PasswordResetToken, error) {
	var passwordResetToken model.PasswordResetToken
	if err := r.db.Where("user_id = ? AND token = ?", userId, token).First(&passwordResetToken).Error; err != nil {
		return nil, err
	}
	if err := r.db.Delete(&passwordResetToken).Error; err != nil {
		return nil, err
	}
	return &passwordResetToken, nil
}
