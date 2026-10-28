package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"netra-uptime/config"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/email"
	"netra-uptime/pkg/jwt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req RegisterRequest) (string, error)
	Login(req LoginRequest) (*LoginResponse, error)
	Logout(refreshToken string) error
	RefreshAccessToken(token string) (*LoginResponse, error)
	GetProfile(userId string) (*ProfileResponse, error)
	UpdateProfile(userId string, req UpdateProfileRequest) (*ProfileResponse, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(newPassword, resetToken string) error
	InitPasswordChange(userId string, req InitPasswordChangeRequest) error
	VerifyPasswordChange(userId string, req VerifyPasswordChangeRequest) error
}

type AuthServiceImpl struct {
	repo AuthRepository
	rdb  *redis.Client
	cfg  *config.Config
}

func NewAuthService(repo AuthRepository, rdb *redis.Client, cfg ...*config.Config) AuthService {
	var c *config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return &AuthServiceImpl{repo: repo, rdb: rdb, cfg: c}
}

func (s *AuthServiceImpl) Register(req RegisterRequest) (string, error) {
	if s.cfg != nil && !s.cfg.AllowRegistration {
		if userCounter, ok := s.repo.(interface{ CountUsers() (int64, error) }); ok {
			count, err := userCounter.CountUsers()
			if err != nil || count > 0 {
				return "", errors.New("Public registration is disabled on this instance")
			}
		} else {
			return "", errors.New("Public registration is disabled on this instance")
		}
	}

	var newUser = model.User{}

	user, _ := s.repo.FindUserByEmail(req.Email)

	if user != nil {
		return "", errors.New("User Already Exist")
	}

	hashedPasword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)

	if err != nil {
		return "", errors.New("Something wrong, try again later")
	}

	newUser.Name = req.Name
	newUser.Email = req.Email
	hashedPaswordStr := string(hashedPasword)
	newUser.PasswordHash = &hashedPaswordStr
	colors := "10b981,3b82f6,4edea3,adc6ff,ffb4ab"
	newUser.AvatarURL = "https://api.dicebear.com/9.x/shapes/svg?seed=" + url.QueryEscape(req.Email) + "&backgroundColor=131314&shape1Color=" + colors + "&shape2Color=" + colors + "&shape3Color=" + colors

	orgName := req.Name + "Workspace"

	slug := strings.ToLower(req.Name)
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano()%10000, 10)
	orgSlug := slug + "-workspace-" + uniqueSuffix

	var newOrg = model.Organization{
		Name: orgName,
		Slug: orgSlug,
	}

	var orgOwner = model.OrganizationMember{
		Role: "owner",
	}

	if err := s.repo.CreateUserWithDefaultOrg(&newUser, &newOrg, &orgOwner); err != nil {
		return "", errors.New("Failed to register")
	}

	return "User registered successfully", nil
}

func (s *AuthServiceImpl) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.FindUserByEmail(req.Email)
	// Validation for user Existence
	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	// Validation for Login method
	if user.PasswordHash == nil {
		return nil, errors.New("Invalid email or password")
	}

	// Validation for Password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))

	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	// Make Access Token and Refresh Token
	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))

	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	expiryRefresh, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))

	if err != nil {
		expiryRefresh = 168 * time.Hour
	}

	accessToken, err := jwt.GenerateToken(user.ID.String(), jwt.TokenTypeAccess, expiryAccess)

	if err != nil {
		return nil, errors.New("Failed to generate access token")
	}

	refreshToken, err := jwt.GenerateToken(user.ID.String(), jwt.TokenTypeRefresh, expiryRefresh)

	if err != nil {
		return nil, errors.New("Failed to generate refresh token")
	}

	// Hashing Refresh Token
	// use Sha256 because bcrypt max char is 72 bytes
	hash := sha256.New()
	hash.Write([]byte(refreshToken))
	hashedRefreshToken := hex.EncodeToString(hash.Sum(nil))

	// Save Refresh Token
	var createRefreshToken model.RefreshToken

	createRefreshToken.UserID = user.ID
	createRefreshToken.Token = string(hashedRefreshToken)
	createRefreshToken.ExpiresAt = time.Now().Add(expiryRefresh)

	// Enforce session limit per user (VULN-04)
	_ = s.repo.LimitActiveRefreshTokens(user.ID.String(), 5)

	if err := s.repo.CreateRefreshToken(&createRefreshToken); err != nil {
		return nil, errors.New("Failed to store refresh token")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) Logout(refreshToken string) error {
	hash := sha256.New()
	hash.Write([]byte(refreshToken))
	hashedToken := hex.EncodeToString(hash.Sum(nil))
	if err := s.repo.RevokeRefreshToken(hashedToken); err != nil {
		return errors.New("Failed to revoke refresh token")
	}
	return nil
}

func (s *AuthServiceImpl) RefreshAccessToken(token string) (*LoginResponse, error) {
	hash := sha256.New()
	hash.Write([]byte(token))
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	refreshToken, err := s.repo.FindRefreshTokenByToken(hashedToken)

	if err != nil {
		return nil, errors.New("Refreshtoken not found")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		_ = s.repo.RevokeRefreshToken(hashedToken)
		return nil, errors.New("Refresh token expired")
	}

	// Revoke old refresh token (Token Rotation - VULN-03)
	_ = s.repo.RevokeRefreshToken(hashedToken)

	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))
	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	expiryRefresh, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))
	if err != nil {
		expiryRefresh = 168 * time.Hour
	}

	newAccessToken, err := jwt.GenerateToken(refreshToken.UserID.String(), jwt.TokenTypeAccess, expiryAccess)
	if err != nil {
		return nil, errors.New("Failed to generate access token")
	}

	newRefreshToken, err := jwt.GenerateToken(refreshToken.UserID.String(), jwt.TokenTypeRefresh, expiryRefresh)
	if err != nil {
		return nil, errors.New("Failed to generate refresh token")
	}

	newHash := sha256.New()
	newHash.Write([]byte(newRefreshToken))
	hashedNewToken := hex.EncodeToString(newHash.Sum(nil))

	_ = s.repo.LimitActiveRefreshTokens(refreshToken.UserID.String(), 5)

	createRefreshToken := model.RefreshToken{
		UserID:    refreshToken.UserID,
		Token:     string(hashedNewToken),
		ExpiresAt: time.Now().Add(expiryRefresh),
	}

	if err := s.repo.CreateRefreshToken(&createRefreshToken); err != nil {
		return nil, errors.New("Failed to store refresh token")
	}

	return &LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthServiceImpl) GetProfile(userId string) (*ProfileResponse, error) {
	user, err := s.repo.GetUserById(userId)

	if err != nil {
		return nil, errors.New("User does't exist")
	}

	response := ProfileResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}

	return &response, nil
}

func (s *AuthServiceImpl) UpdateProfile(userId string, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.repo.UpdateProfile(userId, req.Name)
	if err != nil {
		return nil, errors.New("User doesn't exist")
	}

	response := ProfileResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}

	return &response, nil
}

func (s *AuthServiceImpl) ForgotPassword(email string) (string, error) {
	user, err := s.repo.FindUserByEmail(email)

	if err != nil {
		// User not found — return empty URL (no error) to prevent email enumeration
		return "", nil
	}

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)

	hash := sha256.New()
	hash.Write([]byte(randomString))
	token := hex.EncodeToString(hash.Sum(nil))

	var resetToken = model.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateResetToken(&resetToken); err != nil {
		return "", errors.New("Internal Server Error")
	}

	ORIGIN := os.Getenv("ALLOWED_ORIGIN")
	url := fmt.Sprintf("%s/auth/forgot?token=%s", ORIGIN, randomString)

	return url, nil
}

func (s *AuthServiceImpl) ResetPassword(newPassword, resetToken string) error {
	hash := sha256.New()
	hash.Write([]byte(resetToken))
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	userId, err := s.repo.GetUserByResetToken(hashedToken)

	if err != nil {
		return errors.New("User doesn't exist")
	}

	user, err := s.repo.GetUserById(userId)

	if err != nil {
		return errors.New("User doesn't exist")
	}

	token, err := s.repo.DeleteResetToken(user.ID.String(), hashedToken)

	if err != nil {
		return errors.New("Something wrong, please try again")
	}

	if token.ExpiredAt.Before(time.Now()) {
		return errors.New("Please, request new link")
	}

	hashedPasword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)

	if err != nil {
		return errors.New("Internal Server Error")
	}

	if err := s.repo.CreateNewPassword(user.ID.String(), string(hashedPasword)); err != nil {
		return errors.New("Failed to change password")
	}

	// Revoke all existing sessions for this user (VULN-05)
	_ = s.repo.RevokeAllRefreshTokensByUserID(user.ID.String())

	return nil
}

func (s *AuthServiceImpl) InitPasswordChange(userId string, req InitPasswordChangeRequest) error {
	user, err := s.repo.GetUserById(userId)
	if err != nil {
		return errors.New("User not found")
	}

	if user.PasswordHash == nil {
		return errors.New("Password not set for user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("Incorrect current password")
	}

	// Generate 6-digit random code securely
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return errors.New("Failed to generate code")
	}
	codeNum := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if codeNum < 0 {
		codeNum = -codeNum
	}
	codeStr := fmt.Sprintf("%06d", codeNum%1000000)

	hashedCode, err := bcrypt.GenerateFromPassword([]byte(codeStr), 10)
	if err != nil {
		return errors.New("Failed to hash code")
	}

	ctx := context.Background()
	key := fmt.Sprintf("password_change_code:%s", userId)
	if err := s.rdb.Set(ctx, key, string(hashedCode), 15*time.Minute).Err(); err != nil {
		return errors.New("Failed to save verification code")
	}

	if err := email.SendVerificationCodeEmail(user.Email, codeStr); err != nil {
		return errors.New("Failed to send verification email")
	}

	return nil
}

func (s *AuthServiceImpl) VerifyPasswordChange(userId string, req VerifyPasswordChangeRequest) error {
	ctx := context.Background()
	codeKey := fmt.Sprintf("password_change_code:%s", userId)
	attemptsKey := fmt.Sprintf("password_change_attempts:%s", userId)

	if s.rdb != nil {
		attemptsVal, err := s.rdb.Get(ctx, attemptsKey).Result()
		if err == nil {
			if attempts, err := strconv.Atoi(attemptsVal); err == nil && attempts >= 5 {
				return errors.New("Too many failed attempts. Verification code invalidated.")
			}
		}
	}

	if s.rdb == nil {
		return errors.New("Verification code expired or not found")
	}

	hashedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return errors.New("Verification code expired or not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(req.Code)); err != nil {
		if s.rdb != nil {
			attempts, incrErr := s.rdb.Incr(ctx, attemptsKey).Result()
			if incrErr == nil {
				s.rdb.Expire(ctx, attemptsKey, 15*time.Minute)
				if attempts >= 5 {
					s.rdb.Del(ctx, codeKey, attemptsKey)
					return errors.New("Too many failed attempts. Verification code invalidated.")
				}
			}
		}
		return errors.New("Invalid verification code")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		return errors.New("Internal Server Error")
	}

	if err := s.repo.CreateNewPassword(userId, string(hashedPassword)); err != nil {
		return errors.New("Failed to change password")
	}

	if s.rdb != nil {
		s.rdb.Del(ctx, attemptsKey, codeKey)
	}

	// Revoke all existing sessions for this user (VULN-05)
	_ = s.repo.RevokeAllRefreshTokensByUserID(userId)

	return nil
}
