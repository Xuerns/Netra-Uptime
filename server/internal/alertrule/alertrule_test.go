package alertrule

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// Mock Repository
type mockAlertRuleRepo struct {
	createFunc       func(rule *model.AlertRule) error
	getByMonitorFunc func(orgID, monitorID uuid.UUID) ([]model.AlertRule, error)
	getByIDFunc      func(orgID, ruleID uuid.UUID) (*model.AlertRule, error)
	updateFunc       func(rule *model.AlertRule) error
	deleteFunc       func(orgID, ruleID uuid.UUID) error
}

func (m *mockAlertRuleRepo) Create(rule *model.AlertRule) error {
	if m.createFunc != nil {
		return m.createFunc(rule)
	}
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return nil
}

func (m *mockAlertRuleRepo) GetByMonitor(orgID, monitorID uuid.UUID) ([]model.AlertRule, error) {
	if m.getByMonitorFunc != nil {
		return m.getByMonitorFunc(orgID, monitorID)
	}
	return nil, nil
}

func (m *mockAlertRuleRepo) GetByID(orgID, ruleID uuid.UUID) (*model.AlertRule, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(orgID, ruleID)
	}
	return nil, errors.New("record not found")
}

func (m *mockAlertRuleRepo) Update(rule *model.AlertRule) error {
	if m.updateFunc != nil {
		return m.updateFunc(rule)
	}
	return nil
}

func (m *mockAlertRuleRepo) Delete(orgID, ruleID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgID, ruleID)
	}
	return nil
}

func (m *mockAlertRuleRepo) GetMonitorType(orgID, monitorID uuid.UUID) (string, error) {
	return "http", nil
}

// Mock Service for handler tests
type mockAlertRuleService struct {
	createFunc       func(orgID, monitorID uuid.UUID, req *CreateAlertRuleRequest) (*AlertRuleResponse, error)
	getByMonitorFunc func(orgID, monitorID uuid.UUID) ([]AlertRuleResponse, error)
	getByIDFunc      func(orgID, ruleID uuid.UUID) (*AlertRuleResponse, error)
	updateFunc       func(orgID, ruleID uuid.UUID, req *UpdateAlertRuleRequest) (*AlertRuleResponse, error)
	deleteFunc       func(orgID, ruleID uuid.UUID) error
}

func (m *mockAlertRuleService) Create(orgID, monitorID uuid.UUID, req *CreateAlertRuleRequest) (*AlertRuleResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(orgID, monitorID, req)
	}
	return &AlertRuleResponse{ID: uuid.New(), Name: req.Name}, nil
}

func (m *mockAlertRuleService) GetByMonitor(orgID, monitorID uuid.UUID) ([]AlertRuleResponse, error) {
	if m.getByMonitorFunc != nil {
		return m.getByMonitorFunc(orgID, monitorID)
	}
	return nil, nil
}

func (m *mockAlertRuleService) GetByID(orgID, ruleID uuid.UUID) (*AlertRuleResponse, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(orgID, ruleID)
	}
	return nil, errors.New("Alert rule not found")
}

func (m *mockAlertRuleService) Update(orgID, ruleID uuid.UUID, req *UpdateAlertRuleRequest) (*AlertRuleResponse, error) {
	if m.updateFunc != nil {
		return m.updateFunc(orgID, ruleID, req)
	}
	return &AlertRuleResponse{ID: ruleID}, nil
}

func (m *mockAlertRuleService) Delete(orgID, ruleID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgID, ruleID)
	}
	return nil
}

// ====================== SERVICE TESTS ======================

func TestAlertRuleService_Create(t *testing.T) {
	repo := &mockAlertRuleRepo{}
	svc := NewService(repo)

	orgID := uuid.New()
	monitorID := uuid.New()
	isActive := false
	req := &CreateAlertRuleRequest{
		Name:        "High Response Time",
		Metric:      "response_time",
		Operator:    ">",
		TargetValue: "1000",
		Consecutive: 3,
		Severity:    "critical",
		IsActive:    &isActive,
	}

	res, err := svc.Create(orgID, monitorID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, res.Name)
	}
	if res.IsActive != false {
		t.Errorf("expected isActive false, got %v", res.IsActive)
	}
}

func TestAlertRuleService_GetByMonitor(t *testing.T) {
	orgID := uuid.New()
	monitorID := uuid.New()
	repo := &mockAlertRuleRepo{
		getByMonitorFunc: func(o, m uuid.UUID) ([]model.AlertRule, error) {
			return []model.AlertRule{
				{ID: uuid.New(), Name: "Rule 1", OrganizationID: o, MonitorID: &m},
				{ID: uuid.New(), Name: "Rule 2", OrganizationID: o, MonitorID: &m},
			}, nil
		},
	}
	svc := NewService(repo)

	rules, err := svc.GetByMonitor(orgID, monitorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestAlertRuleService_GetByID_Success(t *testing.T) {
	orgID := uuid.New()
	ruleID := uuid.New()
	repo := &mockAlertRuleRepo{
		getByIDFunc: func(o, r uuid.UUID) (*model.AlertRule, error) {
			return &model.AlertRule{ID: r, OrganizationID: o, Name: "Target Rule"}, nil
		},
	}
	svc := NewService(repo)

	res, err := svc.GetByID(orgID, ruleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "Target Rule" {
		t.Errorf("expected name Target Rule, got %s", res.Name)
	}
}

func TestAlertRuleService_GetByID_NotFound(t *testing.T) {
	repo := &mockAlertRuleRepo{
		getByIDFunc: func(o, r uuid.UUID) (*model.AlertRule, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewService(repo)

	_, err := svc.GetByID(uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlertRuleService_Update_Success(t *testing.T) {
	orgID := uuid.New()
	ruleID := uuid.New()
	existing := &model.AlertRule{
		ID:             ruleID,
		OrganizationID: orgID,
		Name:           "Old Name",
		Metric:         "response_time",
	}
	repo := &mockAlertRuleRepo{
		getByIDFunc: func(o, r uuid.UUID) (*model.AlertRule, error) {
			return existing, nil
		},
		updateFunc: func(rule *model.AlertRule) error {
			return nil
		},
	}
	svc := NewService(repo)

	newName := "New Name"
	res, err := svc.Update(orgID, ruleID, &UpdateAlertRuleRequest{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "New Name" {
		t.Errorf("expected New Name, got %s", res.Name)
	}
}

func TestAlertRuleService_Delete(t *testing.T) {
	deleted := false
	repo := &mockAlertRuleRepo{
		deleteFunc: func(orgID, ruleID uuid.UUID) error {
			deleted = true
			return nil
		},
	}
	svc := NewService(repo)
	err := svc.Delete(uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete to be called")
	}
}

// ====================== HANDLER TESTS ======================

func TestAlertRuleHandler_Create_Success(t *testing.T) {
	orgID := uuid.New()
	monitorID := uuid.New()
	handler := NewHandler(&mockAlertRuleService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.POST("/organization/:orgId/monitors/:monitorId/alert-rules", handler.Create)

	body := CreateAlertRuleRequest{
		Name:        "Test Rule",
		Metric:      "status_code",
		Operator:    "!=",
		TargetValue: "200",
		Consecutive: 1,
		Severity:    "warning",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/monitors/"+monitorID.String()+"/alert-rules", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAlertRuleHandler_Create_InvalidMonitorUUID(t *testing.T) {
	orgID := uuid.New()
	handler := NewHandler(&mockAlertRuleService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.POST("/organization/:orgId/monitors/:monitorId/alert-rules", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/monitors/invalid-uuid/alert-rules", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid monitorId, got %d", w.Code)
	}
}

func TestAlertRuleHandler_GetByID_NotFound(t *testing.T) {
	orgID := uuid.New()
	ruleID := uuid.New()
	handler := NewHandler(&mockAlertRuleService{
		getByIDFunc: func(o, r uuid.UUID) (*AlertRuleResponse, error) {
			return nil, errors.New("Alert rule not found")
		},
	})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/alert-rules/:ruleId", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/alert-rules/"+ruleID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
