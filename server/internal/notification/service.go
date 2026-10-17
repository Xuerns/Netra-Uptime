package notification

import (
	"netra-uptime/internal/model"
	"netra-uptime/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// maxChannelsPerOrg is a hard cap to prevent abuse.
const maxChannelsPerOrg = 10

// ─── Required config keys per channel type ───────────────────────────────────

var requiredConfigKeys = map[string][]string{
	"email":   {"email"},
	"slack":   {"webhook_url"},
	"webhook": {"url"},
	"discord": {"webhook_url"},
}

// sensitiveConfigKeys are masked in API responses so tokens never leak.
var sensitiveConfigKeys = map[string]bool{
	"webhook_url": true,
	"url":         true,
}

// ─── Interface ───────────────────────────────────────────────────────────────

type ChannelService interface {
	Create(orgID string, req *CreateChannelRequest) (*ChannelResponse, error)
	GetAll(orgID string) ([]ChannelResponse, error)
	GetByID(orgID, channelID string) (*ChannelResponse, error)
	Update(orgID, channelID string, req *UpdateChannelRequest) (*ChannelResponse, error)
	Delete(orgID, channelID string) error
	GetLogs(orgID string, limit, offset int) (*LogListResponse, error)
}

// ─── Implementation ──────────────────────────────────────────────────────────

type channelServiceImpl struct {
	repo ChannelRepository
}

func NewChannelService(repo ChannelRepository) ChannelService {
	return &channelServiceImpl{repo: repo}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func (s *channelServiceImpl) Create(orgID string, req *CreateChannelRequest) (*ChannelResponse, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("invalid organization ID")
	}

	// Enforce per-org channel cap
	count, err := s.repo.CountByOrg(orgID)
	if err != nil {
		return nil, errors.New("failed to check channel count")
	}
	if count >= maxChannelsPerOrg {
		return nil, fmt.Errorf("maximum of %d notification channels allowed per organization", maxChannelsPerOrg)
	}

	// Validate config keys and values for the requested type
	if err := validateConfig(req.Type, req.Config); err != nil {
		return nil, err
	}

	configJSON, err := marshalConfig(req.Config)
	if err != nil {
		return nil, errors.New("failed to serialize config")
	}

	ch := &model.NotificationChannel{
		OrganizationID: orgUUID,
		Name:           req.Name,
		Type:           req.Type,
		Config:         configJSON,
		IsActive:       true,
	}

	if err := s.repo.Create(ch); err != nil {
		return nil, errors.New("failed to create notification channel")
	}

	return mapChannelToResponse(ch, req.Config), nil
}


func (s *channelServiceImpl) GetAll(orgID string) ([]ChannelResponse, error) {
	channels, err := s.repo.GetAll(orgID)
	if err != nil {
		return nil, errors.New("failed to fetch notification channels")
	}

	var resp []ChannelResponse
	for _, ch := range channels {
		config, _ := unmarshalConfig(ch.Config)
		resp = append(resp, *mapChannelToResponse(&ch, config))
	}
	return resp, nil
}


func (s *channelServiceImpl) GetByID(orgID, channelID string) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(orgID, channelID)
	if err != nil {
		return nil, errors.New("notification channel not found")
	}
	config, _ := unmarshalConfig(ch.Config)
	return mapChannelToResponse(ch, config), nil
}


func (s *channelServiceImpl) Update(orgID, channelID string, req *UpdateChannelRequest) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(orgID, channelID)
	if err != nil {
		return nil, errors.New("notification channel not found")
	}

	if req.Name != nil {
		ch.Name = *req.Name
	}
	if req.IsActive != nil {
		ch.IsActive = *req.IsActive
	}

	var mergedConfig map[string]string
	if req.Config != nil {
		if err := validateConfig(ch.Type, *req.Config); err != nil {
			return nil, err
		}
		configJSON, err := marshalConfig(*req.Config)
		if err != nil {
			return nil, errors.New("failed to serialize config")
		}
		ch.Config = configJSON
		mergedConfig = *req.Config
	} else {
		mergedConfig, _ = unmarshalConfig(ch.Config)
	}

	if err := s.repo.Update(ch); err != nil {
		return nil, errors.New("failed to update notification channel")
	}

	return mapChannelToResponse(ch, mergedConfig), nil
}


func (s *channelServiceImpl) Delete(orgID, channelID string) error {
	// Verify the channel exists and belongs to the org before deleting
	if _, err := s.repo.GetByID(orgID, channelID); err != nil {
		return errors.New("notification channel not found")
	}
	if err := s.repo.Delete(orgID, channelID); err != nil {
		return errors.New("failed to delete notification channel")
	}
	return nil
}


func (s *channelServiceImpl) GetLogs(orgID string, limit, offset int) (*LogListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.repo.GetLogs(orgID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to fetch notification logs")
	}

	var logResp []LogResponse
	for _, l := range logs {
		entry := LogResponse{
			ID:          l.ID.String(),
			Channel:     l.Channel,
			Message:     l.Message,
			Status:      l.Status,
			ErrorDetail: l.ErrorDetail,
			SentAt:      l.SentAt,
		}
		if l.MonitorID != nil {
			s := l.MonitorID.String()
			entry.MonitorID = &s
		}
		if l.IncidentID != nil {
			s := l.IncidentID.String()
			entry.IncidentID = &s
		}
		if l.ChannelID != nil {
			s := l.ChannelID.String()
			entry.ChannelID = &s
		}
		logResp = append(logResp, entry)
	}

	return &LogListResponse{
		Logs:   logResp,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// validateConfig checks required keys are present and validates URL values
// for channel types that use external URLs (slack, webhook).
func validateConfig(channelType string, config map[string]string) error {
	keys, ok := requiredConfigKeys[channelType]
	if !ok {
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}

	for _, k := range keys {
		v, exists := config[k]
		if !exists || strings.TrimSpace(v) == "" {
			return fmt.Errorf("config key '%s' is required for channel type '%s'", k, channelType)
		}

		// Validate URL format for url-based keys
		if k == "url" || k == "webhook_url" {
			if err := validateExternalURL(v); err != nil {
				return fmt.Errorf("invalid %s: %w", k, err)
			}
		}

		// Validate email format for email channels
		if k == "email" {
			if !strings.Contains(v, "@") || len(v) < 5 {
				return errors.New("invalid email address in config")
			}
		}
	}
	return nil
}

// validateExternalURL rejects private/internal IP ranges to prevent SSRF.
// This is a defense-in-depth measure for webhook/slack channel types.
func validateExternalURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("must be a valid URL")
	}
	if u.Scheme != "https" {
		return errors.New("only HTTPS URLs are allowed")
	}

	host := strings.ToLower(u.Hostname())

	// Block loopback, private ranges, link-local, and cloud metadata endpoints
	blockedPrefixes := []string{
		"localhost",
		"127.",
		"0.0.0.0",
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.", "172.20.",
		"172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"169.254.",  // link-local (AWS metadata default)
		"[::1]",    // IPv6 loopback
		"[fc", "[fd", // IPv6 private
		"metadata.google.internal",
		"169.254.169.254", // AWS/GCP/Azure metadata
	}

	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(host, prefix) || host == strings.TrimSuffix(prefix, ".") {
			return fmt.Errorf("URL points to a private or reserved address (SSRF protection): %s", host)
		}
	}

	if err := utils.ValidateTargetHost(host); err != nil {
		return fmt.Errorf("SSRF protection: %w", err)
	}

	return nil
}

// marshalConfig serializes a config map to a JSON string for storage.
func marshalConfig(config map[string]string) (string, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// unmarshalConfig deserializes a JSON config string back to a map.
func unmarshalConfig(raw string) (map[string]string, error) {
	var config map[string]string
	if raw == "" {
		return config, nil
	}
	err := json.Unmarshal([]byte(raw), &config)
	return config, err
}

// mapChannelToResponse builds the API response, masking sensitive config values.
func mapChannelToResponse(ch *model.NotificationChannel, config map[string]string) *ChannelResponse {
	maskedConfig := make(map[string]string, len(config))
	for k, v := range config {
		if sensitiveConfigKeys[k] {
			// Show only the domain part of a URL for readability, mask the rest
			if parsed, err := url.Parse(v); err == nil {
				maskedConfig[k] = parsed.Scheme + "://" + parsed.Host + "/***"
			} else {
				maskedConfig[k] = "***"
			}
		} else {
			maskedConfig[k] = v
		}
	}

	return &ChannelResponse{
		ID:             ch.ID.String(),
		OrganizationID: ch.OrganizationID.String(),
		Name:           ch.Name,
		Type:           ch.Type,
		Config:         maskedConfig,
		IsActive:       ch.IsActive,
		CreatedAt:      ch.CreatedAt,
	}
}
