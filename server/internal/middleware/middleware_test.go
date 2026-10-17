package middleware

import (
	"netra-uptime/internal/apikey"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/jwt"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-1234567890")
	}
}

// mockAPIKeyService for testing APIKeyAuthMiddleware
type mockAPIKeyService struct {
	validateKeyFunc func(key string) (*model.APIKey, error)
}

func (m *mockAPIKeyService) CreateKey(orgID, userID string, req apikey.CreateAPIKeyRequest, ip string) (*apikey.CreatedAPIKeyResponse, error) {
	return nil, nil
}
func (m *mockAPIKeyService) ListKeys(orgID string) ([]apikey.APIKeyResponse, error) {
	return nil, nil
}
func (m *mockAPIKeyService) RevokeKey(orgID, userID, keyID string, ip string) error {
	return nil
}
func (m *mockAPIKeyService) ValidateKey(plaintextKey string) (*model.APIKey, error) {
	if m.validateKeyFunc != nil {
		return m.validateKeyFunc(plaintextKey)
	}
	return nil, errors.New("invalid API key")
}
func (m *mockAPIKeyService) TouchLastUsed(keyID string, ip string) error {
	return nil
}

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	token, _ := jwt.GenerateToken("user-123", -10*time.Minute)

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "not-a-valid-jwt-token",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken_SetsUserID(t *testing.T) {
	token, _ := jwt.GenerateToken("user-valid-999", 10*time.Minute)

	var extractedUserID string
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		val, _ := c.Get("userID")
		extractedUserID = val.(string)
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: token,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if extractedUserID != "user-valid-999" {
		t.Errorf("expected userID user-valid-999, got %s", extractedUserID)
	}
}

func TestRequireRole_RoleAllowed(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationRole", "owner")
		c.Next()
	})
	r.Use(RequireRole("owner", "admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_RoleDenied(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationRole", "member")
		c.Next()
	})
	r.Use(RequireRole("owner", "admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_NoRoleInContext(t *testing.T) {
	r := gin.New()
	r.Use(RequireRole("owner", "admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestTenantMiddleware_MissingUserID(t *testing.T) {
	r := gin.New()
	r.Use(TenantMiddleware(nil))
	r.GET("/org/:orgId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/org/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing userID, got %d", w.Code)
	}
}

func TestTenantMiddleware_MissingOrgID(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "some-user-id")
		c.Next()
	})
	r.Use(TenantMiddleware(nil))
	r.GET("/org", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/org", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty orgId, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_NoHeader(t *testing.T) {
	mock := &mockAPIKeyService{}
	r := gin.New()
	r.Use(APIKeyAuthMiddleware(mock, nil))
	r.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_InvalidKey(t *testing.T) {
	mock := &mockAPIKeyService{
		validateKeyFunc: func(key string) (*model.APIKey, error) {
			return nil, errors.New("invalid API key")
		},
	}
	r := gin.New()
	r.Use(APIKeyAuthMiddleware(mock, nil))
	r.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("X-API-Key", "bad-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_ValidKey_XAPIKey(t *testing.T) {
	orgID := uuid.New()
	keyID := uuid.New()
	mock := &mockAPIKeyService{
		validateKeyFunc: func(key string) (*model.APIKey, error) {
			return &model.APIKey{
				ID:             keyID,
				OrganizationID: orgID,
				Permissions:    "write",
			}, nil
		},
	}

	var extractedOrgID string
	var extractedPerms string

	r := gin.New()
	r.Use(APIKeyAuthMiddleware(mock, nil))
	r.GET("/public", func(c *gin.Context) {
		val1, _ := c.Get("org_id")
		extractedOrgID = val1.(string)
		val2, _ := c.Get("api_key_permissions")
		extractedPerms = val2.(string)
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("X-API-Key", "pw_live_testkey")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if extractedOrgID != orgID.String() {
		t.Errorf("expected orgID %s, got %s", orgID.String(), extractedOrgID)
	}
	if extractedPerms != "write" {
		t.Errorf("expected perms write, got %s", extractedPerms)
	}
}

func TestAPIKeyAuthMiddleware_ValidKey_BearerHeader(t *testing.T) {
	orgID := uuid.New()
	mock := &mockAPIKeyService{
		validateKeyFunc: func(key string) (*model.APIKey, error) {
			return &model.APIKey{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Permissions:    "read",
			}, nil
		},
	}

	r := gin.New()
	r.Use(APIKeyAuthMiddleware(mock, nil))
	r.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer pw_live_bearerkey")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAPIKeyPermission_MissingPerms(t *testing.T) {
	r := gin.New()
	r.Use(RequireAPIKeyPermission("read"))
	r.GET("/res", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/res", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAPIKeyPermission_WriteRequired_HasRead(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key_permissions", "read")
		c.Next()
	})
	r.Use(RequireAPIKeyPermission("write"))
	r.POST("/res", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodPost, "/res", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAPIKeyPermission_WriteRequired_HasWrite(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key_permissions", "write")
		c.Next()
	})
	r.Use(RequireAPIKeyPermission("write"))
	r.POST("/res", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodPost, "/res", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAPIKeyPermission_ReadRequired_HasWrite(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key_permissions", "write")
		c.Next()
	})
	r.Use(RequireAPIKeyPermission("read"))
	r.GET("/res", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/res", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
