package apikey

import (
	"PulseWatch/internal/audit"
	"PulseWatch/internal/model"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type APIKeyService interface {
	CreateKey(orgID, userID string, req CreateAPIKeyRequest, ip string) (*CreatedAPIKeyResponse, error)
	ListKeys(orgID string) ([]APIKeyResponse, error)
	RevokeKey(orgID, userID, keyID string, ip string) error
	ValidateKey(plaintextKey string) (*model.APIKey, error)
	TouchLastUsed(keyID string, ip string) error
}

type APIKeyServiceImpl struct {
	repo         APIKeyRepository
	auditService audit.AuditService
}

func NewAPIKeyService(repo APIKeyRepository, auditService audit.AuditService) APIKeyService {
	return &APIKeyServiceImpl{
		repo:         repo,
		auditService: auditService,
	}
}

func HashAPIKey(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func generateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *APIKeyServiceImpl) CreateKey(orgID, userID string, req CreateAPIKeyRequest, ip string) (*CreatedAPIKeyResponse, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, errors.New("invalid organization ID")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	randomHex, err := generateRandomHex(32)
	if err != nil {
		return nil, errors.New("failed to generate secure key")
	}

	plaintextKey := fmt.Sprintf("pw_live_%s", randomHex)
	keyHash := HashAPIKey(plaintextKey)
	keyPrefix := fmt.Sprintf("pw_live_%s", randomHex[:8])

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &exp
	}

	apiKey := model.APIKey{
		OrganizationID: orgUUID,
		CreatedByID:    userUUID,
		Name:           req.Name,
		KeyHash:        keyHash,
		KeyPrefix:      keyPrefix,
		Permissions:    req.Permissions,
		ExpiresAt:      expiresAt,
	}

	if err := s.repo.Create(&apiKey); err != nil {
		return nil, errors.New("failed to create API key")
	}

	if s.auditService != nil {
		keyIDStr := apiKey.ID.String()
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "API_KEY_CREATE",
			EntityType:     "api_key",
			EntityID:       &keyIDStr,
			Description:    fmt.Sprintf("Created API Key '%s' (%s scope)", apiKey.Name, apiKey.Permissions),
			IPAddress:      ip,
		})
	}

	resp := mapToResponse(apiKey)
	return &CreatedAPIKeyResponse{
		APIKeyResponse: resp,
		PlaintextKey:   plaintextKey,
	}, nil
}

func (s *APIKeyServiceImpl) ListKeys(orgID string) ([]APIKeyResponse, error) {
	keys, err := s.repo.GetAllByOrgID(orgID)
	if err != nil {
		return nil, errors.New("failed to fetch API keys")
	}

	var res []APIKeyResponse
	for _, k := range keys {
		res = append(res, mapToResponse(k))
	}
	return res, nil
}

func (s *APIKeyServiceImpl) RevokeKey(orgID, userID, keyID string, ip string) error {
	if err := s.repo.Delete(orgID, keyID); err != nil {
		return errors.New("failed to revoke API key")
	}

	if s.auditService != nil {
		_ = s.auditService.Record(audit.RecordAuditLogParams{
			OrganizationID: orgID,
			UserID:         userID,
			Action:         "API_KEY_REVOKE",
			EntityType:     "api_key",
			EntityID:       &keyID,
			Description:    fmt.Sprintf("Revoked API Key (ID: %s)", keyID),
			IPAddress:      ip,
		})
	}
	return nil
}

func (s *APIKeyServiceImpl) ValidateKey(plaintextKey string) (*model.APIKey, error) {
	keyHash := HashAPIKey(plaintextKey)
	key, err := s.repo.GetByHash(keyHash)
	if err != nil {
		return nil, errors.New("invalid API key")
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, errors.New("API key has expired")
	}

	return key, nil
}

func (s *APIKeyServiceImpl) TouchLastUsed(keyID string, ip string) error {
	return s.repo.UpdateLastUsed(keyID, ip)
}

func mapToResponse(k model.APIKey) APIKeyResponse {
	resp := APIKeyResponse{
		ID:          k.ID.String(),
		Name:        k.Name,
		KeyPrefix:   k.KeyPrefix,
		Permissions: k.Permissions,
		LastUsedIP:  k.LastUsedIP,
		CreatedAt:   k.CreatedAt.Format(time.RFC3339),
	}

	if k.CreatedBy.Name != "" {
		resp.CreatedBy = k.CreatedBy.Name
	} else if k.CreatedByID != uuid.Nil {
		resp.CreatedBy = k.CreatedByID.String()
	}

	if k.LastUsedAt != nil {
		tStr := k.LastUsedAt.Format(time.RFC3339)
		resp.LastUsedAt = &tStr
	}
	if k.ExpiresAt != nil {
		tStr := k.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &tStr
	}
	return resp
}
