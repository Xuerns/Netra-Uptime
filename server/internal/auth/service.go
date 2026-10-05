package auth

import (
	"PulseWatch/internal/model"
	"PulseWatch/pkg/jwt"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"PulseWatch/pkg/email"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService interface {
	Register(req RegisterRequest) (string, error)
	Login(req LoginRequest) (*LoginResponse, error)
	OauthLoginAndRegister(req OAuthRequest) (*LoginResponse, error)
	Logout(refreshToken string) error
	RefreshAccessToken(token string) (string, error)
	GetProfile(userId string) (*ProfileResponse, error)
	UpdateProfile(userId string, req UpdateProfileRequest) (*ProfileResponse, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(newPassword, resetToken string) error
	InitPasswordChange(userId string, req InitPasswordChangeRequest) error
	VerifyPasswordChange(userId string, req VerifyPasswordChangeRequest) error
}

var oauthHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			// Prefer IPv4 (tcp4) to prevent "network is unreachable" on systems/containers without an IPv6 route
			conn, err := dialer.DialContext(ctx, "tcp4", addr)
			if err == nil {
				return conn, nil
			}
			return dialer.DialContext(ctx, network, addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

type AuthServiceImpl struct {
	repo AuthRepository
	rdb  *redis.Client
}

func NewAuthService(repo AuthRepository, rdb *redis.Client) AuthService {
	return &AuthServiceImpl{repo: repo, rdb: rdb}
}

func (s *AuthServiceImpl) Register(req RegisterRequest) (string, error) {
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
	newUser.AuthProvider = "local"
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

	subscription := model.Subscription{
		Plan:   "hobby",
		Status: "active",
	}

	if err := s.repo.CreateUserWithDefaultOrg(&newUser, &newOrg, &orgOwner, &subscription); err != nil {
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
	if user.PasswordHash == nil && user.AuthProvider != "local" {
		return nil, errors.New("Invalid email or password")
	}

	// Validation for Password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))

	if err != nil {
		return nil, errors.New("Password Incorrent, Please try again")
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

	accessToken, err := jwt.GenerateToken(user.ID.String(), expiryAccess)

	if err != nil {
		return nil, errors.New("Failed to generate access token")
	}

	refreshToken, err := jwt.GenerateToken(user.ID.String(), expiryRefresh)

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

	if err := s.repo.CreateRefreshToken(&createRefreshToken); err != nil {
		return nil, errors.New("Failed to store refresh token")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) OauthLoginAndRegister(req OAuthRequest) (*LoginResponse, error) {
	var conf *oauth2.Config
	var userInfoURL string

	if req.Provider == "google" {
		clientID := os.Getenv("PUBLIC_GOOGLE_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
		redirectURL := os.Getenv("PUBLIC_GOOGLE_REDIRECT_URL")

		if clientID == "" || clientSecret == "" {
			return nil, errors.New("Google OAuth is not configured: missing client ID or client secret")
		}

		conf = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}

		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	} else {
		return nil, errors.New("Unsupported Provider")
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, oauthHTTPClient)

	// Trade Authorization code with Token
	token, err := conf.Exchange(ctx, req.Code)

	if err != nil {
		return nil, errors.New("Failed to exchange token: " + err.Error())
	}

	// Get profile data from user
	response, err := oauthHTTPClient.Get(userInfoURL + token.AccessToken)

	if err != nil {
		return nil, errors.New("Failed to get user info")
	}
	defer response.Body.Close()

	userData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("Failed to read user data")
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(userData, &userInfo); err != nil {
		return nil, errors.New("Failed to parce user data")
	}

	email, ok := userInfo["email"].(string)
	if !ok {
		return nil, errors.New("Email not found")
	}

	user, err := s.repo.FindUserByEmail(email)

	// If user not registered yet
	if err != nil {
		name, _ := userInfo["name"].(string)
		picture, _ := userInfo["picture"].(string)

		newUser := model.User{
			Name:         name,
			Email:        email,
			AuthProvider: req.Provider,
			AvatarURL:    picture,
		}

		orgName := name + "Workspace"

		slug := strings.ToLower(name)
		uniqueSuffix := strconv.FormatInt(time.Now().UnixNano()%10000, 10)
		orgSlug := slug + "-workspace-" + uniqueSuffix

		newOrg := model.Organization{
			Name: orgName,
			Slug: orgSlug,
		}

		orgOwner := model.OrganizationMember{
			Role: "owner",
		}

		subscription := model.Subscription{
			Plan:   "hobby",
			Status: "active",
		}

		if err := s.repo.CreateUserWithDefaultOrg(&newUser, &newOrg, &orgOwner, &subscription); err != nil {
			return nil, errors.New("Failed to register user")
		}

		user = &newUser
	}

	// Generate token
	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))
	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	expiryRefresh, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))
	if err != nil {
		expiryRefresh = 168 * time.Hour
	}

	accessToken, err := jwt.GenerateToken(user.ID.String(), expiryAccess)
	if err != nil {
		return nil, errors.New("Something went wrong, please try again")
	}

	refreshToken, err := jwt.GenerateToken(user.ID.String(), expiryRefresh)
	if err != nil {
		return nil, errors.New("Something went wrong, please try again")
	}

	hash := sha256.New()
	hash.Write([]byte(refreshToken))
	hashedRefreshToken := hex.EncodeToString(hash.Sum(nil))

	createRefreshToken := model.RefreshToken{
		UserID:    user.ID,
		Token:     string(hashedRefreshToken),
		ExpiresAt: time.Now().Add(expiryRefresh),
	}

	if err := s.repo.CreateRefreshToken(&createRefreshToken); err != nil {
		return nil, errors.New("Something went wrong, please try again")
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

func (s *AuthServiceImpl) RefreshAccessToken(token string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(token))
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	refreshToken, err := s.repo.FindRefreshTokenByToken(hashedToken)

	if err != nil {
		return "", errors.New("Refreshtoken not found")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return "", errors.New("Refresh token expired")
	}

	expiryAccess, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))

	if err != nil {
		expiryAccess = 15 * time.Minute
	}

	AccessToken, err := jwt.GenerateToken(refreshToken.UserID.String(), expiryAccess)

	if err != nil {
		return "", errors.New("Failed to generate access token")
	}

	return AccessToken, nil
}

func (s *AuthServiceImpl) GetProfile(userId string) (*ProfileResponse, error) {
	user, err := s.repo.GetUserById(userId)

	if err != nil {
		return nil, errors.New("User does't exist")
	}

	response := ProfileResponse{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		AuthProvider: user.AuthProvider,
	}

	return &response, nil
}

func (s *AuthServiceImpl) UpdateProfile(userId string, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.repo.UpdateProfile(userId, req.Name)
	if err != nil {
		return nil, errors.New("User doesn't exist")
	}

	response := ProfileResponse{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		AuthProvider: user.AuthProvider,
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

	return nil
}

func (s *AuthServiceImpl) InitPasswordChange(userId string, req InitPasswordChangeRequest) error {
	user, err := s.repo.GetUserById(userId)
	if err != nil {
		return errors.New("User not found")
	}

	if user.PasswordHash == nil {
		return errors.New("Cannot change password for OAuth user")
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
	key := fmt.Sprintf("password_change_code:%s", userId)

	hashedCode, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return errors.New("Verification code expired or not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(req.Code)); err != nil {
		return errors.New("Invalid verification code")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		return errors.New("Internal Server Error")
	}

	if err := s.repo.CreateNewPassword(userId, string(hashedPassword)); err != nil {
		return errors.New("Failed to change password")
	}
	s.rdb.Del(ctx, key)

	return nil
}
