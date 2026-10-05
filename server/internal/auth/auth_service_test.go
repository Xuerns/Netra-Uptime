package auth

import (
	"PulseWatch/internal/model"
	"PulseWatch/pkg/utils"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	gin.SetMode(gin.TestMode)
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-1234567890")
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// Mock Repository
type mockAuthRepo struct {
	findUserByEmailFunc          func(email string) (*model.User, error)
	createRefreshTokenFunc       func(refreshToken *model.RefreshToken) error
	revokeRefreshTokenFunc       func(hashedToken string) error
	findRefreshTokenByTokenFunc  func(hashedToken string) (*model.RefreshToken, error)
	createUserWithDefaultOrgFunc func(user *model.User, org *model.Organization, orgMember *model.OrganizationMember, subscription *model.Subscription) error
	getUserByIdFunc              func(userId string) (*model.User, error)
	updateProfileFunc            func(userId string, name string) (*model.User, error)
	createResetTokenFunc         func(token *model.PasswordResetToken) error
	createNewPasswordFunc        func(userId, password string) error
	deleteResetTokenFunc         func(userId, token string) (*model.PasswordResetToken, error)
	getUserByResetTokenFunc      func(token string) (string, error)
}

func (m *mockAuthRepo) FindUserByEmail(email string) (*model.User, error) {
	if m.findUserByEmailFunc != nil {
		return m.findUserByEmailFunc(email)
	}
	return nil, errors.New("user not found")
}

func (m *mockAuthRepo) CreateRefreshToken(refreshToken *model.RefreshToken) error {
	if m.createRefreshTokenFunc != nil {
		return m.createRefreshTokenFunc(refreshToken)
	}
	return nil
}

func (m *mockAuthRepo) RevokeRefreshToken(hashedToken string) error {
	if m.revokeRefreshTokenFunc != nil {
		return m.revokeRefreshTokenFunc(hashedToken)
	}
	return nil
}

func (m *mockAuthRepo) FindRefreshTokenByToken(hashedToken string) (*model.RefreshToken, error) {
	if m.findRefreshTokenByTokenFunc != nil {
		return m.findRefreshTokenByTokenFunc(hashedToken)
	}
	return nil, errors.New("token not found")
}

func (m *mockAuthRepo) CreateUserWithDefaultOrg(user *model.User, org *model.Organization, orgMember *model.OrganizationMember, subscription *model.Subscription) error {
	if m.createUserWithDefaultOrgFunc != nil {
		return m.createUserWithDefaultOrgFunc(user, org, orgMember, subscription)
	}
	return nil
}

func (m *mockAuthRepo) GetUserById(userId string) (*model.User, error) {
	if m.getUserByIdFunc != nil {
		return m.getUserByIdFunc(userId)
	}
	return nil, errors.New("user not found")
}

func (m *mockAuthRepo) UpdateProfile(userId string, name string) (*model.User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(userId, name)
	}
	return nil, errors.New("user not found")
}

func (m *mockAuthRepo) CreateResetToken(token *model.PasswordResetToken) error {
	if m.createResetTokenFunc != nil {
		return m.createResetTokenFunc(token)
	}
	return nil
}

func (m *mockAuthRepo) CreateNewPassword(userId, password string) error {
	if m.createNewPasswordFunc != nil {
		return m.createNewPasswordFunc(userId, password)
	}
	return nil
}

func (m *mockAuthRepo) DeleteResetToken(userId, token string) (*model.PasswordResetToken, error) {
	if m.deleteResetTokenFunc != nil {
		return m.deleteResetTokenFunc(userId, token)
	}
	return nil, nil
}

func (m *mockAuthRepo) GetUserByResetToken(token string) (string, error) {
	if m.getUserByResetTokenFunc != nil {
		return m.getUserByResetTokenFunc(token)
	}
	return "", errors.New("reset token not found")
}

// ====================== SERVICE TESTS ======================

func TestRegister_Success(t *testing.T) {
	created := false
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return nil, errors.New("not found")
		},
		createUserWithDefaultOrgFunc: func(u *model.User, o *model.Organization, m *model.OrganizationMember, s *model.Subscription) error {
			created = true
			return nil
		},
	}
	svc := NewAuthService(repo, nil)

	msg, err := svc.Register(RegisterRequest{
		Name:     "Alice Doe",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected user to be created")
	}
	if msg != "User registered successfully" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestRegister_UserAlreadyExist(t *testing.T) {
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return &model.User{Email: email}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.Register(RegisterRequest{
		Name:     "Alice Doe",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for existing user, got nil")
	}
	if err.Error() != "User Already Exist" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), 10)
	hashedPassStr := string(hashedPassword)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: &hashedPassStr,
		AuthProvider: "local",
	}

	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return user, nil
		},
		createRefreshTokenFunc: func(refreshToken *model.RefreshToken) error {
			return nil
		},
	}
	svc := NewAuthService(repo, nil)

	resp, err := svc.Login(LoginRequest{
		Email:    "user@example.com",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Errorf("expected tokens to be generated, got empty")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.Login(LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), 10)
	hashedPassStr := string(hashedPassword)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: &hashedPassStr,
		AuthProvider: "local",
	}

	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return user, nil
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.Login(LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_OAuthUser_NoPassword(t *testing.T) {
	user := &model.User{
		ID:           uuid.New(),
		Email:        "oauth@example.com",
		PasswordHash: nil,
		AuthProvider: "google",
	}

	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return user, nil
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.Login(LoginRequest{
		Email:    "oauth@example.com",
		Password: "some-password",
	})
	if err == nil {
		t.Fatal("expected error for OAuth user logging in with password, got nil")
	}
}

func TestLogout_Success(t *testing.T) {
	revoked := false
	repo := &mockAuthRepo{
		revokeRefreshTokenFunc: func(hashedToken string) error {
			revoked = true
			return nil
		},
	}
	svc := NewAuthService(repo, nil)

	err := svc.Logout("some-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Fatal("expected refresh token to be revoked")
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{
		findRefreshTokenByTokenFunc: func(hashedToken string) (*model.RefreshToken, error) {
			return &model.RefreshToken{
				UserID:    userID,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	token, err := svc.RefreshAccessToken("valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected new access token, got empty string")
	}
}

func TestRefreshAccessToken_Expired(t *testing.T) {
	repo := &mockAuthRepo{
		findRefreshTokenByTokenFunc: func(hashedToken string) (*model.RefreshToken, error) {
			return &model.RefreshToken{
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.RefreshAccessToken("expired-token")
	if err == nil {
		t.Fatal("expected error for expired refresh token, got nil")
	}
}

func TestGetProfile_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{
		getUserByIdFunc: func(id string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Name:         "John Doe",
				Email:        "john@example.com",
				AuthProvider: "local",
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	profile, err := svc.GetProfile(userID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Name != "John Doe" {
		t.Errorf("expected John Doe, got %s", profile.Name)
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{
		getUserByIdFunc: func(id string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.GetProfile(uuid.New().String())
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{
		updateProfileFunc: func(id string, name string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Name:         name,
				Email:        "john@example.com",
				AuthProvider: "local",
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	profile, err := svc.UpdateProfile(userID.String(), UpdateProfileRequest{Name: "Jane Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Name != "Jane Doe" {
		t.Errorf("expected Jane Doe, got %s", profile.Name)
	}
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{
		updateProfileFunc: func(id string, name string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewAuthService(repo, nil)

	_, err := svc.UpdateProfile(uuid.New().String(), UpdateProfileRequest{Name: "Jane Doe"})
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
}

func TestForgotPassword_UserNotFound_NoError(t *testing.T) {
	// Anti-enumeration test: user not found should return empty URL and no error!
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewAuthService(repo, nil)

	url, err := svc.ForgotPassword("nonexistent@example.com")
	if err != nil {
		t.Fatalf("expected nil error for anti-enumeration, got: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty URL, got: %s", url)
	}
}

func TestForgotPassword_Success(t *testing.T) {
	os.Setenv("ALLOWED_ORIGIN", "https://app.pulsewatch.dev")
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return &model.User{ID: uuid.New(), Email: email}, nil
		},
		createResetTokenFunc: func(token *model.PasswordResetToken) error {
			return nil
		},
	}
	svc := NewAuthService(repo, nil)

	resetURL, err := svc.ForgotPassword("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resetURL, "https://app.pulsewatch.dev/auth/forgot?token=") {
		t.Errorf("unexpected reset URL: %s", resetURL)
	}
}

func TestResetPassword_Success(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID}

	repo := &mockAuthRepo{
		getUserByResetTokenFunc: func(token string) (string, error) {
			return userID.String(), nil
		},
		getUserByIdFunc: func(id string) (*model.User, error) {
			return user, nil
		},
		deleteResetTokenFunc: func(userId, token string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				UserID:    userID,
				ExpiredAt: time.Now().Add(10 * time.Minute),
			}, nil
		},
		createNewPasswordFunc: func(userId, password string) error {
			return nil
		},
	}
	svc := NewAuthService(repo, nil)

	err := svc.ResetPassword("new-secure-password", "raw-token-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID}

	repo := &mockAuthRepo{
		getUserByResetTokenFunc: func(token string) (string, error) {
			return userID.String(), nil
		},
		getUserByIdFunc: func(id string) (*model.User, error) {
			return user, nil
		},
		deleteResetTokenFunc: func(userId, token string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				UserID:    userID,
				ExpiredAt: time.Now().Add(-10 * time.Minute), // Expired!
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	err := svc.ResetPassword("new-password", "expired-token")
	if err == nil {
		t.Fatal("expected error for expired reset token, got nil")
	}
}

func TestInitPasswordChange_OAuthUser(t *testing.T) {
	repo := &mockAuthRepo{
		getUserByIdFunc: func(id string) (*model.User, error) {
			return &model.User{
				ID:           uuid.New(),
				PasswordHash: nil, // OAuth user has no password
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	err := svc.InitPasswordChange(uuid.New().String(), InitPasswordChangeRequest{
		CurrentPassword: "password",
	})
	if err == nil {
		t.Fatal("expected error for OAuth user changing password, got nil")
	}
	if err.Error() != "Cannot change password for OAuth user" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInitPasswordChange_WrongCurrentPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), 10)
	hStr := string(hashedPassword)
	repo := &mockAuthRepo{
		getUserByIdFunc: func(id string) (*model.User, error) {
			return &model.User{
				ID:           uuid.New(),
				PasswordHash: &hStr,
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)

	err := svc.InitPasswordChange(uuid.New().String(), InitPasswordChangeRequest{
		CurrentPassword: "wrong-pass",
	})
	if err == nil {
		t.Fatal("expected error for incorrect current password, got nil")
	}
	if err.Error() != "Incorrect current password" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ====================== HANDLER TESTS ======================

func TestAuthHandler_Register_Success(t *testing.T) {
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return nil, errors.New("not found")
		},
		createUserWithDefaultOrgFunc: func(u *model.User, o *model.Organization, m *model.OrganizationMember, s *model.Subscription) error {
			return nil
		},
	}
	svc := NewAuthService(repo, nil)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/register", handler.Register)

	body := RegisterRequest{
		Name:     "Bob Dylan",
		Email:    "bob@example.com",
		Password: "password123",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), 10)
	hStr := string(hashedPassword)
	repo := &mockAuthRepo{
		findUserByEmailFunc: func(email string) (*model.User, error) {
			return &model.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: &hStr,
				AuthProvider: "local",
			}, nil
		},
		createRefreshTokenFunc: func(refreshToken *model.RefreshToken) error {
			return nil
		},
	}
	svc := NewAuthService(repo, nil)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.POST("/auth/login", handler.Login)

	body := LoginRequest{
		Email:    "bob@example.com",
		Password: "secret123",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	// Verify cookies were set
	cookies := w.Result().Cookies()
	foundAccess := false
	for _, c := range cookies {
		if c.Name == "access_token" {
			foundAccess = true
		}
	}
	if !foundAccess {
		t.Errorf("expected access_token cookie to be set")
	}
}

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{
		getUserByIdFunc: func(id string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Name:         "John Profile",
				Email:        "john@profile.com",
				AuthProvider: "local",
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID.String())
		c.Next()
	})
	r.GET("/auth/profile", handler.GetProfile)

	req, _ := http.NewRequest(http.MethodGet, "/auth/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_UpdateProfile_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{
		updateProfileFunc: func(id string, name string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Name:         name,
				Email:        "john@profile.com",
				AuthProvider: "local",
			}, nil
		},
	}
	svc := NewAuthService(repo, nil)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID.String())
		c.Next()
	})
	r.PUT("/auth/profile", handler.UpdateProfile)

	body := `{"name":"Updated Name"}`
	req, _ := http.NewRequest(http.MethodPut, "/auth/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_UpdateProfile_ValidationError(t *testing.T) {
	userID := uuid.New()
	repo := &mockAuthRepo{}
	svc := NewAuthService(repo, nil)
	handler := NewAuthHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID.String())
		c.Next()
	})
	r.PUT("/auth/profile", handler.UpdateProfile)

	// Name too short (< 3 chars)
	body := `{"name":"ab"}`
	req, _ := http.NewRequest(http.MethodPut, "/auth/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short name, got %d", w.Code)
	}
}
