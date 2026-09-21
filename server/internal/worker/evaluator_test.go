package worker

import (
	"PulseWatch/internal/model"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvaluateRules_DefaultFallback(t *testing.T) {
	mon := &model.Monitor{
		Type:               "http",
		ExpectedStatusCode: 200,
	}

	// 1. Success 200 OK
	res := EvaluateRules(mon, nil, 200, 150*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected no breach for 200 OK, got breached: %v", res)
	}

	// 2. Status code mismatch (e.g. 500)
	res = EvaluateRules(mon, nil, 500, 80*time.Millisecond, nil)
	if !res.IsBreached || res.Severity != "critical" {
		t.Errorf("Expected critical breach for 500 mismatch, got: %+v", res)
	}

	// 3. Network error / Timeout
	res = EvaluateRules(mon, nil, 0, 10*time.Second, errors.New("context deadline exceeded"))
	if !res.IsBreached || res.Severity != "critical" {
		t.Errorf("Expected critical breach for timeout error, got: %+v", res)
	}
}

func TestEvaluateRules_CustomRules(t *testing.T) {
	mon := &model.Monitor{
		Type: "http",
	}

	ruleID1 := uuid.New()
	ruleID2 := uuid.New()

	rules := []model.AlertRule{
		{
			ID:          ruleID1,
			Name:        "5xx Server Error Alert",
			Metric:      "status_code",
			Operator:    "==",
			TargetValue: "5xx",
			Consecutive: 2,
			Severity:    "critical",
			IsActive:    true,
		},
		{
			ID:          ruleID2,
			Name:        "High Latency Warning",
			Metric:      "response_time_ms",
			Operator:    ">",
			TargetValue: "1500",
			Consecutive: 3,
			Severity:    "warning",
			IsActive:    true,
		},
	}

	// 1. Normal check (200 OK, 300ms) -> Pass
	res := EvaluateRules(mon, rules, 200, 300*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected pass, got: %+v", res)
	}

	// 2. High latency (200 OK, 2100ms) -> Warning breach
	res = EvaluateRules(mon, rules, 200, 2100*time.Millisecond, nil)
	if !res.IsBreached || res.Severity != "warning" || res.RequiredBreaches != 3 {
		t.Errorf("Expected warning breach with required breaches 3, got: %+v", res)
	}

	// 3. 502 Bad Gateway (502, 100ms) -> Critical breach
	res = EvaluateRules(mon, rules, 502, 100*time.Millisecond, nil)
	if !res.IsBreached || res.Severity != "critical" || res.RequiredBreaches != 2 {
		t.Errorf("Expected critical breach with required breaches 2, got: %+v", res)
	}

	// 4. Inactive rule should not trigger
	inactiveRules := []model.AlertRule{
		{
			ID:          ruleID1,
			Name:        "Disabled rule",
			Metric:      "status_code",
			Operator:    "==",
			TargetValue: "5xx",
			Severity:    "critical",
			IsActive:    false,
		},
	}
	res = EvaluateRules(mon, inactiveRules, 500, 100*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected inactive rule not to trigger, got: %+v", res)
	}
}

func TestResolveIncidentSeverity(t *testing.T) {
	tests := []struct {
		criticality string
		alertSev    string
		expected    string
	}{
		{"extreme", "critical", "critical"},
		{"extreme", "warning", "high"},
		{"high", "critical", "critical"},
		{"high", "warning", "high"},
		{"medium", "critical", "high"},
		{"medium", "warning", "medium"},
		{"low", "critical", "medium"},
		{"low", "warning", "low"},
		{"unknown", "critical", "high"},
		{"unknown", "warning", "medium"},
	}

	for _, tc := range tests {
		actual := ResolveIncidentSeverity(tc.criticality, tc.alertSev)
		if actual != tc.expected {
			t.Errorf("ResolveIncidentSeverity(%q, %q) = %q; want %q", tc.criticality, tc.alertSev, actual, tc.expected)
		}
	}
}
