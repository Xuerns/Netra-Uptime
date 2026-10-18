package checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type ErrorType string

const (
	ErrorTypeNone              ErrorType = ""
	ErrorTypeTimeout           ErrorType = "timeout"
	ErrorTypeDNS               ErrorType = "dns"
	ErrorTypeConnectionRefused ErrorType = "connection_refused"
	ErrorTypeTLS               ErrorType = "tls"
	ErrorTypeGeneral           ErrorType = "general"
)

// IsTimeoutError checks whether an error is a timeout.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "context deadline exceeded") ||
		strings.Contains(msg, "client.timeout exceeded") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "timeout")
}

// ClassifyNetworkError determines the category of connection error.
func ClassifyNetworkError(err error) ErrorType {
	if err == nil {
		return ErrorTypeNone
	}
	if IsTimeoutError(err) {
		return ErrorTypeTimeout
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ErrorTypeDNS
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "no such host") || strings.Contains(msg, "lookup ") {
		return ErrorTypeDNS
	}
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "connect: connection refused") {
		return ErrorTypeConnectionRefused
	}
	if strings.Contains(msg, "certificate") || strings.Contains(msg, "x509") || strings.Contains(msg, "tls:") {
		return ErrorTypeTLS
	}

	return ErrorTypeGeneral
}

// FormatErrorMessage returns a clean, human-readable error explanation for alerts and check logs.
func FormatErrorMessage(err error, timeout time.Duration) string {
	if err == nil {
		return ""
	}

	errType := ClassifyNetworkError(err)
	switch errType {
	case ErrorTypeTimeout:
		secs := int(timeout.Seconds())
		if secs > 0 {
			return fmt.Sprintf("Request timed out after %ds (Service did not respond in time)", secs)
		}
		return "Request timed out (Service did not respond in time)"

	case ErrorTypeDNS:
		return "DNS resolution failed: Host could not be reached or resolved"

	case ErrorTypeConnectionRefused:
		return "Connection refused: Host is offline or port is closed"

	case ErrorTypeTLS:
		return fmt.Sprintf("TLS/SSL error: %s", cleanRawError(err.Error()))

	default:
		return cleanRawError(err.Error())
	}
}

func cleanRawError(raw string) string {
	cleaned := raw
	prefixes := []string{
		"Get ",
		"Post ",
		"dial tcp: ",
		"dial tcp ",
	}
	for _, p := range prefixes {
		cleaned = strings.TrimPrefix(cleaned, p)
	}
	if len(cleaned) > 250 {
		cleaned = cleaned[:247] + "..."
	}
	return cleaned
}
