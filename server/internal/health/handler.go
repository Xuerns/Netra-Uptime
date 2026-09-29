package health

import (
	"net/http"
	"time"

	"PulseWatch/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb, cfg: cfg}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Redis     string    `json:"redis"`
	AppMode   string    `json:"app_mode"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "connected"
	redisStatus := "connected"
	isHealthy := true

	// Check Postgres
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

	// Check Redis
	if h.rdb != nil {
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus = "unreachable"
			isHealthy = false
		}
	} else {
		redisStatus = "unreachable"
		isHealthy = false
	}

	status := "healthy"
	httpCode := http.StatusOK
	if !isHealthy {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	appMode := "selfhosted"
	if h.cfg != nil && h.cfg.AppMode != "" {
		appMode = h.cfg.AppMode
	}

	c.JSON(httpCode, HealthResponse{
		Status:    status,
		Database:  dbStatus,
		Redis:     redisStatus,
		AppMode:   appMode,
		Timestamp: time.Now().UTC(),
	})
}
