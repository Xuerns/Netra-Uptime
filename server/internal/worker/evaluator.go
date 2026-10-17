package worker

import (
	"fmt"
	"netra-uptime/internal/model"
	"netra-uptime/internal/worker/checker"
	"strconv"
	"strings"
	"time"
)

// RuleEvaluationResult represents the outcome of evaluating check results against rules.
type RuleEvaluationResult struct {
	IsBreached       bool
	Severity         string // "critical" | "warning"
	RuleName         string
	Reason           string
	RequiredBreaches int
	RuleID           *string
	ErrorType        checker.ErrorType
}

// EvaluateRules evaluates check metrics against custom monitor rules or default rules.
func EvaluateRules(monitor *model.Monitor, rules []model.AlertRule, statusCode int, responseTime time.Duration, checkErr error) RuleEvaluationResult {
	respMs := int(responseTime.Milliseconds())

	timeoutDur := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeoutDur <= 0 {
		timeoutDur = 10 * time.Second
	}

	// Filter active rules
	var activeRules []model.AlertRule
	for _, r := range rules {
		if r.IsActive {
			activeRules = append(activeRules, r)
		}
	}

	// 1. If custom rules exist, evaluate them in order of severity (critical first)
	if len(activeRules) > 0 {
		var warningResult *RuleEvaluationResult

		for _, rule := range activeRules {
			breached, reason := evaluateSingleRule(rule, statusCode, respMs, checkErr, timeoutDur)
			if breached {
				ruleIDStr := rule.ID.String()
				consecutive := rule.Consecutive
				if consecutive <= 0 {
					consecutive = 1
				}

				res := RuleEvaluationResult{
					IsBreached:       true,
					Severity:         rule.Severity,
					RuleName:         rule.Name,
					Reason:           reason,
					RequiredBreaches: consecutive,
					RuleID:           &ruleIDStr,
					ErrorType:        checker.ClassifyNetworkError(checkErr),
				}

				if rule.Severity == "critical" {
					return res // Immediate critical breach takes highest precedence
				} else if warningResult == nil {
					warningResult = &res
				}
			}
		}

		if warningResult != nil {
			return *warningResult
		}

		// Also evaluate basic fatal connection errors if no rule explicitly matched but connection failed
		if checkErr != nil {
			isTimeout := checker.IsTimeoutError(checkErr)
			if isTimeout {
				return RuleEvaluationResult{
					IsBreached:       true,
					Severity:         "warning",
					RuleName:         "Request Timeout",
					Reason:           checker.FormatErrorMessage(checkErr, timeoutDur),
					RequiredBreaches: 3,
					ErrorType:        checker.ErrorTypeTimeout,
				}
			}
			return RuleEvaluationResult{
				IsBreached:       true,
				Severity:         "critical",
				RuleName:         "Connection Failure",
				Reason:           checker.FormatErrorMessage(checkErr, timeoutDur),
				RequiredBreaches: 3,
				ErrorType:        checker.ClassifyNetworkError(checkErr),
			}
		}

		// All rules passed
		return RuleEvaluationResult{
			IsBreached: false,
		}
	}

	// 2. Default Fallback Rules (when monitor has no custom alert rules configured)
	if checkErr != nil {
		isTimeout := checker.IsTimeoutError(checkErr)
		if isTimeout {
			return RuleEvaluationResult{
				IsBreached:       true,
				Severity:         "warning",
				RuleName:         "Request Timeout",
				Reason:           checker.FormatErrorMessage(checkErr, timeoutDur),
				RequiredBreaches: 3,
				ErrorType:        checker.ErrorTypeTimeout,
			}
		}
		return RuleEvaluationResult{
			IsBreached:       true,
			Severity:         "critical",
			RuleName:         "Connection Failure",
			Reason:           checker.FormatErrorMessage(checkErr, timeoutDur),
			RequiredBreaches: 3,
			ErrorType:        checker.ClassifyNetworkError(checkErr),
		}
	}

	if monitor.Type == "http" && monitor.ExpectedStatusCode != 0 && statusCode != monitor.ExpectedStatusCode {
		return RuleEvaluationResult{
			IsBreached:       true,
			Severity:         "critical",
			RuleName:         "Status Code Mismatch",
			Reason:           fmt.Sprintf("Status code mismatch (expected %d, got %d)", monitor.ExpectedStatusCode, statusCode),
			RequiredBreaches: 3,
		}
	}

	return RuleEvaluationResult{
		IsBreached: false,
	}
}

func evaluateSingleRule(rule model.AlertRule, statusCode, respMs int, checkErr error, timeout time.Duration) (bool, string) {
	switch strings.ToLower(rule.Metric) {
	case "status_code":
		if checkErr != nil {
			// If connection failed entirely, status code is 0
			if rule.Operator == "!=" || rule.Operator == "not_in" {
				return true, fmt.Sprintf("Connection error (%s) before HTTP status code", checker.FormatErrorMessage(checkErr, timeout))
			}
			return false, ""
		}
		return evaluateStatusCode(statusCode, rule.Operator, rule.TargetValue)

	case "response_time_ms", "latency":
		if checkErr != nil {
			return false, ""
		}
		targetVal, err := strconv.Atoi(strings.TrimSpace(rule.TargetValue))
		if err != nil {
			return false, ""
		}
		return evaluateNumeric(respMs, rule.Operator, targetVal, fmt.Sprintf("Response time %dms", respMs))

	case "timeout":
		if checkErr != nil {
			if checker.IsTimeoutError(checkErr) {
				return true, checker.FormatErrorMessage(checkErr, timeout)
			}
			return false, ""
		}
		// Also evaluate if explicit latency target given in ms
		targetVal, err := strconv.Atoi(strings.TrimSpace(rule.TargetValue))
		if err == nil && targetVal > 0 {
			return evaluateNumeric(respMs, rule.Operator, targetVal, fmt.Sprintf("Response time %dms exceeded timeout threshold", respMs))
		}
		return false, ""

	case "connection":
		if checkErr != nil {
			return true, fmt.Sprintf("Connection error: %s", checker.FormatErrorMessage(checkErr, timeout))
		}
		return false, ""

	default:
		return false, ""
	}
}

func evaluateStatusCode(actual int, op, target string) (bool, string) {
	targetClean := strings.TrimSpace(strings.ToLower(target))

	// Handle status families like "5xx", "4xx", "3xx", "2xx"
	if strings.HasSuffix(targetClean, "xx") && len(targetClean) == 3 {
		prefix := targetClean[0]
		actualPrefix := fmt.Sprintf("%d", actual)
		isFamilyMatch := len(actualPrefix) == 3 && actualPrefix[0] == prefix

		switch op {
		case "==", "=", "in":
			if isFamilyMatch {
				return true, fmt.Sprintf("Status code %d matches %s", actual, targetClean)
			}
		case "!=", "not_in":
			if !isFamilyMatch {
				return true, fmt.Sprintf("Status code %d is not %s", actual, targetClean)
			}
		}
		return false, ""
	}

	// Handle comma-separated list e.g. "500,502,503"
	if strings.Contains(targetClean, ",") {
		parts := strings.Split(targetClean, ",")
		found := false
		for _, p := range parts {
			if val, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && val == actual {
				found = true
				break
			}
		}
		if (op == "in" || op == "==") && found {
			return true, fmt.Sprintf("Status code %d in [%s]", actual, targetClean)
		}
		if (op == "not_in" || op == "!=") && !found {
			return true, fmt.Sprintf("Status code %d not in [%s]", actual, targetClean)
		}
		return false, ""
	}

	// Handle numeric comparisons e.g. != 200 or == 500 or >= 400
	targetNum, err := strconv.Atoi(targetClean)
	if err != nil {
		return false, ""
	}

	return evaluateNumeric(actual, op, targetNum, fmt.Sprintf("Status code %d", actual))
}

func evaluateNumeric(actual int, op string, target int, prefix string) (bool, string) {
	switch op {
	case ">":
		if actual > target {
			return true, fmt.Sprintf("%s exceeded %d", prefix, target)
		}
	case ">=":
		if actual >= target {
			return true, fmt.Sprintf("%s >= %d", prefix, target)
		}
	case "<":
		if actual < target {
			return true, fmt.Sprintf("%s below %d", prefix, target)
		}
	case "<=":
		if actual <= target {
			return true, fmt.Sprintf("%s <= %d", prefix, target)
		}
	case "==":
		if actual == target {
			return true, fmt.Sprintf("%s equals %d", prefix, target)
		}
	case "!=":
		if actual != target {
			return true, fmt.Sprintf("%s != %d", prefix, target)
		}
	}
	return false, ""
}

// ResolveIncidentSeverity resolves an incident severity level (critical, high, medium, low)
// based on monitor criticality (extreme, high, medium, low) and alert rule severity (critical, warning).
func ResolveIncidentSeverity(monitorCriticality, alertSeverity string) string {
	crit := strings.ToLower(strings.TrimSpace(monitorCriticality))
	sev := strings.ToLower(strings.TrimSpace(alertSeverity))

	switch crit {
	case "extreme":
		if sev == "critical" {
			return "critical"
		}
		return "high"
	case "high":
		if sev == "critical" {
			return "critical"
		}
		return "high"
	case "medium":
		if sev == "critical" {
			return "high"
		}
		return "medium"
	case "low":
		if sev == "critical" {
			return "medium"
		}
		return "low"
	default:
		if sev == "critical" {
			return "high"
		}
		return "medium"
	}
}
