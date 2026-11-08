package checker

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"netra-uptime/pkg/utils"
	"strings"
	"testing"
	"time"
)

func startPrivateTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	utils.SetDefaultAllowPrivateNetwork(true)
	t.Cleanup(func() {
		utils.SetDefaultAllowPrivateNetwork(false)
	})

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("failed to list interface addresses: %v", err)
	}
	var privateIP string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil && ip4.IsPrivate() {
				privateIP = ip4.String()
				break
			}
		}
	}
	if privateIP == "" {
		t.Skip("no private network interface available for test server")
	}

	ln, err := net.Listen("tcp", privateIP+":0")
	if err != nil {
		t.Fatalf("failed to listen on %s:0: %v", privateIP, err)
	}
	srv := httptest.NewUnstartedServer(handler)
	_ = srv.Listener.Close()
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestCheckHTTPWithOptions_MethodHeadersBodyAndSnippet(t *testing.T) {
	var receivedMethod string
	var receivedToken string
	var receivedContentType string
	var receivedBody string

	srv := startPrivateTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedToken = r.Header.Get("X-Custom-Token")
		receivedContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)

		if r.URL.Path == "/fail-502" {
			w.WriteHeader(http.StatusBadGateway)
			longSuffix := strings.Repeat("A", 700)
			_, _ = w.Write([]byte("{\"error\":\"upstream database connection failed\"}\x00\x01\x02\n" + longSuffix))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"created"}`))
	}))

	// 1. Test POST with custom headers and JSON body
	code, dur, snippet, err := CheckHTTPWithOptions(srv.URL+"/create", 5*time.Second, HTTPCheckOptions{
		Method:         "POST",
		HeadersJSON:    `{"X-Custom-Token":"secret123"}`,
		Body:           `{"ping":"pong"}`,
		ExpectedStatus: 201,
	})
	if err != nil {
		t.Fatalf("CheckHTTPWithOptions unexpected error: %v", err)
	}
	if code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", code)
	}
	if dur < 0 {
		t.Errorf("expected non-negative duration, got %v", dur)
	}
	if snippet != "" {
		t.Errorf("expected empty snippet for 2xx response, got %q", snippet)
	}
	if receivedMethod != "POST" {
		t.Errorf("expected server to receive POST, got %s", receivedMethod)
	}
	if receivedToken != "secret123" {
		t.Errorf("expected X-Custom-Token 'secret123', got %q", receivedToken)
	}
	if receivedContentType != "application/json" {
		t.Errorf("expected default Content-Type 'application/json', got %q", receivedContentType)
	}
	if receivedBody != `{"ping":"pong"}` {
		t.Errorf("expected body '{\"ping\":\"pong\"}', got %q", receivedBody)
	}

	// 2. Test 502 error response snippet capture, sanitization, and 512-byte cap
	codeErr, _, snippetErr, err := CheckHTTPWithOptions(srv.URL+"/fail-502", 5*time.Second, HTTPCheckOptions{
		Method:         "GET",
		ExpectedStatus: 200,
	})
	if err != nil {
		t.Fatalf("CheckHTTPWithOptions on 502 unexpected network error: %v", err)
	}
	if codeErr != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", codeErr)
	}
	if !strings.Contains(snippetErr, `{"error":"upstream database connection failed"}`) {
		t.Errorf("expected snippet to contain error JSON, got %q", snippetErr)
	}
	if strings.ContainsAny(snippetErr, "\x00\x01\x02") {
		t.Errorf("expected non-printable control bytes to be stripped from snippet")
	}
	if len(snippetErr) > 512 {
		t.Errorf("expected snippet length <= 512, got %d", len(snippetErr))
	}
}

func TestCheckHTTP_SSRFProtection(t *testing.T) {
	targets := []string{
		"http://127.0.0.1:8080/health",
		"http://169.254.169.254/latest/meta-data/",
		"http://localhost:3000",
		"http://postgres:5432",
		"http://10.0.0.1:80",
		"http://192.168.1.1:80",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			_, _, err := CheckHTTP(target, 500*time.Millisecond, 200)
			if err == nil {
				t.Fatalf("CheckHTTP(%q) expected error for restricted target, got nil", target)
			}
		})
	}
}

func TestCheckTCP_SSRFProtection(t *testing.T) {
	targets := []string{
		"127.0.0.1:8080",
		"169.254.169.254:80",
		"localhost:3000",
		"postgres:5432",
		"10.0.0.1:80",
		"192.168.1.1:80",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			_, err := CheckTCP(target, 500*time.Millisecond)
			if err == nil {
				t.Fatalf("CheckTCP(%q) expected error for restricted target, got nil", target)
			}
		})
	}
}

func TestCheckPing_SSRFProtection(t *testing.T) {
	targets := []string{
		"127.0.0.1",
		"169.254.169.254",
		"localhost",
		"postgres",
		"10.0.0.1",
		"192.168.1.1",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			_, err := CheckPing(target, 500*time.Millisecond)
			if err == nil {
				t.Fatalf("CheckPing(%q) expected error for restricted target, got nil", target)
			}
		})
	}
}

func TestCheckSSLExpiry_SSRFProtection(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantMsg string
	}{
		{
			name:    "Block localhost",
			target:  "localhost:443",
			wantMsg: "Restricted Target",
		},
		{
			name:    "Block postgres container host",
			target:  "postgres:443",
			wantMsg: "Restricted Target",
		},
		{
			name:    "Block IP address target",
			target:  "127.0.0.1:443",
			wantMsg: "Target is IP address",
		},
		{
			name:    "Block metadata IP address target",
			target:  "169.254.169.254:443",
			wantMsg: "Target is IP address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckSSLExpiry(tt.target, 500*time.Millisecond)
			if res.Status != "error" {
				t.Fatalf("CheckSSLExpiry(%q) status = %q, want 'error'", tt.target, res.Status)
			}
			if !strings.Contains(res.ErrorMsg, tt.wantMsg) {
				t.Fatalf("CheckSSLExpiry(%q) errorMsg = %q, want it to contain %q", tt.target, res.ErrorMsg, tt.wantMsg)
			}
		})
	}
}
