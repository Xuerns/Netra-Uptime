package utils

import (
	"os"
	"testing"
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
			target:  "http://127.0.0.1:8080/health",
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
			target:       "http://127.0.0.1:8080",
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
