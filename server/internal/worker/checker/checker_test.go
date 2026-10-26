package checker

import (
	"strings"
	"testing"
	"time"
)

func TestCheckHTTP_SSRFProtection(t *testing.T) {
	targets := []string{
		"http://127.0.0.1:8080/health",
		"http://169.254.169.254/latest/meta-data/",
		"http://localhost:3000",
		"http://postgres:5432",
		"http://redis:6379",
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
		"redis:6379",
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
		"redis",
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
			name:    "Block redis container host",
			target:  "redis:443",
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
