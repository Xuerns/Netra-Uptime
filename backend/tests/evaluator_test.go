package tests

import (
	"PulseWatch/internal/model"
	"PulseWatch/internal/worker"
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
	res := worker.EvaluateRules(mon, nil, 200, 150*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected no breach for 200 OK, got breached: %v", res)
	}

	// 2. Status code mismatch (e.g. 500)
	res = worker.EvaluateRules(mon, nil, 500, 80*time.Millisecond, nil)
	if !res.IsBreached || res.Severity != "critical" {
		t.Errorf("Expected critical breach for 500 mismatch, got: %+v", res)
	}

	// 3. Network error / Timeout
	res = worker.EvaluateRules(mon, nil, 0, 10*time.Second, errors.New("context deadline exceeded"))
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
	res := worker.EvaluateRules(mon, rules, 200, 300*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected pass, got: %+v", res)
	}

	// 2. High latency (200 OK, 2100ms) -> Warning breach
	res = worker.EvaluateRules(mon, rules, 200, 2100*time.Millisecond, nil)
	if !res.IsBreached || res.Severity != "warning" || res.RequiredBreaches != 3 {
		t.Errorf("Expected warning breach with required breaches 3, got: %+v", res)
	}

	// 3. 502 Bad Gateway (502, 100ms) -> Critical breach
	res = worker.EvaluateRules(mon, rules, 502, 100*time.Millisecond, nil)
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
	res = worker.EvaluateRules(mon, inactiveRules, 500, 100*time.Millisecond, nil)
	if res.IsBreached {
		t.Errorf("Expected inactive rule not to trigger, got: %+v", res)
	}
}
