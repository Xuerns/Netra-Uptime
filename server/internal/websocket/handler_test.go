package websocket

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	// Case 1: Empty Origin (CLI/curl) -> should allow
	req := httptest.NewRequest("GET", "/ws", nil)
	if !upgrader.CheckOrigin(req) {
		t.Fatal("expected empty origin to be allowed, got rejected")
	}

	// Case 2: Development mode (ALLOWED_ORIGIN empty) -> allow localhost
	_ = os.Unsetenv("ALLOWED_ORIGIN")
	reqLocal := httptest.NewRequest("GET", "/ws", nil)
	reqLocal.Header.Set("Origin", "http://localhost:5173")
	if !upgrader.CheckOrigin(reqLocal) {
		t.Fatal("expected localhost origin to be allowed in dev mode, got rejected")
	}

	reqEvilDev := httptest.NewRequest("GET", "/ws", nil)
	reqEvilDev.Header.Set("Origin", "https://evil-attacker.com")
	if upgrader.CheckOrigin(reqEvilDev) {
		t.Fatal("expected external origin to be rejected in dev mode, got allowed")
	}

	// Case 3: Production mode with ALLOWED_ORIGIN configured
	os.Setenv("ALLOWED_ORIGIN", "https://app.netra-uptime.dev, https://netra-uptime.dev")
	defer os.Unsetenv("ALLOWED_ORIGIN")

	reqValid := httptest.NewRequest("GET", "/ws", nil)
	reqValid.Header.Set("Origin", "https://app.netra-uptime.dev")
	if !upgrader.CheckOrigin(reqValid) {
		t.Fatal("expected ALLOWED_ORIGIN to be allowed, got rejected")
	}

	reqAttacker := httptest.NewRequest("GET", "/ws", nil)
	reqAttacker.Header.Set("Origin", "https://evil-attacker.com")
	if upgrader.CheckOrigin(reqAttacker) {
		t.Fatal("expected evil-attacker.com to be rejected, got allowed")
	}
}
