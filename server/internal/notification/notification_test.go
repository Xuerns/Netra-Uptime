package notification

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
type mockChannelRepo struct {
	createFunc     func(channel *model.NotificationChannel) error
	getAllFunc     func(orgID string) ([]model.NotificationChannel, error)
	getByIDFunc    func(orgID, channelID string) (*model.NotificationChannel, error)
	updateFunc     func(channel *model.NotificationChannel) error
	deleteFunc     func(orgID, channelID string) error
	countByOrgFunc func(orgID string) (int64, error)
	createLogFunc  func(log *model.NotificationLog) error
	getLogsFunc    func(orgID string, limit, offset int) ([]model.NotificationLog, int64, error)
}

func (m *mockChannelRepo) Create(channel *model.NotificationChannel) error {
	if m.createFunc != nil {
		return m.createFunc(channel)
	}
	channel.ID = uuid.New()
	channel.CreatedAt = time.Now()
	return nil
}

func (m *mockChannelRepo) GetAll(orgID string) ([]model.NotificationChannel, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(orgID)
	}
	return nil, nil
}

func (m *mockChannelRepo) GetByID(orgID, channelID string) (*model.NotificationChannel, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(orgID, channelID)
	}
	return nil, errors.New("not found")
}

func (m *mockChannelRepo) Update(channel *model.NotificationChannel) error {
	if m.updateFunc != nil {
		return m.updateFunc(channel)
	}
	return nil
}

func (m *mockChannelRepo) Delete(orgID, channelID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(orgID, channelID)
	}
	return nil
}

func (m *mockChannelRepo) CountByOrg(orgID string) (int64, error) {
	if m.countByOrgFunc != nil {
		return m.countByOrgFunc(orgID)
	}
	return 0, nil
}

func (m *mockChannelRepo) CreateLog(log *model.NotificationLog) error {
	if m.createLogFunc != nil {
		return m.createLogFunc(log)
	}
	return nil
}

func (m *mockChannelRepo) GetLogs(orgID string, limit, offset int) ([]model.NotificationLog, int64, error) {
	if m.getLogsFunc != nil {
		return m.getLogsFunc(orgID, limit, offset)
	}
	return nil, 0, nil
}

// ====================== SERVICE UNIT TESTS ======================

func TestValidateConfig_Email_Valid(t *testing.T) {
	err := validateConfig("email", map[string]string{"email": "ops@example.com"})
	if err != nil {
		t.Fatalf("expected valid email, got error: %v", err)
	}
}

func TestValidateConfig_Email_MissingAt(t *testing.T) {
	err := validateConfig("email", map[string]string{"email": "opsexample.com"})
	if err == nil {
		t.Fatal("expected error for email without @, got nil")
	}
}

func TestValidateConfig_Slack_ValidWebhook(t *testing.T) {
	err := validateConfig("slack", map[string]string{
		"webhook_url": "https://hooks.slack.com/services/T00/B00/XXXX",
	})
	if err != nil {
		t.Fatalf("expected valid webhook, got error: %v", err)
	}
}

func TestValidateConfig_Slack_HTTP_Blocked(t *testing.T) {
	err := validateConfig("slack", map[string]string{
		"webhook_url": "http://hooks.slack.com/services/T00/B00/XXXX",
	})
	if err == nil {
		t.Fatal("expected error for HTTP slack webhook, got nil")
	}
}

func TestValidateConfig_Slack_PrivateIP_10x(t *testing.T) {
	err := validateConfig("slack", map[string]string{
		"webhook_url": "https://10.0.0.1/hook",
	})
	if err == nil {
		t.Fatal("expected SSRF error for 10.x.x.x, got nil")
	}
}

func TestValidateConfig_Slack_Localhost(t *testing.T) {
	err := validateConfig("slack", map[string]string{
		"webhook_url": "https://localhost:27346/hook",
	})
	if err == nil {
		t.Fatal("expected SSRF error for localhost, got nil")
	}
}

func TestValidateConfig_Slack_AWSMetadata(t *testing.T) {
	err := validateConfig("slack", map[string]string{
		"webhook_url": "https://169.254.169.254/latest/meta-data",
	})
	if err == nil {
		t.Fatal("expected SSRF error for AWS metadata, got nil")
	}
}

func TestValidateConfig_UnsupportedType(t *testing.T) {
	err := validateConfig("telegram", map[string]string{
		"chat_id": "12345",
	})
	if err == nil {
		t.Fatal("expected error for unsupported channel type, got nil")
	}
}

func TestNotificationService_Create_ExceedsChannelLimit(t *testing.T) {
	repo := &mockChannelRepo{
		countByOrgFunc: func(orgID string) (int64, error) {
			return 10, nil
		},
	}
	svc := NewChannelService(repo)

	req := &CreateChannelRequest{
		Name:   "Slack Alerts",
		Type:   "slack",
		Config: map[string]string{"webhook_url": "https://hooks.slack.com/services/T/B/X"},
	}

	_, err := svc.Create(uuid.New().String(), req)
	if err == nil {
		t.Fatal("expected limit error when count >= 10, got nil")
	}
}

func TestNotificationService_Create_MissingRequiredConfigKey(t *testing.T) {
	repo := &mockChannelRepo{
		countByOrgFunc: func(orgID string) (int64, error) {
			return 0, nil
		},
	}
	svc := NewChannelService(repo)

	req := &CreateChannelRequest{
		Name:   "Slack Alerts",
		Type:   "slack",
		Config: map[string]string{}, // empty config
	}

	_, err := svc.Create(uuid.New().String(), req)
	if err == nil {
		t.Fatal("expected error for missing required config key, got nil")
	}
}

func TestNotificationService_Create_Success(t *testing.T) {
	repo := &mockChannelRepo{
		countByOrgFunc: func(orgID string) (int64, error) {
			return 2, nil
		},
	}
	svc := NewChannelService(repo)

	req := &CreateChannelRequest{
		Name:   "Operations Email",
		Type:   "email",
		Config: map[string]string{"email": "admin@example.com"},
	}

	res, err := svc.Create(uuid.New().String(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "Operations Email" {
		t.Errorf("expected name Operations Email, got %s", res.Name)
	}
}

func TestNotificationService_GetLogs_PaginationDefaults(t *testing.T) {
	var capturedLimit, capturedOffset int
	repo := &mockChannelRepo{
		getLogsFunc: func(orgID string, limit, offset int) ([]model.NotificationLog, int64, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []model.NotificationLog{}, 0, nil
		},
	}
	svc := NewChannelService(repo)

	// Test negative offset clamped to 0, and limit > 100 clamped to 50
	_, err := svc.GetLogs(uuid.New().String(), 150, -5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 50 {
		t.Errorf("expected limit 50, got %d", capturedLimit)
	}
	if capturedOffset != 0 {
		t.Errorf("expected offset 0, got %d", capturedOffset)
	}
}

func TestNotificationService_Delete_NotFound(t *testing.T) {
	repo := &mockChannelRepo{
		getByIDFunc: func(orgID, channelID string) (*model.NotificationChannel, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewChannelService(repo)

	err := svc.Delete(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error when deleting non-existent channel, got nil")
	}
}

// ====================== HANDLER TESTS ======================

func TestChannelHandler_Create_Success(t *testing.T) {
	orgID := uuid.New()
	repo := &mockChannelRepo{
		countByOrgFunc: func(orgID string) (int64, error) {
			return 0, nil
		},
	}
	svc := NewChannelService(repo)
	handler := NewChannelHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.POST("/organization/:orgId/notifications/channels", handler.Create)

	body := CreateChannelRequest{
		Name:   "Slack Alerts",
		Type:   "slack",
		Config: map[string]string{"webhook_url": "https://hooks.slack.com/services/1/2/3"},
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/organization/"+orgID.String()+"/notifications/channels", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestNotificationService_GetAll(t *testing.T) {
	orgID := uuid.New().String()
	repo := &mockChannelRepo{
		getAllFunc: func(org string) ([]model.NotificationChannel, error) {
			return []model.NotificationChannel{
				{ID: uuid.New(), Name: "Slack 1", Type: "slack", Config: `{"webhook_url":"https://hooks.slack.com/1"}`},
				{ID: uuid.New(), Name: "Email 1", Type: "email", Config: `{"email":"admin@corp.com"}`},
			}, nil
		},
	}
	svc := NewChannelService(repo)
	channels, err := svc.GetAll(orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
}

func TestNotificationService_GetByID(t *testing.T) {
	orgID := uuid.New().String()
	chID := uuid.New().String()
	repo := &mockChannelRepo{
		getByIDFunc: func(org, ch string) (*model.NotificationChannel, error) {
			return &model.NotificationChannel{
				ID:     uuid.MustParse(chID),
				Name:   "Slack Channel",
				Type:   "slack",
				Config: `{"webhook_url":"https://hooks.slack.com/xyz"}`,
			}, nil
		},
	}
	svc := NewChannelService(repo)
	res, err := svc.GetByID(orgID, chID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "Slack Channel" {
		t.Errorf("expected Slack Channel, got %s", res.Name)
	}
}

func TestNotificationService_Update(t *testing.T) {
	orgID := uuid.New().String()
	chID := uuid.New().String()
	existing := &model.NotificationChannel{
		ID:     uuid.MustParse(chID),
		Name:   "Old Name",
		Type:   "slack",
		Config: `{"webhook_url":"https://hooks.slack.com/xyz"}`,
	}
	repo := &mockChannelRepo{
		getByIDFunc: func(org, ch string) (*model.NotificationChannel, error) {
			return existing, nil
		},
		updateFunc: func(ch *model.NotificationChannel) error {
			return nil
		},
	}
	svc := NewChannelService(repo)

	newName := "New Slack Name"
	updated, err := svc.Update(orgID, chID, &UpdateChannelRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New Slack Name" {
		t.Errorf("expected New Slack Name, got %s", updated.Name)
	}
}

func TestChannelHandler_GetAll(t *testing.T) {
	orgID := uuid.New()
	repo := &mockChannelRepo{
		getAllFunc: func(orgID string) ([]model.NotificationChannel, error) {
			return []model.NotificationChannel{}, nil
		},
	}
	svc := NewChannelService(repo)
	handler := NewChannelHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/notifications/channels", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/notifications/channels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestChannelHandler_Delete(t *testing.T) {
	orgID := uuid.New()
	channelID := uuid.New()
	repo := &mockChannelRepo{
		getByIDFunc: func(orgID, chID string) (*model.NotificationChannel, error) {
			return &model.NotificationChannel{ID: channelID}, nil
		},
		deleteFunc: func(orgID, chID string) error {
			return nil
		},
	}
	svc := NewChannelService(repo)
	handler := NewChannelHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.DELETE("/organization/:orgId/notifications/channels/:channelId", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/organization/"+orgID.String()+"/notifications/channels/"+channelID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestChannelHandler_GetLogs(t *testing.T) {
	orgID := uuid.New()
	repo := &mockChannelRepo{
		getLogsFunc: func(orgID string, limit, offset int) ([]model.NotificationLog, int64, error) {
			return []model.NotificationLog{}, 0, nil
		},
	}
	svc := NewChannelService(repo)
	handler := NewChannelHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("organizationID", orgID)
		c.Next()
	})
	r.GET("/organization/:orgId/notifications/logs", handler.GetLogs)

	req, _ := http.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/notifications/logs?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
