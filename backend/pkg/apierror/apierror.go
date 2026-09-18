package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var rawValidatorRegex = regexp.MustCompile(`(?i)Field validation for '([^']+)' failed on the '([^']+)' tag`)

var fieldAcronyms = map[string]string{
	"id":          "ID",
	"url":         "URL",
	"api":         "API",
	"key":         "Key",
	"ssl":         "SSL",
	"ip":          "IP",
	"pics":        "PICs",
	"html":        "HTML",
	"http":        "HTTP",
	"uuid":        "UUID",
	"dns":         "DNS",
	"tcp":         "TCP",
	"smtp":        "SMTP",
	"status_code": "Status Code",
}

// HumanizeFieldName converts struct or JSON field names into clean, readable labels.
// e.g. "MonitorPICs" -> "Monitor PICs", "OrganizationID" -> "Organization ID",
// "current_password" -> "Current Password", "target_value" -> "Target Value",
// "IntervalSeconds" -> "Interval Seconds", "nohtml" -> "Disallowed HTML"
func HumanizeFieldName(field string) string {
	if field == "" {
		return "Field"
	}

	// Replace underscores and hyphens with spaces
	cleaned := strings.ReplaceAll(field, "_", " ")
	cleaned = strings.ReplaceAll(cleaned, "-", " ")

	// Split camelCase and PascalCase
	var words []string
	var current strings.Builder

	runes := []rune(cleaned)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			var next rune
			if i+1 < len(runes) {
				next = runes[i+1]
			}

			// Boundary detection:
			// lowercase followed by uppercase (e.g. "monitorId" -> "monitor", "Id")
			// or uppercase sequence followed by lowercase (e.g. "APIKey" or "HTTPStatus")
			if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z' && next >= 'a' && next <= 'z') {
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	for i, w := range words {
		lower := strings.ToLower(w)
		if ac, ok := fieldAcronyms[lower]; ok {
			words[i] = ac
		} else if len(w) > 0 {
			// Title case
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}

	result := strings.Join(words, " ")
	return strings.TrimSpace(result)
}

// FormatValidationError formats a validator error or binding error into a clean,
// user-friendly, specific message without technical jargon.
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}

	// 1. Check if validator.ValidationErrors
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		var msgs []string
		for _, fe := range valErrs {
			msgs = append(msgs, formatFieldError(fe))
		}
		if len(msgs) > 0 {
			return strings.Join(msgs, "; ")
		}
	}

	// 2. Check for empty body / EOF
	if errors.Is(err, io.EOF) || err.Error() == "EOF" {
		return "Request body cannot be empty"
	}

	// 3. Check for JSON unmarshal type error
	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		field := unmarshalTypeErr.Field
		if field == "" {
			return "Invalid data type in request payload"
		}
		return fmt.Sprintf("Invalid data type for %s: expected %s", HumanizeFieldName(field), unmarshalTypeErr.Type.String())
	}

	// 4. Check for JSON syntax error
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "Invalid JSON format in request payload"
	}

	// 5. Check if raw error string contains validator pattern
	rawMsg := err.Error()
	if parsed := parseRawValidatorString(rawMsg); parsed != "" {
		return parsed
	}

	// Fallback for general unmarshal/binding errors
	if strings.Contains(strings.ToLower(rawMsg), "json") || strings.Contains(rawMsg, "unmarshal") {
		return "Invalid request payload format"
	}

	return rawMsg
}

func formatFieldError(fe validator.FieldError) string {
	field := HumanizeFieldName(fe.Field())
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		switch fe.Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at least %s characters", field, param)
		case reflect.Slice, reflect.Array:
			return fmt.Sprintf("%s must contain at least %s item(s)", field, param)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%s must be at least %s", field, param)
		default:
			return fmt.Sprintf("%s must be at least %s characters", field, param)
		}
	case "max":
		switch fe.Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at most %s characters", field, param)
		case reflect.Slice, reflect.Array:
			return fmt.Sprintf("%s must contain at most %s item(s)", field, param)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%s must be at most %s", field, param)
		default:
			return fmt.Sprintf("%s must be at most %s characters", field, param)
		}
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		options := strings.Join(strings.Fields(param), ", ")
		return fmt.Sprintf("%s must be one of: %s", field, options)
	case "nohtml":
		return fmt.Sprintf("%s must not contain HTML tags", field)
	case "printascii":
		return fmt.Sprintf("%s must contain only printable ASCII characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "eq":
		if strings.Contains(param, "|") {
			return fmt.Sprintf("%s has an invalid value", field)
		}
		return fmt.Sprintf("%s must equal '%s'", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func parseRawValidatorString(raw string) string {
	matches := rawValidatorRegex.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return ""
	}

	var msgs []string
	for _, m := range matches {
		if len(m) >= 3 {
			fieldName := HumanizeFieldName(m[1])
			tag := m[2]
			switch tag {
			case "required":
				msgs = append(msgs, fmt.Sprintf("%s is required", fieldName))
			case "min":
				msgs = append(msgs, fmt.Sprintf("%s is too short", fieldName))
			case "max":
				msgs = append(msgs, fmt.Sprintf("%s is too long", fieldName))
			case "email":
				msgs = append(msgs, fmt.Sprintf("%s must be a valid email address", fieldName))
			case "url":
				msgs = append(msgs, fmt.Sprintf("%s must be a valid URL", fieldName))
			case "nohtml":
				msgs = append(msgs, fmt.Sprintf("%s must not contain HTML tags", fieldName))
			case "oneof":
				msgs = append(msgs, fmt.Sprintf("%s has an invalid option", fieldName))
			default:
				msgs = append(msgs, fmt.Sprintf("%s is invalid", fieldName))
			}
		}
	}

	if len(msgs) > 0 {
		return strings.Join(msgs, "; ")
	}
	return ""
}

// SanitizeServiceError sanitizes database errors, internal leaks, and driver exceptions
// into clear, safe, user-friendly error messages while preserving clean application business errors.
func SanitizeServiceError(err error) string {
	if err == nil {
		return ""
	}

	raw := err.Error()
	lowered := strings.ToLower(raw)

	// 1. Duplicate key constraints (Postgres/GORM)
	if strings.Contains(lowered, "duplicate key") || strings.Contains(lowered, "unique constraint") {
		if strings.Contains(lowered, "email") {
			return "An account with this email already exists."
		}
		if strings.Contains(lowered, "slug") || strings.Contains(lowered, "subdomain") {
			return "This identifier or subdomain is already taken."
		}
		if strings.Contains(lowered, "name") {
			return "A record with this name already exists."
		}
		return "A record with this information already exists."
	}

	// 2. Record not found
	if strings.Contains(lowered, "record not found") || strings.Contains(lowered, "no rows in result set") {
		return "The requested resource was not found."
	}

	// 3. Foreign key violation
	if strings.Contains(lowered, "violates foreign key constraint") {
		return "Referenced resource does not exist or cannot be modified."
	}

	// 4. Low level DB / Network / Driver errors
	if strings.Contains(lowered, "connection refused") ||
		strings.Contains(lowered, "driver: bad connection") ||
		strings.Contains(lowered, "dial tcp") ||
		strings.Contains(lowered, "pq:") ||
		strings.Contains(lowered, "sql:") ||
		strings.Contains(lowered, "gorm:") {
		return "A database error occurred. Please try again later."
	}

	// 5. Raw validator string in service error
	if strings.Contains(raw, "Field validation for") {
		return FormatValidationError(err)
	}

	// 6. Prevent internal runtime traces from leaking
	if strings.Contains(raw, "0x") || strings.Contains(raw, ".go:") || strings.Contains(raw, "goroutine") {
		return "An unexpected server error occurred. Please try again later."
	}

	return raw
}

// SanitizeDispatchError sanitizes notification dispatch errors (SMTP, Webhook, Discord, Slack)
// into concise, readable error notes without leaking full stack traces.
func SanitizeDispatchError(err error) string {
	if err == nil {
		return ""
	}

	raw := err.Error()
	lowered := strings.ToLower(raw)

	// SMTP issues
	if strings.Contains(lowered, "daily user sending limit exceeded") || strings.Contains(lowered, "5.4.5") {
		return "Email delivery failed: Daily sending limit exceeded on SMTP provider"
	}
	if strings.Contains(lowered, "550") {
		return "Email delivery rejected by mail server (550)"
	}
	if strings.Contains(lowered, "authentication failed") || (strings.Contains(lowered, "auth") && strings.Contains(lowered, "smtp")) {
		return "Email delivery failed: SMTP authentication error"
	}
	if strings.Contains(lowered, "dial tcp") || strings.Contains(lowered, "connect: connection refused") {
		if strings.Contains(lowered, "smtp") || strings.Contains(lowered, "mail") || strings.Contains(lowered, "25") || strings.Contains(lowered, "587") || strings.Contains(lowered, "465") {
			return "Email delivery failed: Unable to connect to SMTP server"
		}
		return "Connection failed: Host unreachable or connection refused"
	}

	// Webhook / HTTP issues
	if strings.Contains(lowered, "context deadline exceeded") || strings.Contains(lowered, "timeout") {
		return "Notification delivery timed out while reaching target endpoint"
	}
	if strings.Contains(lowered, "no such host") {
		return "Notification delivery failed: Target host could not be resolved"
	}

	// Clean out verbose prefixes
	cleaned := raw
	prefixes := []string{
		"Failed to send email: ",
		"send failed: ",
		"sending SMTP DATA command: ",
	}
	for _, p := range prefixes {
		cleaned = strings.TrimPrefix(cleaned, p)
	}

	if len(cleaned) > 200 {
		cleaned = cleaned[:197] + "..."
	}

	return cleaned
}
