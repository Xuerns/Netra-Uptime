package apikey

import (
	"PulseWatch/internal/model"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Mock Repository
type mockAPIKeyRepo struct {
	createFunc         func(key *model.APIKey) error
	getAllByOrgIDFunc  func(orgID string) ([]model.APIKey, error)
	getByHashFunc      func(keyHash string) (*model.APIKey, error)
	deleteFunc         func(orgID, keyID string) error
	updateLastUsedFunc func(keyID string, ip string) error
}

func (m *mockAPIKeyRepo) Create(key *model.APIKey) error {
	if m.createFunc != nil {
		return m.createFunc(key)
	}
	key.ID = uuid.New()
	key.CreatedAt = time.Now()
	return nil
}

func (m *mockAPIKeyRepo) GetAllByOrgID(orgID string) ([]model.APIKey, error) {
	if m.getAllByOrgIDFunc != nil {
		return m.getAllByOrgIDFunc(orgID)
	}
	return nil, nil
}

func (m *mockAPIKeyRepo) GetByHash(keyHash string) (*model.APIKey, error) {
	if m.getByHashFunc != nil {
		return m.getByHashFunc(keyHash)
	}
	return nil, errors.New("key not found")
}

func (m *mockAPIKeyRepo) Delete(orgID, keyID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgID, keyID)
	}
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(keyID string, ip string) error {
	if m.updateLastUsedFunc != nil {
		return m.updateLastUsedFunc(keyID, ip)
	}
	return nil
}

// ====================== SERVICE UNIT TESTS ======================

func TestHashAPIKey_Deterministic(t *testing.T) {
	h1 := HashAPIKey("test-key-123")
	h2 := HashAPIKey("test-key-123")
	h3 := HashAPIKey("different-key")

	if h1 != h2 {
		t.Errorf("expected hash to be deterministic, got %s != %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("expected different keys to have different hashes")
	}
}

func TestCreateKey_FormatStartsWithPwLive(t *testing.T) {
	repo := &mockAPIKeyRepo{}
	svc := NewAPIKeyService(repo, nil)

	orgID := uuid.New().String()
	userID := uuid.New().String()

	resp, err := svc.CreateKey(orgID, userID, CreateAPIKeyRequest{
		Name:        "Test Key",
		Permissions: "read",
	}, "127.0.0.1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(resp.PlaintextKey, "pw_live_") {
		t.Errorf("expected plaintext key to start with pw_live_, got %s", resp.PlaintextKey)
	}
	if !strings.HasPrefix(resp.KeyPrefix, "pw_live_") {
		t.Errorf("expected prefix to start with pw_live_, got %s", resp.KeyPrefix)
	}
}

func TestCreateKey_InvalidOrgUUID(t *testing.T) {
	svc := NewAPIKeyService(&mockAPIKeyRepo{}, nil)
	_, err := svc.CreateKey("not-a-uuid", uuid.New().String(), CreateAPIKeyRequest{
		Name:        "Test Key",
		Permissions: "read",
	}, "127.0.0.1")

	if err == nil {
		t.Fatal("expected error for invalid org uuid, got nil")
	}
}

func TestCreateKey_InvalidUserUUID(t *testing.T) {
	svc := NewAPIKeyService(&mockAPIKeyRepo{}, nil)
	_, err := svc.CreateKey(uuid.New().String(), "not-a-uuid", CreateAPIKeyRequest{
		Name:        "Test Key",
		Permissions: "read",
	}, "127.0.0.1")

	if err == nil {
		t.Fatal("expected error for invalid user uuid, got nil")
	}
}

func TestCreateKey_WithExpiry(t *testing.T) {
	repo := &mockAPIKeyRepo{}
	svc := NewAPIKeyService(repo, nil)

	days := 30
	resp, err := svc.CreateKey(uuid.New().String(), uuid.New().String(), CreateAPIKeyRequest{
		Name:        "Expiring Key",
		Permissions: "write",
		ExpiresIn:   &days,
	}, "127.0.0.1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpiresAt == nil {
		t.Fatal("expected expiresAt to be set, got nil")
	}
}

func TestListKeys_Success(t *testing.T) {
	orgID := uuid.New()
	repo := &mockAPIKeyRepo{
		getAllByOrgIDFunc: func(o string) ([]model.APIKey, error) {
			return []model.APIKey{
				{ID: uuid.New(), OrganizationID: orgID, Name: "Key 1", Permissions: "read", CreatedAt: time.Now()},
				{ID: uuid.New(), OrganizationID: orgID, Name: "Key 2", Permissions: "write", CreatedAt: time.Now()},
			}, nil
		},
	}
	svc := NewAPIKeyService(repo, nil)

	keys, err := svc.ListKeys(orgID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestRevokeKey_Success(t *testing.T) {
	revoked := false
	repo := &mockAPIKeyRepo{
		deleteFunc: func(orgID, keyID string) error {
			revoked = true
			return nil
		},
	}
	svc := NewAPIKeyService(repo, nil)

	err := svc.RevokeKey("org-1", "user-1", "key-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Fatal("expected delete to be called")
	}
}

func TestValidateKey_Success(t *testing.T) {
	keyID := uuid.New()
	orgID := uuid.New()
	plainKey := "pw_live_validkey123"
	keyHash := HashAPIKey(plainKey)

	repo := &mockAPIKeyRepo{
		getByHashFunc: func(h string) (*model.APIKey, error) {
			if h == keyHash {
				return &model.APIKey{
					ID:             keyID,
					OrganizationID: orgID,
					KeyHash:        keyHash,
					Permissions:    "read",
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	svc := NewAPIKeyService(repo, nil)

	key, err := svc.ValidateKey(plainKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != keyID {
		t.Errorf("expected key ID %s, got %s", keyID, key.ID)
	}
}

func TestValidateKey_Expired(t *testing.T) {
	plainKey := "pw_live_expiredkey123"
	keyHash := HashAPIKey(plainKey)
	past := time.Now().Add(-1 * time.Hour)

	repo := &mockAPIKeyRepo{
		getByHashFunc: func(h string) (*model.APIKey, error) {
			return &model.APIKey{
				ID:        uuid.New(),
				KeyHash:   keyHash,
				ExpiresAt: &past,
			}, nil
		},
	}
	svc := NewAPIKeyService(repo, nil)

	_, err := svc.ValidateKey(plainKey)
	if err == nil {
		t.Fatal("expected error for expired key, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTouchLastUsed(t *testing.T) {
	touched := false
	repo := &mockAPIKeyRepo{
		updateLastUsedFunc: func(keyID string, ip string) error {
			touched = true
			return nil
		},
	}
	svc := NewAPIKeyService(repo, nil)
	err := svc.TouchLastUsed("key-123", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !touched {
		t.Fatal("expected updateLastUsed to be called")
	}
}

// ====================== HANDLER TESTS ======================

func TestAPIKeyHandler_Create_Success(t *testing.T) {
	repo := &mockAPIKeyRepo{}
	svc := NewAPIKeyService(repo, nil)
	handler := NewAPIKeyHandler(svc)

	orgID := uuid.New().String()
	userID := uuid.New().String()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	r.POST("/organization/:orgId/api-keys", handler.Create)

	body := CreateAPIKeyRequest{
		Name:        "Production Key",
		Permissions: "read",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID+"/api-keys", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAPIKeyHandler_Create_MissingUserContext(t *testing.T) {
	svc := NewAPIKeyService(&mockAPIKeyRepo{}, nil)
	handler := NewAPIKeyHandler(svc)

	r := gin.New()
	r.POST("/organization/:orgId/api-keys", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/organization/some-org/api-keys", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyHandler_List_Success(t *testing.T) {
	svc := NewAPIKeyService(&mockAPIKeyRepo{}, nil)
	handler := NewAPIKeyHandler(svc)

	r := gin.New()
	r.GET("/organization/:orgId/api-keys", handler.List)

	req, _ := http.NewRequest(http.MethodGet, "/organization/some-org/api-keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Revoke_Success(t *testing.T) {
	svc := NewAPIKeyService(&mockAPIKeyRepo{}, nil)
	handler := NewAPIKeyHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		c.Next()
	})
	r.DELETE("/organization/:orgId/api-keys/:keyId", handler.Revoke)

	req, _ := http.NewRequest(http.MethodDelete, "/organization/some-org/api-keys/some-key", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
