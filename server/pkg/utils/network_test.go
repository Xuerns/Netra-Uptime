package utils

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestValidateTargetHost(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		allowPrivate bool
		wantErr      bool
	}{
		{
			name:    "Block loopback IP",
			target:  "http://127.0.0.1:27346/health",
			wantErr: true,
		},
		{
			name:    "Block cloud metadata IP",
			target:  "http://169.254.169.254/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "Block localhost hostname",
			target:  "http://localhost:3000",
			wantErr: true,
		},
		{
			name:    "Block postgres container host",
			target:  "tcp://postgres:5432",
			wantErr: true,
		},
		{
			name:    "Block redis container host",
			target:  "tcp://redis:6379",
			wantErr: true,
		},
		{
			name:    "Block private IP 10.x.x.x by default",
			target:  "http://10.0.0.1:80",
			wantErr: true,
		},
		{
			name:    "Block private IP 192.168.x.x by default",
			target:  "http://192.168.1.1:80",
			wantErr: true,
		},
		{
			name:         "Allow private IP when ALLOW_PRIVATE_NETWORK_MONITORING is true",
			target:       "http://192.168.1.1:80",
			allowPrivate: true,
			wantErr:      false,
		},
		{
			name:         "Cloud metadata still blocked even when allowPrivate is true",
			target:       "http://169.254.169.254/secret",
			allowPrivate: true,
			wantErr:      true,
		},
		{
			name:         "Loopback still blocked even when allowPrivate is true",
			target:       "http://127.0.0.1:27346",
			allowPrivate: true,
			wantErr:      true,
		},
		{
			name:    "Allow public IP",
			target:  "http://93.184.215.14:80", // example.com IP
			wantErr: false,
		},
		{
			name:    "Allow public hostname",
			target:  "https://example.com",
			wantErr: false,
		},
		{
			name:    "Fail-closed on unresolvable hostname",
			target:  "http://nonexistent-host-that-does-not-exist.invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.allowPrivate {
				os.Setenv("ALLOW_PRIVATE_NETWORK_MONITORING", "true")
			} else {
				os.Unsetenv("ALLOW_PRIVATE_NETWORK_MONITORING")
			}
			defer os.Unsetenv("ALLOW_PRIVATE_NETWORK_MONITORING")

			err := ValidateTargetHost(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTargetHost(%q) error = %v, wantErr = %v", tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestResolveSafeIPs(t *testing.T) {
	ctx := context.Background()

	// 1. Loopback should fail
	if _, err := ResolveSafeIPs(ctx, "127.0.0.1"); err == nil {
		t.Errorf("ResolveSafeIPs('127.0.0.1') expected error, got nil")
	}

	// 2. Metadata IP should fail
	if _, err := ResolveSafeIPs(ctx, "169.254.169.254"); err == nil {
		t.Errorf("ResolveSafeIPs('169.254.169.254') expected error, got nil")
	}

	// 3. Blocked hostname should fail
	if _, err := ResolveSafeIPs(ctx, "localhost"); err == nil {
		t.Errorf("ResolveSafeIPs('localhost') expected error, got nil")
	}

	// 4. Invalid hostname should fail-closed
	if _, err := ResolveSafeIPs(ctx, "unresolvable-domain.invalid"); err == nil {
		t.Errorf("ResolveSafeIPs('unresolvable-domain.invalid') expected error, got nil")
	}

	// 5. Public IP should succeed
	ips, err := ResolveSafeIPs(ctx, "93.184.215.14")
	if err != nil {
		t.Errorf("ResolveSafeIPs('93.184.215.14') unexpected error: %v", err)
	}
	if len(ips) == 0 {
		t.Errorf("ResolveSafeIPs('93.184.215.14') returned no IPs")
	}
}

func TestSafeDialTimeout_BlocksRestricted(t *testing.T) {
	timeout := 100 * time.Millisecond

	// 1. Loopback IP dial must be blocked
	_, err := SafeDialTimeout("tcp", "127.0.0.1:8080", timeout)
	if err == nil {
		t.Errorf("SafeDialTimeout('tcp', '127.0.0.1:8080') expected error, got nil")
	}

	// 2. Cloud metadata IP dial must be blocked
	_, err = SafeDialTimeout("tcp", "169.254.169.254:80", timeout)
	if err == nil {
		t.Errorf("SafeDialTimeout('tcp', '169.254.169.254:80') expected error, got nil")
	}

	// 3. Blocked hostname dial must be blocked
	_, err = SafeDialTimeout("tcp", "localhost:3000", timeout)
	if err == nil {
		t.Errorf("SafeDialTimeout('tcp', 'localhost:3000') expected error, got nil")
	}

	// 4. Container hostname dial must be blocked
	_, err = SafeDialTimeout("tcp", "postgres:5432", timeout)
	if err == nil {
		t.Errorf("SafeDialTimeout('tcp', 'postgres:5432') expected error, got nil")
	}

	// 5. Unresolvable host must fail-closed
	_, err = SafeDialTimeout("tcp", "does-not-exist.invalid:80", timeout)
	if err == nil {
		t.Errorf("SafeDialTimeout with invalid domain expected error, got nil")
	}
}

func TestNewSafeHTTPClient_RedirectPolicy(t *testing.T) {
	client := NewSafeHTTPClient(2 * time.Second)
	if client.CheckRedirect == nil {
		t.Fatalf("expected CheckRedirect to be non-nil")
	}

	// Redirect to loopback must be rejected
	reqLoopback, _ := http.NewRequest("GET", "http://127.0.0.1:8080/admin", nil)
	if err := client.CheckRedirect(reqLoopback, nil); err == nil {
		t.Errorf("CheckRedirect to 127.0.0.1 expected error, got nil")
	}

	// Redirect to metadata must be rejected
	reqMeta, _ := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/", nil)
	if err := client.CheckRedirect(reqMeta, nil); err == nil {
		t.Errorf("CheckRedirect to metadata IP expected error, got nil")
	}

	// Redirect to internal host must be rejected
	reqInternal, _ := http.NewRequest("GET", "http://localhost:3000/internal", nil)
	if err := client.CheckRedirect(reqInternal, nil); err == nil {
		t.Errorf("CheckRedirect to localhost expected error, got nil")
	}

	// Max 5 redirects enforced
	targetURL, _ := url.Parse("https://example.com/target")
	reqValid := &http.Request{URL: targetURL}
	via5 := make([]*http.Request, 5)
	if err := client.CheckRedirect(reqValid, via5); err == nil {
		t.Errorf("CheckRedirect with 5 hops expected error, got nil")
	}
}
