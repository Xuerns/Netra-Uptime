package alerttemplate

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"netra-uptime/internal/alertrule"
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

type mockTemplateRepo struct {
	createFunc                  func(template *model.AlertRuleTemplate) error
	getAllFunc                  func(orgID uuid.UUID) ([]model.AlertRuleTemplate, error)
	getByIDFunc                 func(orgID, templateID uuid.UUID) (*model.AlertRuleTemplate, error)
	updateFunc                  func(template *model.AlertRuleTemplate) error
	deleteFunc                  func(orgID, templateID uuid.UUID) error
	createRuleFunc              func(rule *model.AlertRule) error
	getRulesFunc                func(templateID uuid.UUID) ([]model.AlertRule, error)
	getRuleByIDFunc             func(templateID, ruleID uuid.UUID) (*model.AlertRule, error)
	updateRuleFunc              func(rule *model.AlertRule) error
	deleteRuleFunc              func(templateID, ruleID uuid.UUID) error
	assignTemplateToMonitorFunc func(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error
}

func (m *mockTemplateRepo) Create(template *model.AlertRuleTemplate) error {
	if m.createFunc != nil {
		return m.createFunc(template)
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	return nil
}

func (m *mockTemplateRepo) GetAll(orgID uuid.UUID) ([]model.AlertRuleTemplate, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(orgID)
	}
	return nil, nil
}

func (m *mockTemplateRepo) GetByID(orgID, templateID uuid.UUID) (*model.AlertRuleTemplate, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(orgID, templateID)
	}
	return nil, errors.New("template not found")
}

func (m *mockTemplateRepo) Update(template *model.AlertRuleTemplate) error {
	if m.updateFunc != nil {
		return m.updateFunc(template)
	}
	return nil
}

func (m *mockTemplateRepo) Delete(orgID, templateID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgID, templateID)
	}
	return nil
}

func (m *mockTemplateRepo) CreateRule(rule *model.AlertRule) error {
	if m.createRuleFunc != nil {
		return m.createRuleFunc(rule)
	}
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return nil
}

func (m *mockTemplateRepo) GetRules(templateID uuid.UUID) ([]model.AlertRule, error) {
	if m.getRulesFunc != nil {
		return m.getRulesFunc(templateID)
	}
	return nil, nil
}

func (m *mockTemplateRepo) GetRuleByID(templateID, ruleID uuid.UUID) (*model.AlertRule, error) {
	if m.getRuleByIDFunc != nil {
		return m.getRuleByIDFunc(templateID, ruleID)
	}
	return nil, errors.New("rule not found")
}

func (m *mockTemplateRepo) UpdateRule(rule *model.AlertRule) error {
	if m.updateRuleFunc != nil {
		return m.updateRuleFunc(rule)
	}
	return nil
}

func (m *mockTemplateRepo) DeleteRule(templateID, ruleID uuid.UUID) error {
	if m.deleteRuleFunc != nil {
		return m.deleteRuleFunc(templateID, ruleID)
	}
	return nil
}

func (m *mockTemplateRepo) AssignTemplateToMonitor(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error {
	if m.assignTemplateToMonitorFunc != nil {
		return m.assignTemplateToMonitorFunc(orgID, monitorID, templateID)
	}
	return nil
}

func TestTemplateService_CreateAndGet(t *testing.T) {
	orgID := uuid.New()
	templateID := uuid.New()

	repo := &mockTemplateRepo{
		createFunc: func(tmpl *model.AlertRuleTemplate) error {
			tmpl.ID = templateID
			tmpl.CreatedAt = time.Now()
			return nil
		},
		getByIDFunc: func(oID, tID uuid.UUID) (*model.AlertRuleTemplate, error) {
			if oID == orgID && tID == templateID {
				return &model.AlertRuleTemplate{
					ID:             templateID,
					OrganizationID: orgID,
					Name:           "Standard API",
					Description:    "For web services",
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	svc := NewService(repo)

	created, err := svc.Create(orgID, &CreateAlertTemplateRequest{
		Name:        "Standard API",
		Description: "For web services",
	})
	if err != nil {
		t.Fatalf("unexpected error creating template: %v", err)
	}
	if created.Name != "Standard API" {
		t.Errorf("expected name 'Standard API', got %s", created.Name)
	}

	fetched, err := svc.GetByID(orgID, templateID)
	if err != nil {
		t.Fatalf("unexpected error fetching template: %v", err)
	}
	if fetched.ID != templateID {
		t.Errorf("expected ID %v, got %v", templateID, fetched.ID)
	}
}

func TestTemplateService_AssignToMonitor(t *testing.T) {
	orgID := uuid.New()
	templateID := uuid.New()
	monitorID := uuid.New()

	assigned := false
	repo := &mockTemplateRepo{
		getByIDFunc: func(oID, tID uuid.UUID) (*model.AlertRuleTemplate, error) {
			if oID == orgID && tID == templateID {
				return &model.AlertRuleTemplate{ID: templateID}, nil
			}
			return nil, errors.New("not found")
		},
		assignTemplateToMonitorFunc: func(oID, mID uuid.UUID, tID *uuid.UUID) error {
			if oID == orgID && mID == monitorID && tID != nil && *tID == templateID {
				assigned = true
				return nil
			}
			return errors.New("invalid arguments")
		},
	}

	svc := NewService(repo)
	err := svc.AssignToMonitor(orgID, monitorID, &templateID)
	if err != nil {
		t.Fatalf("unexpected error assigning template: %v", err)
	}
	if !assigned {
		t.Errorf("expected assign function to be called")
	}
}

func TestTemplateHandler_HTTPFlow(t *testing.T) {
	orgID := uuid.New()
	templateID := uuid.New()

	repo := &mockTemplateRepo{
		createFunc: func(tmpl *model.AlertRuleTemplate) error {
			tmpl.ID = templateID
			return nil
		},
		getAllFunc: func(oID uuid.UUID) ([]model.AlertRuleTemplate, error) {
			return []model.AlertRuleTemplate{
				{
					ID:             templateID,
					OrganizationID: orgID,
					Name:           "Template A",
				},
			}, nil
		},
	}

	svc := NewService(repo)
	handler := NewHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/alert-templates", handler.List)
	r.POST("/alert-templates", handler.Create)

	// Test List
	reqList, _ := http.NewRequest(http.MethodGet, "/alert-templates", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wList.Code)
	}

	// Test Create
	payload := CreateAlertTemplateRequest{
		Name: "New Template",
		Rules: []alertrule.CreateAlertRuleRequest{
			{
				Name:        "5xx Error",
				Metric:      "status_code",
				Operator:    "==",
				TargetValue: "5xx",
				Consecutive: 2,
				Severity:    "critical",
			},
		},
	}
	body, _ := json.Marshal(payload)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/alert-templates", bytes.NewBuffer(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	r.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", wCreate.Code, wCreate.Body.String())
	}
}
