package apierror_test

import (
	"errors"
	"io"
	"netra-uptime/pkg/apierror"
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestUser struct {
	Name        string   `validate:"required,min=2,max=50"`
	Email       string   `validate:"required,email"`
	URL         string   `validate:"url"`
	Role        string   `validate:"oneof=admin member viewer"`
	Age         int      `validate:"min=18,max=120"`
	MonitorPICs []string `validate:"min=1,max=3"`
}

func TestFormatValidationError_WithValidator(t *testing.T) {
	validate := validator.New()

	u := TestUser{
		Name:  "A", // fails min=2
		Email: "invalid-email",
		Role:  "superuser", // fails oneof
	}

	err := validate.Struct(u)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	msg := apierror.FormatValidationError(err)
	if msg == "" {
		t.Fatal("expected non-empty message")
	}

	// Should contain human-friendly phrases
	expectedSubstrings := []string{
		"Name must be at least 2 characters",
		"Email must be a valid email address",
		"Role must be one of: admin, member, viewer",
	}

	for _, sub := range expectedSubstrings {
		if !contains(msg, sub) {
			t.Errorf("expected message to contain %q, got %q", sub, msg)
		}
	}
}

func TestFormatValidationError_SpecialCases(t *testing.T) {
	if got := apierror.FormatValidationError(io.EOF); got != "Request body cannot be empty" {
		t.Errorf("expected Request body cannot be empty, got %q", got)
	}

	rawStringErr := errors.New("Key: 'CreateAPIKeyRequest.Name' Error:Field validation for 'Name' failed on the 'min' tag")
	got := apierror.FormatValidationError(rawStringErr)
	if got != "Name is too short" {
		t.Errorf("expected 'Name is too short', got %q", got)
	}
}

func TestSanitizeServiceError(t *testing.T) {
	tests := []struct {
		input    error
		expected string
	}{
		{
			input:    errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`),
			expected: "An account with this email already exists.",
		},
		{
			input:    errors.New(`pq: duplicate key value violates unique constraint "organizations_slug_key"`),
			expected: "This identifier or subdomain is already taken.",
		},
		{
			input:    errors.New("record not found"),
			expected: "The requested resource was not found.",
		},
		{
			input:    errors.New("dial tcp 127.0.0.1:5432: connect: connection refused"),
			expected: "A database error occurred. Please try again later.",
		},
		{
			input:    errors.New("You are not member of this organization"),
			expected: "You are not member of this organization",
		},
	}

	for _, tc := range tests {
		got := apierror.SanitizeServiceError(tc.input)
		if got != tc.expected {
			t.Errorf("input %v: expected %q, got %q", tc.input, tc.expected, got)
		}
	}
}

func TestSanitizeDispatchError(t *testing.T) {
	tests := []struct {
		input    error
		expected string
	}{
		{
			input:    errors.New("Failed to send email: send failed: sending SMTP DATA command: 550 '5.4.5 Daily user sending limit exceeded'"),
			expected: "Email delivery failed: Daily sending limit exceeded on SMTP provider",
		},
		{
			input:    errors.New("connect: connection refused on smtp.gmail.com:587"),
			expected: "Email delivery failed: Unable to connect to SMTP server",
		},
		{
			input:    errors.New("context deadline exceeded"),
			expected: "Notification delivery timed out while reaching target endpoint",
		},
	}

	for _, tc := range tests {
		got := apierror.SanitizeDispatchError(tc.input)
		if got != tc.expected {
			t.Errorf("input %v: expected %q, got %q", tc.input, tc.expected, got)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
