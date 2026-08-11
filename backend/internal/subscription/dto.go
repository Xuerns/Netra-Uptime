package subscription

type MidtransNotification struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
}

type CreateSubscriptionTransactionRequest struct {
	Plan  string `json:"plan"`
	Email string `json:"email"`
}

type SubscriptionResponse struct {
	OrganizationID   string              `json:"organization_id"`
	Organization     SubscriptionOrgInfo `json:"organization"`
	Plan             string              `json:"plan"`
	Status           string              `json:"status"`
	CurrentPeriodEnd string              `json:"current_period_end"`
}

type SubscriptionOrgInfo struct {
	Name string `json:"name"`
}
