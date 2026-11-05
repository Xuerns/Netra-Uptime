package health

import (
	"net/http"
	"time"

	"netra-uptime/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewHealthHandler(db *gorm.DB, cfg *config.Config) *HealthHandler {
	return &HealthHandler{db: db, cfg: cfg}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	DBDriver  string    `json:"db_driver"`
	AppMode   string    `json:"app_mode"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "connected"
	isHealthy := true

	// Check Database
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			dbStatus = "unreachable"
			isHealthy = false
		}
	} else {
		dbStatus = "unreachable"
		isHealthy = false
	}

	dbDriver := "postgres"
	if h.cfg != nil && h.cfg.DBDriver != "" {
		dbDriver = h.cfg.DBDriver
	}

	status := "ok"
	httpCode := http.StatusOK
	if !isHealthy {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	c.JSON(httpCode, HealthResponse{
		Status:    status,
		Database:  dbStatus,
		DBDriver:  dbDriver,
		AppMode:   "selfhosted",
		Timestamp: time.Now().UTC(),
	})
}
