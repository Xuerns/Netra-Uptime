package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"PulseWatch/config"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck_DegradedWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewHealthHandler(nil, nil, &config.Config{AppMode: "selfhosted"})
	r.GET("/healthz", h.Check)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status degraded, got %s", resp.Status)
	}
	if resp.Database != "unreachable" {
		t.Errorf("expected database unreachable, got %s", resp.Database)
	}
}
