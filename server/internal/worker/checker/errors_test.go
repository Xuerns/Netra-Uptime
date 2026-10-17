package checker

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type customTimeoutErr struct{}

func (e *customTimeoutErr) Error() string   { return "custom i/o timeout" }
func (e *customTimeoutErr) Timeout() bool   { return true }
func (e *customTimeoutErr) Temporary() bool { return true }

func TestClassifyNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: ErrorTypeNone,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: ErrorTypeTimeout,
		},
		{
			name:     "net.Error timeout",
			err:      &customTimeoutErr{},
			expected: ErrorTypeTimeout,
		},
		{
			name:     "http client timeout message",
			err:      errors.New("Get \"https://api.example.com\": Client.Timeout exceeded while awaiting headers"),
			expected: ErrorTypeTimeout,
		},
		{
			name:     "dns no such host",
			err:      &net.DNSError{Err: "no such host", Name: "nonexistent.example.internal"},
			expected: ErrorTypeDNS,
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp 127.0.0.1:8080: connect: connection refused"),
			expected: ErrorTypeConnectionRefused,
		},
		{
			name:     "tls cert invalid",
			err:      errors.New("tls: failed to verify certificate: x509: certificate has expired"),
			expected: ErrorTypeTLS,
		},
		{
			name:     "general error",
			err:      errors.New("unexpected EOF"),
			expected: ErrorTypeGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyNetworkError(tt.err)
			if got != tt.expected {
				t.Errorf("ClassifyNetworkError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatErrorMessage(t *testing.T) {
	timeoutErr := errors.New("context deadline exceeded")
	formatted := FormatErrorMessage(timeoutErr, 15*time.Second)
	expected := "Request timed out after 15s (Service did not respond in time)"
	if formatted != expected {
		t.Errorf("FormatErrorMessage() = %q, want %q", formatted, expected)
	}

	dnsErr := errors.New("dial tcp: lookup nonexisting.invalid: no such host")
	formattedDNS := FormatErrorMessage(dnsErr, 10*time.Second)
	if formattedDNS != "DNS resolution failed: Host could not be reached or resolved" {
		t.Errorf("FormatErrorMessage() for DNS = %q", formattedDNS)
	}
}
