package statuspage

import (
	"strings"
	"testing"

	"netra-uptime/pkg/apierror"

	"github.com/gin-gonic/gin/binding"
)

func TestStatusPageDTOValidation(t *testing.T) {
	t.Run("CreateStatusPageRequest description length validation", func(t *testing.T) {
		validDesc := strings.Repeat("a", 255)
		req := CreateStatusPageRequest{
			Subdomain:   "valid-subdomain",
			Title:       "Valid Title",
			Description: &validDesc,
			IsPublic:    true,
			MonitorID:   []string{"11111111-1111-1111-1111-111111111111"},
		}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Errorf("expected 255-char description to be valid, got: %v", err)
		}

		tooLongDesc := strings.Repeat("a", 256)
		req.Description = &tooLongDesc
		err := binding.Validator.ValidateStruct(&req)
		if err == nil {
			t.Error("expected 256-char description to fail validation, got nil")
		} else {
			formatted := apierror.FormatValidationError(err)
			if !strings.Contains(formatted, "Description must be at most 255 characters") {
				t.Errorf("expected user-friendly max error message, got: %q", formatted)
			}
		}

		req.Description = nil
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Errorf("expected nil description to be valid, got: %v", err)
		}
	})

	t.Run("UpdateStatusPageRequest description length validation", func(t *testing.T) {
		validDesc := strings.Repeat("a", 255)
		req := UpdateStatusPageRequest{
			Description: &validDesc,
		}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Errorf("expected 255-char description to be valid, got: %v", err)
		}

		tooLongDesc := strings.Repeat("a", 256)
		req.Description = &tooLongDesc
		err := binding.Validator.ValidateStruct(&req)
		if err == nil {
			t.Error("expected 256-char description to fail validation, got nil")
		} else {
			formatted := apierror.FormatValidationError(err)
			if !strings.Contains(formatted, "Description must be at most 255 characters") {
				t.Errorf("expected user-friendly max error message, got: %q", formatted)
			}
		}

		req.Description = nil
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Errorf("expected nil description to be valid, got: %v", err)
		}
	})
}
