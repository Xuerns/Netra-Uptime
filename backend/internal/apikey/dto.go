package apikey

type CreateAPIKeyRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Permissions string `json:"permissions" binding:"required,oneof=read write"` // read or write
	ExpiresIn   *int   `json:"expires_in_days"`                         // optional: days until expiry (e.g., 30, 90, 365, or null for never)
}

type APIKeyResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	KeyPrefix   string  `json:"key_prefix"`
	Permissions string  `json:"permissions"`
	LastUsedAt  *string `json:"last_used_at"`
	LastUsedIP  string  `json:"last_used_ip"`
	ExpiresAt   *string `json:"expires_at"`
	CreatedAt   string  `json:"created_at"`
	CreatedBy   string  `json:"created_by"`
}

type CreatedAPIKeyResponse struct {
	APIKeyResponse
	PlaintextKey string `json:"plaintext_key"` // Shown ONLY once upon creation!
}
