package router

import (
	"PulseWatch/internal/alertrule"
	"PulseWatch/internal/apikey"
	"PulseWatch/internal/audit"
	"PulseWatch/internal/auth"
	"PulseWatch/internal/incident"
	"PulseWatch/internal/model"
	"PulseWatch/internal/monitor"
	"PulseWatch/internal/notification"
	"PulseWatch/internal/organization"
	"PulseWatch/internal/publicapi"
	"PulseWatch/internal/statuspage"
	"PulseWatch/internal/websocket"
	"PulseWatch/pkg/jwt"
	"PulseWatch/pkg/utils"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("APP_MODE", "saas")
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "super-secret-key-for-test-purposes-123456")
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nohtml", utils.NoHTMLValidator)
	}
}

// ====================== MOCK REDIS FOR RATE LIMITER ======================

func newMockRedisClient(t *testing.T) (*redis.Client, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on random port: %v", err)
	}

	store := make(map[string]string)
	var mu sync.Mutex

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				r := bufio.NewReader(c)
				for {
					line, err := r.ReadString('\n')
					if err != nil {
						return
					}
					line = strings.TrimSpace(line)
					if len(line) == 0 {
						continue
					}

					var args []string
					if strings.HasPrefix(line, "*") {
						numArgs, _ := strconv.Atoi(line[1:])
						for i := 0; i < numArgs; i++ {
							lenLine, err := r.ReadString('\n')
							if err != nil {
								return
							}
							argLen, _ := strconv.Atoi(strings.TrimSpace(lenLine[1:]))
							buf := make([]byte, argLen+2)
							if _, err := io.ReadFull(r, buf); err != nil {
								return
							}
							args = append(args, string(buf[:argLen]))
						}
					} else {
						args = strings.Fields(line)
					}

					if len(args) == 0 {
						continue
					}

					cmd := strings.ToUpper(args[0])
					switch cmd {
					case "PING":
						c.Write([]byte("+PONG\r\n"))
					case "INCR":
						mu.Lock()
						val, _ := strconv.Atoi(store[args[1]])
						val++
						store[args[1]] = strconv.Itoa(val)
						mu.Unlock()
						fmt.Fprintf(c, ":%d\r\n", val)
					case "EXPIRE":
						c.Write([]byte(":1\r\n"))
					case "GET":
						mu.Lock()
						val, ok := store[args[1]]
						mu.Unlock()
						if !ok {
							c.Write([]byte("$-1\r\n"))
						} else {
							fmt.Fprintf(c, "$%d\r\n%s\r\n", len(val), val)
						}
					case "SET":
						mu.Lock()
						store[args[1]] = args[2]
						mu.Unlock()
						c.Write([]byte("+OK\r\n"))
					case "DEL":
						mu.Lock()
						for _, k := range args[1:] {
							delete(store, k)
						}
						mu.Unlock()
						c.Write([]byte(":1\r\n"))
					default:
						c.Write([]byte("-ERR unknown command\r\n"))
					}
				}
			}(conn)
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr:     l.Addr().String(),
		Protocol: 2,
	})

	cleanup := func() {
		client.Close()
		l.Close()
	}

	return client, cleanup
}

// ====================== MOCK SQL DRIVER FOR GORM ======================

var (
	mockMembersMu sync.RWMutex
	mockMembers   = make(map[string]string) // "orgID:userID" -> role ("owner", "admin", "member")
)

func setMockMember(orgID, userID, role string) {
	mockMembersMu.Lock()
	defer mockMembersMu.Unlock()
	mockMembers[orgID+":"+userID] = role
}

type mockDriverConnector struct{}

func (c *mockDriverConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &mockDBConn{}, nil
}
func (c *mockDriverConnector) Driver() driver.Driver {
	return &mockDriver{}
}

type mockDriver struct{}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockDBConn{}, nil
}

type mockDBConn struct{}

func (c *mockDBConn) Prepare(query string) (driver.Stmt, error) {
	return &mockDBStmt{query: query}, nil
}
func (c *mockDBConn) Close() error { return nil }
func (c *mockDBConn) Begin() (driver.Tx, error) {
	return &mockDBTx{}, nil
}

type mockDBTx struct{}

func (t *mockDBTx) Commit() error   { return nil }
func (t *mockDBTx) Rollback() error { return nil }

type mockDBStmt struct {
	query string
}

func (s *mockDBStmt) Close() error  { return nil }
func (s *mockDBStmt) NumInput() int { return -1 }
func (s *mockDBStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *mockDBStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := strings.ToUpper(s.query)
	if strings.Contains(q, "VERSION()") {
		return &mockDBRows{
			cols: []string{"version"},
			vals: [][]driver.Value{{"PostgreSQL 15.0"}},
		}, nil
	}

	if strings.Contains(q, "ORGANIZATION_MEMBERS") {
		if len(args) >= 2 {
			orgID := fmt.Sprintf("%v", args[0])
			userID := fmt.Sprintf("%v", args[1])

			mockMembersMu.RLock()
			role, found := mockMembers[orgID+":"+userID]
			mockMembersMu.RUnlock()

			if found {
				return &mockDBRows{
					cols: []string{"id", "organization_id", "user_id", "role", "created_at"},
					vals: [][]driver.Value{{uuid.New().String(), orgID, userID, role, time.Now()}},
				}, nil
			}
		}
		// Not found -> return empty rows
		return &mockDBRows{
			cols: []string{"id", "organization_id", "user_id", "role", "created_at"},
			vals: [][]driver.Value{},
		}, nil
	}

	return &mockDBRows{
		cols: []string{"col"},
		vals: [][]driver.Value{},
	}, nil
}

type mockDBRows struct {
	cols []string
	vals [][]driver.Value
	pos  int
}

func (r *mockDBRows) Columns() []string { return r.cols }
func (r *mockDBRows) Close() error      { return nil }
func (r *mockDBRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.vals) {
		return io.EOF
	}
	for i, v := range r.vals[r.pos] {
		dest[i] = v
	}
	r.pos++
	return nil
}

func setupMockGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB := sql.OpenDB(&mockDriverConnector{})
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to create mock gorm DB: %v", err)
	}
	return db
}

// ====================== ROUTER TEST SUITE HELPER ======================

type testEnv struct {
	router *gin.Engine
	db     *gorm.DB
}

func setupTestRouter(t *testing.T) (*testEnv, func()) {
	t.Helper()
	r := gin.New()
	db := setupMockGormDB(t)
	rdb, cleanupRedis := newMockRedisClient(t)

	// Mock Services
	mockAuthSvc := &mockAuthService{}
	mockOrgSvc := &mockOrgService{}
	mockMonSvc := &mockMonService{}
	mockIncSvc := &mockIncService{}
	mockStatSvc := &mockStatService{}
	mockAudSvc := &mockAudService{}
	mockKeySvc := &mockKeyService{}
	mockNotifSvc := &mockNotifService{}
	mockAlertSvc := &mockAlertService{}

	h := Handlers{
		AuthHandler:         auth.NewAuthHandler(mockAuthSvc),
		OrganizationHandler: organization.NewOrganizationHandler(mockOrgSvc),
		MonitorHandler:      monitor.NewMonitorHandler(mockMonSvc),
		IncidentHandler:     incident.NewIncidentHandler(mockIncSvc),
		StatusPageHandler:   statuspage.NewStatusPageHandler(mockStatSvc),
		AuditHandler:        audit.NewAuditHandler(mockAudSvc),
		APIKeyHandler:       apikey.NewAPIKeyHandler(mockKeySvc),
		PublicAPIHandler:    publicapi.NewPublicAPIHandler(mockMonSvc, mockIncSvc, mockStatSvc),
		APIKeyService:       mockKeySvc,
		NotificationHandler: notification.NewChannelHandler(mockNotifSvc),
		AlertRuleHandler:    alertrule.NewHandler(mockAlertSvc),
	}

	wsHandler := &websocket.WsHandler{}

	SetupRouter(r, db, rdb, h, wsHandler)

	cleanup := func() {
		cleanupRedis()
	}

	return &testEnv{
		router: r,
		db:     db,
	}, cleanup
}

func createValidToken(userID string) string {
	tok, _ := jwt.GenerateToken(userID, time.Hour)
	return tok
}

func executeRequest(r *gin.Engine, method, url string, body []byte, token string, extraHeaders map[string]string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, _ := http.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: token,
		})
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Mock service structs for SetupRouter handlers
type mockAuthService struct{}

func (m *mockAuthService) Register(req auth.RegisterRequest) (string, error) { return "ok", nil }
func (m *mockAuthService) Login(req auth.LoginRequest) (*auth.LoginResponse, error) {
	return &auth.LoginResponse{AccessToken: "at", RefreshToken: "rt"}, nil
}
func (m *mockAuthService) OauthLoginAndRegister(req auth.OAuthRequest) (*auth.LoginResponse, error) {
	return nil, nil
}
func (m *mockAuthService) Logout(refreshToken string) error             { return nil }
func (m *mockAuthService) RefreshAccessToken(token string) (string, error) { return "at", nil }
func (m *mockAuthService) GetProfile(userId string) (*auth.ProfileResponse, error) {
	return &auth.ProfileResponse{ID: userId}, nil
}
func (m *mockAuthService) UpdateProfile(userId string, req auth.UpdateProfileRequest) (*auth.ProfileResponse, error) {
	return &auth.ProfileResponse{ID: userId, Name: req.Name}, nil
}
func (m *mockAuthService) ForgotPassword(email string) (string, error) { return "", nil }
func (m *mockAuthService) ResetPassword(newPassword, resetToken string) error { return nil }
func (m *mockAuthService) InitPasswordChange(userId string, req auth.InitPasswordChangeRequest) error {
	return nil
}
func (m *mockAuthService) VerifyPasswordChange(userId string, req auth.VerifyPasswordChangeRequest) error {
	return nil
}

type mockOrgService struct{}

func (m *mockOrgService) GetAllOrganization(userId string) (*[]organization.GetAllOrganizationResponse, error) {
	return &[]organization.GetAllOrganizationResponse{}, nil
}
func (m *mockOrgService) GetOrganizationAllMember(orgId, memberId string) (*[]organization.MemberResponse, error) {
	return &[]organization.MemberResponse{}, nil
}
func (m *mockOrgService) CreateOrganization(req *organization.CreateOrganizationRequest, ownerMemberId string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) GetInvitation(userId string) (*[]organization.InvitationResponse, error) {
	return &[]organization.InvitationResponse{}, nil
}
func (m *mockOrgService) KickMember(orgId, actorId, memberId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) InvitationAction(req *organization.InvitationActionRequest, userId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) InviteMember(orgId, ownerId, ipAddress string, invitations organization.BulkInviteMemberRequest) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) Disband(orgId, ownerId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) UpdateOrganization(req *organization.UpdateOrganizationRequest, orgId, actorId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) SearchPersons(name string, excludeUserId string, page int, orgId string) ([]organization.SearchResponse, error) {
	return nil, nil
}
func (m *mockOrgService) PromoteToAdmin(userId, orgId, actorId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) DemotedToMember(userId, orgId, actorId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) TransferOwnership(ownerId, userId, orgId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) LeaveOrganization(userId, orgId, ipAddress string) (string, error) {
	return "ok", nil
}
func (m *mockOrgService) GetOrganizationPlan(orgId string) (*model.Subscription, error) {
	return &model.Subscription{Plan: "hobby"}, nil
}

type mockMonService struct{}

func (m *mockMonService) Create(orgId, userId, ipAddress string, req *monitor.CreateMonitorRequest) (*model.Monitor, error) {
	return &model.Monitor{Name: req.Name}, nil
}
func (m *mockMonService) GetById(orgId, monitorId string) (*monitor.MonitorResponse, error) {
	return &monitor.MonitorResponse{ID: monitorId}, nil
}
func (m *mockMonService) GetAllMonitor(orgId string) ([]monitor.MonitorResponse, error) {
	return []monitor.MonitorResponse{}, nil
}
func (m *mockMonService) UpdateMonitor(orgId, userId, ipAddress, monitorId string, req monitor.UpdatedMonitorRequest) (*model.Monitor, error) {
	return &model.Monitor{}, nil
}
func (m *mockMonService) DeleteMonitor(orgId, userId, ipAddress, monitorId string) error {
	return nil
}
func (m *mockMonService) GetAllMonitorByUserId(userId string) ([]monitor.MonitorResponse, error) {
	return []monitor.MonitorResponse{}, nil
}
func (m *mockMonService) GetDailyStats(orgId, monitorId string, days int) ([]model.MonitorDailyStat, error) {
	return []model.MonitorDailyStat{}, nil
}
func (m *mockMonService) TriggerCheck(orgId, monitorId string) (*monitor.RecentCheck, error) {
	return &monitor.RecentCheck{ID: "chk-1"}, nil
}

type mockIncService struct{}

func (m *mockIncService) GetIncidentDetail(orgId, incidentId string) (*incident.IncidentResponse, error) {
	return &incident.IncidentResponse{ID: incidentId}, nil
}
func (m *mockIncService) GetAllIncident(orgId string) ([]incident.IncidentResponse, error) {
	return []incident.IncidentResponse{}, nil
}
func (m *mockIncService) AddIncidentUpdate(orgid, incidentId, userId, message string) (*model.IncidentUpdate, error) {
	return &model.IncidentUpdate{Message: message}, nil
}
func (m *mockIncService) JoinAssignee(orgId, incidentId, userId string) error { return nil }
func (m *mockIncService) GetIncidentUpdate(orgId, incidentId string) ([]incident.GetIncidentUpdateResponse, error) {
	return []incident.GetIncidentUpdateResponse{}, nil
}

type mockStatService struct{}

func (m *mockStatService) Create(orgID, userID, ipAddress string, req *statuspage.CreateStatusPageRequest) (*model.StatusPage, error) {
	return &model.StatusPage{Title: req.Title}, nil
}
func (m *mockStatService) GetAllByOrganizationID(orgID, userID string) ([]statuspage.StatusPageResponse, error) {
	return []statuspage.StatusPageResponse{}, nil
}
func (m *mockStatService) Delete(statusPageID, orgID, UserID, ipAddress string) error { return nil }
func (m *mockStatService) Update(statuspageID, orgID, userID, ipAddress string, req *statuspage.UpdateStatusPageRequest) (*model.StatusPage, error) {
	return &model.StatusPage{}, nil
}
func (m *mockStatService) GetById(statusPageID, orgID, userID string) (*statuspage.StatusPageResponse, error) {
	return &statuspage.StatusPageResponse{ID: statusPageID}, nil
}
func (m *mockStatService) GetBySubdomain(subdomain string) (*statuspage.PublicStatusPageResponse, error) {
	return &statuspage.PublicStatusPageResponse{Subdomain: subdomain}, nil
}
func (m *mockStatService) GetStatusPageMonitor(statusPageID string) ([]model.StatusPageMonitor, error) {
	return []model.StatusPageMonitor{}, nil
}
func (m *mockStatService) GetStatusPageByUserId(userId string) ([]statuspage.StatusPageResponse, error) {
	return []statuspage.StatusPageResponse{}, nil
}
func (m *mockStatService) CreateMaintenance(orgID, statusPageID, userID, ipAddress string, req *statuspage.CreateMaintenanceRequest) (*model.ScheduledMaintenance, error) {
	return &model.ScheduledMaintenance{Title: req.Title}, nil
}
func (m *mockStatService) GetMaintenances(orgID, statusPageID, userID string) ([]statuspage.MaintenanceResponse, error) {
	return []statuspage.MaintenanceResponse{}, nil
}
func (m *mockStatService) UpdateMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string, req *statuspage.UpdateMaintenanceRequest) (*model.ScheduledMaintenance, error) {
	return &model.ScheduledMaintenance{}, nil
}
func (m *mockStatService) DeleteMaintenance(orgID, statusPageID, maintenanceID, userID, ipAddress string) error {
	return nil
}
func (m *mockStatService) CreateComponentGroup(orgID, statusPageID, userID, ipAddress string, req *statuspage.CreateComponentGroupRequest) (*model.ComponentGroup, error) {
	return &model.ComponentGroup{Name: req.Name}, nil
}
func (m *mockStatService) GetComponentGroups(orgID, statusPageID, userID string) ([]statuspage.ComponentGroupResponse, error) {
	return []statuspage.ComponentGroupResponse{}, nil
}
func (m *mockStatService) UpdateComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string, req *statuspage.UpdateComponentGroupRequest) (*model.ComponentGroup, error) {
	return &model.ComponentGroup{}, nil
}
func (m *mockStatService) DeleteComponentGroup(orgID, statusPageID, groupID, userID, ipAddress string) error {
	return nil
}
func (m *mockStatService) AssignMonitorToGroup(orgID, statusPageID, monitorID, userID, ipAddress string, req *statuspage.AssignMonitorGroupRequest) error {
	return nil
}

type mockAudService struct{}

func (m *mockAudService) Record(params audit.RecordAuditLogParams) error { return nil }
func (m *mockAudService) GetLogs(orgID string, query audit.GetAuditLogsQuery) (*audit.PaginatedAuditLogsResponse, error) {
	return &audit.PaginatedAuditLogsResponse{}, nil
}

type mockKeyService struct{}

func (m *mockKeyService) CreateKey(orgID, userID string, req apikey.CreateAPIKeyRequest, ip string) (*apikey.CreatedAPIKeyResponse, error) {
	return &apikey.CreatedAPIKeyResponse{}, nil
}
func (m *mockKeyService) ListKeys(orgID string) ([]apikey.APIKeyResponse, error) {
	return []apikey.APIKeyResponse{}, nil
}
func (m *mockKeyService) RevokeKey(orgID, userID, keyID string, ip string) error { return nil }
func (m *mockKeyService) ValidateKey(plaintextKey string) (*model.APIKey, error) {
	if plaintextKey == "pw_live_valid_read" {
		return &model.APIKey{ID: uuid.New(), OrganizationID: uuid.New(), Permissions: "read"}, nil
	}
	if plaintextKey == "pw_live_valid_write" {
		return &model.APIKey{ID: uuid.New(), OrganizationID: uuid.New(), Permissions: "write"}, nil
	}
	if plaintextKey == "pw_live_expired" {
		return nil, errors.New("API key has expired")
	}
	return nil, errors.New("invalid API key")
}
func (m *mockKeyService) TouchLastUsed(keyID string, ip string) error { return nil }

type mockNotifService struct{}

func (m *mockNotifService) Create(orgID string, req *notification.CreateChannelRequest) (*notification.ChannelResponse, error) {
	return &notification.ChannelResponse{Name: req.Name}, nil
}
func (m *mockNotifService) GetAll(orgID string) ([]notification.ChannelResponse, error) {
	return []notification.ChannelResponse{}, nil
}
func (m *mockNotifService) GetByID(orgID, channelID string) (*notification.ChannelResponse, error) {
	return &notification.ChannelResponse{ID: channelID}, nil
}
func (m *mockNotifService) Update(orgID, channelID string, req *notification.UpdateChannelRequest) (*notification.ChannelResponse, error) {
	return &notification.ChannelResponse{}, nil
}
func (m *mockNotifService) Delete(orgID, channelID string) error { return nil }
func (m *mockNotifService) GetLogs(orgID string, limit, offset int) (*notification.LogListResponse, error) {
	return &notification.LogListResponse{}, nil
}

type mockAlertService struct{}

func (m *mockAlertService) Create(orgID, monitorID uuid.UUID, req *alertrule.CreateAlertRuleRequest) (*alertrule.AlertRuleResponse, error) {
	return &alertrule.AlertRuleResponse{ID: uuid.New()}, nil
}
func (m *mockAlertService) GetByMonitor(orgID, monitorID uuid.UUID) ([]alertrule.AlertRuleResponse, error) {
	return []alertrule.AlertRuleResponse{}, nil
}
func (m *mockAlertService) GetByID(orgID, ruleID uuid.UUID) (*alertrule.AlertRuleResponse, error) {
	return &alertrule.AlertRuleResponse{ID: ruleID}, nil
}
func (m *mockAlertService) Update(orgID, ruleID uuid.UUID, req *alertrule.UpdateAlertRuleRequest) (*alertrule.AlertRuleResponse, error) {
	return &alertrule.AlertRuleResponse{ID: ruleID}, nil
}
func (m *mockAlertService) Delete(orgID, ruleID uuid.UUID) error { return nil }

// ==============================================================================
// 1. ❌ WRONG HTTP METHOD (404 / 405)
// ==============================================================================

func TestBadRoute_RegisterWithGET(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/register", nil, "", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405 for GET register, got %d", w.Code)
	}
}

func TestBadRoute_LoginWithDELETE(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodDelete, "/api/v1/auth/login", nil, "", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405 for DELETE login, got %d", w.Code)
	}
}

func TestBadRoute_MonitorsWithPATCH(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPatch, "/api/v1/organization/org-123/monitors", nil, "", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405 for PATCH monitors, got %d", w.Code)
	}
}

func TestBadRoute_ProfileWithPOST(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/profile", nil, "", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405 for POST profile, got %d", w.Code)
	}
}

// ==============================================================================
// 2. ❌ ROUTE TIDAK ADA (404)
// ==============================================================================

func TestBadRoute_NonExistentEndpoint(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/nonexistent", nil, "", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent endpoint, got %d", w.Code)
	}
}

func TestBadRoute_OldAPIVersion(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v2/auth/login", nil, "", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for v2 API, got %d", w.Code)
	}
}

func TestBadRoute_RandomPath(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/random/path/here", nil, "", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for /api/v1 random path, got %d", w.Code)
	}
}

// ==============================================================================
// 3. 🔐 UNAUTHORIZED — PROTECTED ROUTES TANPA TOKEN (401)
// ==============================================================================

func TestBadRoute_ProfileWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/profile", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for profile without token, got %d", w.Code)
	}
}

func TestBadRoute_OrgListWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/all", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for org list without token, got %d", w.Code)
	}
}

func TestBadRoute_MonitorListWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/monitors", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for monitor list without token, got %d", w.Code)
	}
}

func TestBadRoute_IncidentWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/incident", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for incident list without token, got %d", w.Code)
	}
}

func TestBadRoute_NotificationWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/notifications/channels", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for notification without token, got %d", w.Code)
	}
}

func TestBadRoute_APIKeyWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/api-keys", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for api-keys without token, got %d", w.Code)
	}
}

func TestBadRoute_StatusPageWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/statuspage", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for statuspage without token, got %d", w.Code)
	}
}

func TestBadRoute_AuditLogWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/123/audit-logs", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for audit-logs without token, got %d", w.Code)
	}
}

func TestBadRoute_PasswordChangeWithoutToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/password/change/init", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for password change init without token, got %d", w.Code)
	}
}

// ==============================================================================
// 4. 🔐 INVALID TOKEN (401)
// ==============================================================================

func TestBadRoute_ProfileWithExpiredToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	token, _ := jwt.GenerateToken("user-123", -1*time.Hour)
	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/profile", nil, token, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

func TestBadRoute_ProfileWithMalformedToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/profile", nil, "invalid.jwt.token", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for malformed token, got %d", w.Code)
	}
}

func TestBadRoute_ProfileWithWrongSignature(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	os.Setenv("JWT_SECRET", "secret-other")
	token, _ := jwt.GenerateToken("user-123", time.Hour)
	os.Setenv("JWT_SECRET", "super-secret-key-for-test-purposes-123456")

	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/profile", nil, token, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong signature key, got %d", w.Code)
	}
}

func TestBadRoute_ProfileWithEmptyCookie(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/auth/profile", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for empty cookie, got %d", w.Code)
	}
}

// ==============================================================================
// 5. 🏢 CROSS-ORG ACCESS (403)
// ==============================================================================

func TestBadRoute_AccessOtherOrgMonitors(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	foreignOrgID := uuid.New().String()
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/"+foreignOrgID+"/monitors", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for cross-org monitor access, got %d", w.Code)
	}
}

func TestBadRoute_AccessOtherOrgIncidents(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	foreignOrgID := uuid.New().String()
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/"+foreignOrgID+"/incident", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for cross-org incident access, got %d", w.Code)
	}
}

func TestBadRoute_AccessOtherOrgNotifications(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	foreignOrgID := uuid.New().String()
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/"+foreignOrgID+"/notifications/channels", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for cross-org notifications access, got %d", w.Code)
	}
}

func TestBadRoute_AccessOtherOrgAPIKeys(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	foreignOrgID := uuid.New().String()
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/"+foreignOrgID+"/api-keys", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for cross-org api-keys access, got %d", w.Code)
	}
}

// ==============================================================================
// 6. 👮 INSUFFICIENT ROLE (403)
// ==============================================================================

func TestBadRoute_MemberCannotCreateMonitor(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member") // only member!
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/monitors", []byte(`{}`), token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member creating monitor, got %d", w.Code)
	}
}

func TestBadRoute_MemberCannotInvite(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/invite", []byte(`{}`), token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member inviting, got %d", w.Code)
	}
}

func TestBadRoute_MemberCannotDisband(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodDelete, "/api/v1/organization/"+orgID+"/disband", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member disbanding org, got %d", w.Code)
	}
}

func TestBadRoute_MemberCannotCreateAPIKey(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/api-keys", []byte(`{}`), token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member creating api key, got %d", w.Code)
	}
}

func TestBadRoute_MemberCannotCreateNotifChannel(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/notifications/channels", []byte(`{}`), token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member creating channel, got %d", w.Code)
	}
}

func TestBadRoute_MemberCannotKick(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "member")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodDelete, "/api/v1/organization/"+orgID+"/kick", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for member kicking someone, got %d", w.Code)
	}
}

func TestBadRoute_AdminCannotDisband(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "admin") // admin is not owner!
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodDelete, "/api/v1/organization/"+orgID+"/disband", nil, token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for admin disbanding org, got %d", w.Code)
	}
}

func TestBadRoute_AdminCannotTransferOwnership(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "admin")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPut, "/api/v1/organization/"+orgID+"/transfer", []byte(`{}`), token, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for admin transferring ownership, got %d", w.Code)
	}
}

// ==============================================================================
// 7. 🆔 INVALID URL PARAMS (400 / 403 / 500)
// ==============================================================================

func TestBadRoute_InvalidMonitorUUID(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "owner")
	token := createValidToken(userID)

	// monitorId is not a uuid for alert-rules creation
	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/monitors/not-a-uuid/alert-rules", []byte(`{}`), token, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid monitorId UUID, got %d", w.Code)
	}
}

func TestBadRoute_InvalidOrgUUID(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	token := createValidToken(userID)

	// Tenant middleware fails to find member for invalid org -> 403
	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/invalid-org-id/monitors", nil, token, nil)
	if w.Code != http.StatusForbidden && w.Code != http.StatusBadRequest {
		t.Errorf("expected 403/400 for invalid orgId, got %d", w.Code)
	}
}

func TestBadRoute_InvalidRuleID(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "owner")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodGet, "/api/v1/organization/"+orgID+"/alert-rules/not-a-valid-uuid", nil, token, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid alert ruleId, got %d", w.Code)
	}
}

// ==============================================================================
// 8. 📦 MALFORMED REQUEST BODY (400)
// ==============================================================================

func TestBadRoute_RegisterEmptyBody(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/register", []byte(`{}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body register, got %d", w.Code)
	}
}

func TestBadRoute_RegisterMissingEmail(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/register", []byte(`{"name":"Alice","password":"password123"}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for register missing email, got %d", w.Code)
	}
}

func TestBadRoute_RegisterShortPassword(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/register", []byte(`{"name":"Alice","email":"alice@test.com","password":"short"}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for register short password (<8), got %d", w.Code)
	}
}

func TestBadRoute_LoginEmptyBody(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/login", []byte(`{}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body login, got %d", w.Code)
	}
}

func TestBadRoute_LoginInvalidEmail(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/login", []byte(`{"email":"notanemail","password":"password123"}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for login invalid email, got %d", w.Code)
	}
}

func TestBadRoute_ResetPasswordMissingToken(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPatch, "/api/v1/auth/resetpassword", []byte(`{"new_password":"newpassword123"}`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for resetpassword without token, got %d", w.Code)
	}
}

func TestBadRoute_CreateMonitorMissingFields(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	setMockMember(orgID, userID, "owner")
	token := createValidToken(userID)

	w := executeRequest(env.router, http.MethodPost, "/api/v1/organization/"+orgID+"/monitors", []byte(`{"name":"Short"}`), token, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for monitor missing fields, got %d", w.Code)
	}
}

func TestBadRoute_PlainTextBody(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("hello text world"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for text/plain body, got %d", w.Code)
	}
}

func TestBadRoute_BrokenJSON(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodPost, "/api/v1/auth/register", []byte(`{broken json`), "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for broken JSON, got %d", w.Code)
	}
}

// ==============================================================================
// 9. 🔑 PUBLIC API KEY AUTH (401 / 403)
// ==============================================================================

func TestBadRoute_PublicAPI_NoAPIKey(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/public/monitors", nil, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for public api without api key, got %d", w.Code)
	}
}

func TestBadRoute_PublicAPI_InvalidAPIKey(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/public/monitors", nil, "", map[string]string{
		"X-API-Key": "pw_live_invalid_key_xyz",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid api key, got %d", w.Code)
	}
}

func TestBadRoute_PublicAPI_ReadKeyOnWriteRoute(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	// POST monitors requires 'write' permission, but key only has 'read'
	w := executeRequest(env.router, http.MethodPost, "/api/v1/public/monitors", []byte(`{}`), "", map[string]string{
		"X-API-Key": "pw_live_valid_read",
	})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for read key accessing write route, got %d", w.Code)
	}
}

func TestBadRoute_PublicAPI_ExpiredKey(t *testing.T) {
	env, cleanup := setupTestRouter(t)
	defer cleanup()
	w := executeRequest(env.router, http.MethodGet, "/api/v1/public/monitors", nil, "", map[string]string{
		"X-API-Key": "pw_live_expired",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired api key, got %d", w.Code)
	}
}
