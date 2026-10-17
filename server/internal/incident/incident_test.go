package incident

import (
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
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
type mockIncidentRepo struct {
	getByIdFunc          func(orgId, incidentId string) (*model.Incident, error)
	getAllByOrgFunc      func(orgId string) ([]model.Incident, error)
	createUpdateFunc     func(update *model.IncidentUpdate) error
	getIncidentOnlyFunc  func(incidentId string) (*model.Incident, error)
	addAssigneeFunc      func(assignee *model.IncidentAssignee) error
	getAssigneeFunc      func(userId, incidentId string) (*model.IncidentAssignee, error)
	getMessageUpdateFunc func(incidentId string) ([]model.IncidentUpdate, error)
}

func (m *mockIncidentRepo) GetById(orgId, incidentId string) (*model.Incident, error) {
	if m.getByIdFunc != nil {
		return m.getByIdFunc(orgId, incidentId)
	}
	return nil, errors.New("incident not found")
}

func (m *mockIncidentRepo) GetAllByOrg(orgId string) ([]model.Incident, error) {
	if m.getAllByOrgFunc != nil {
		return m.getAllByOrgFunc(orgId)
	}
	return nil, nil
}

func (m *mockIncidentRepo) CreateUpdate(update *model.IncidentUpdate) error {
	if m.createUpdateFunc != nil {
		return m.createUpdateFunc(update)
	}
	return nil
}

func (m *mockIncidentRepo) GetIncidentOnly(incidentId string) (*model.Incident, error) {
	if m.getIncidentOnlyFunc != nil {
		return m.getIncidentOnlyFunc(incidentId)
	}
	return nil, nil
}

func (m *mockIncidentRepo) AddAssignee(assignee *model.IncidentAssignee) error {
	if m.addAssigneeFunc != nil {
		return m.addAssigneeFunc(assignee)
	}
	return nil
}

func (m *mockIncidentRepo) GetAssignee(userId, incidentId string) (*model.IncidentAssignee, error) {
	if m.getAssigneeFunc != nil {
		return m.getAssigneeFunc(userId, incidentId)
	}
	return nil, errors.New("assignee not found")
}

func (m *mockIncidentRepo) GetMessageUpdate(incidentId string) ([]model.IncidentUpdate, error) {
	if m.getMessageUpdateFunc != nil {
		return m.getMessageUpdateFunc(incidentId)
	}
	return nil, nil
}

// Mock Service
type mockIncidentService struct {
	getIncidentDetailFunc func(orgId, incidentId string) (*IncidentResponse, error)
	getAllIncidentFunc    func(orgId string) ([]IncidentResponse, error)
	addIncidentUpdateFunc func(orgId, incidentId, userId, message string) (*model.IncidentUpdate, error)
	joinAssigneeFunc      func(orgId, incidentId, userId string) error
	getIncidentUpdateFunc func(orgId, incidentId string) ([]GetIncidentUpdateResponse, error)
}

func (m *mockIncidentService) GetIncidentDetail(orgId, incidentId string) (*IncidentResponse, error) {
	if m.getIncidentDetailFunc != nil {
		return m.getIncidentDetailFunc(orgId, incidentId)
	}
	return &IncidentResponse{ID: incidentId}, nil
}

func (m *mockIncidentService) GetAllIncident(orgId string) ([]IncidentResponse, error) {
	if m.getAllIncidentFunc != nil {
		return m.getAllIncidentFunc(orgId)
	}
	return nil, nil
}

func (m *mockIncidentService) AddIncidentUpdate(orgId, incidentId, userId, message string) (*model.IncidentUpdate, error) {
	if m.addIncidentUpdateFunc != nil {
		return m.addIncidentUpdateFunc(orgId, incidentId, userId, message)
	}
	return &model.IncidentUpdate{Message: message}, nil
}

func (m *mockIncidentService) JoinAssignee(orgId, incidentId, userId string) error {
	if m.joinAssigneeFunc != nil {
		return m.joinAssigneeFunc(orgId, incidentId, userId)
	}
	return nil
}

func (m *mockIncidentService) GetIncidentUpdate(orgId, incidentId string) ([]GetIncidentUpdateResponse, error) {
	if m.getIncidentUpdateFunc != nil {
		return m.getIncidentUpdateFunc(orgId, incidentId)
	}
	return nil, nil
}

// ====================== SERVICE TESTS ======================

func TestIncidentService_GetIncidentDetail_Success(t *testing.T) {
	incID := uuid.New()
	monID := uuid.New()
	orgID := uuid.New().String()
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{
				ID:        incID,
				MonitorID: monID,
				Status:    "investigating",
				Severity:  "critical",
				Summary:   "API down",
				StartedAt: time.Now(),
			}, nil
		},
	}
	svc := NewIncidentService(repo)

	resp, err := svc.GetIncidentDetail(orgID, incID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != incID.String() {
		t.Errorf("expected ID %s, got %s", incID.String(), resp.ID)
	}
	if resp.Status != "investigating" {
		t.Errorf("expected status investigating, got %s", resp.Status)
	}
	if resp.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", resp.Severity)
	}
}

func TestIncidentService_GetAllIncident_Success(t *testing.T) {
	repo := &mockIncidentRepo{
		getAllByOrgFunc: func(orgId string) ([]model.Incident, error) {
			return []model.Incident{
				{ID: uuid.New(), Status: "resolved", StartedAt: time.Now()},
				{ID: uuid.New(), Status: "investigating", StartedAt: time.Now()},
			}, nil
		},
	}
	svc := NewIncidentService(repo)

	resp, err := svc.GetAllIncident(uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 incidents, got %d", len(resp))
	}
}

func TestIncidentService_AddIncidentUpdate_ResolvedIncident(t *testing.T) {
	incID := uuid.New().String()
	userID := uuid.New().String()
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{
				ID:     uuid.MustParse(incID),
				Status: "resolved",
			}, nil
		},
	}
	svc := NewIncidentService(repo)

	_, err := svc.AddIncidentUpdate("org-1", incID, userID, "new message")
	if err == nil {
		t.Fatal("expected error when adding update to resolved incident, got nil")
	}
	if err.Error() != "Cannot add update to resolved incident" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestIncidentService_AddIncidentUpdate_NotAssignee(t *testing.T) {
	incID := uuid.New().String()
	userID := uuid.New().String()
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{
				ID:     uuid.MustParse(incID),
				Status: "investigating",
			}, nil
		},
		getAssigneeFunc: func(u, i string) (*model.IncidentAssignee, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewIncidentService(repo)

	_, err := svc.AddIncidentUpdate("org-1", incID, userID, "new message")
	if err == nil {
		t.Fatal("expected error when user is not assignee, got nil")
	}
}

func TestIncidentService_AddIncidentUpdate_Success(t *testing.T) {
	incID := uuid.New().String()
	userID := uuid.New().String()
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{
				ID:     uuid.MustParse(incID),
				Status: "investigating",
			}, nil
		},
		getAssigneeFunc: func(u, i string) (*model.IncidentAssignee, error) {
			return &model.IncidentAssignee{
				IncidentID: uuid.MustParse(incID),
				UserID:     uuid.MustParse(userID),
			}, nil
		},
		createUpdateFunc: func(update *model.IncidentUpdate) error {
			return nil
		},
	}
	svc := NewIncidentService(repo)

	upd, err := svc.AddIncidentUpdate("org-1", incID, userID, "new update message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upd.Message != "new update message" {
		t.Errorf("expected message 'new update message', got %s", upd.Message)
	}
}

func TestIncidentService_JoinAssignee_AlreadyJoined(t *testing.T) {
	incID := uuid.New().String()
	userID := uuid.New().String()
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{ID: uuid.MustParse(incID)}, nil
		},
		getAssigneeFunc: func(u, i string) (*model.IncidentAssignee, error) {
			return &model.IncidentAssignee{IncidentID: uuid.MustParse(incID)}, nil
		},
	}
	svc := NewIncidentService(repo)

	err := svc.JoinAssignee("org-1", incID, userID)
	if err == nil {
		t.Fatal("expected error for already joined assignee, got nil")
	}
}

func TestIncidentService_JoinAssignee_Success(t *testing.T) {
	incID := uuid.New().String()
	userID := uuid.New().String()
	added := false
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return &model.Incident{ID: uuid.MustParse(incID)}, nil
		},
		getAssigneeFunc: func(u, i string) (*model.IncidentAssignee, error) {
			return nil, errors.New("not found")
		},
		addAssigneeFunc: func(assignee *model.IncidentAssignee) error {
			added = true
			return nil
		},
	}
	svc := NewIncidentService(repo)

	err := svc.JoinAssignee("org-1", incID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !added {
		t.Fatal("expected AddAssignee to be called")
	}
}

func TestIncidentService_GetIncidentUpdate_IncidentNotFound(t *testing.T) {
	repo := &mockIncidentRepo{
		getByIdFunc: func(o, i string) (*model.Incident, error) {
			return nil, errors.New("incident not found")
		},
	}
	svc := NewIncidentService(repo)

	_, err := svc.GetIncidentUpdate("org-1", "inc-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ====================== HANDLER TESTS ======================

func TestIncidentHandler_GetIncidentDetail(t *testing.T) {
	orgID := uuid.New()
	handler := NewIncidentHandler(&mockIncidentService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/incident/:incidentId", handler.GetIncidentDetail)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/incident/inc-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIncidentHandler_AddIncidentUpdate_MissingUserContext(t *testing.T) {
	orgID := uuid.New()
	handler := NewIncidentHandler(&mockIncidentService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		// Missing userID in context
		c.Next()
	})
	r.POST("/organization/:orgId/incident/:incidentId/update", handler.AddIncidentUpdate)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/incident/inc-123/update", bytes.NewBufferString(`{"message":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing user context, got %d", w.Code)
	}
}

func TestIncidentHandler_JoinAssignee_Success(t *testing.T) {
	orgID := uuid.New()
	handler := NewIncidentHandler(&mockIncidentService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Set("userID", "user-123")
		c.Next()
	})
	r.POST("/organization/:orgId/incident/:incidentId/join", handler.JoinAssignee)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/incident/inc-123/join", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
