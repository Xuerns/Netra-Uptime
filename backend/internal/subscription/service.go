package subscription

import (
	"PulseWatch/internal/audit"
	"PulseWatch/internal/model"
	"PulseWatch/pkg/payment"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/redis/go-redis/v9"
)

type SubscriptionService interface {
	CreateSubscriptionTransaction(OrganizationId, plan, Email string) (*snap.Response, error)
	ProcessSubscriptionWebhook(notification *MidtransNotification) error
	GetAllSubsByUserId(userId string) ([]SubscriptionResponse, error)
}

type SubscriptionServiceImpl struct {
	repo         SubscriptionRepository
	rdb          *redis.Client
	auditService audit.AuditService
}

func NewSubscriptionService(repo SubscriptionRepository, rdb *redis.Client, auditService audit.AuditService) SubscriptionService {
	return &SubscriptionServiceImpl{repo: repo, rdb: rdb, auditService: auditService}
}

func (s *SubscriptionServiceImpl) CreateSubscriptionTransaction(OrganizationId, plan, Email string) (*snap.Response, error) {
	var price int64

	switch plan {
	case model.PlanPro:
		price = model.PricePro
	case model.PlanMax:
		price = model.PriceMax
	default:
		return nil, errors.New("Invalid plan type")
	}

	existingTx, _ := s.repo.FindPendingTransaction(OrganizationId, plan)

	if existingTx != nil {
		return &snap.Response{
			Token:       existingTx.SnapToken,
			RedirectURL: existingTx.SnapRedirectURL,
		}, nil
	}

	// Create unique orderId
	orderId := fmt.Sprintf("org_%s_%d", OrganizationId, time.Now().Unix())

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderId,
			GrossAmt: price,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    plan,
				Price: price,
				Qty:   1,
				Name:  fmt.Sprintf("PulseWatch %s Plan - 1 Month", plan),
			},
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: Email,
		},
	}

	snapResp, err := payment.SnapClient.CreateTransaction(req)
	if err != nil {
		return nil, err
	}

	orgUUID, _ := uuid.Parse(OrganizationId)

	transaction := &model.Transaction{
		OrderID:         orderId,
		OrganizationID:  orgUUID,
		PlanType:        plan,
		Amount:          price,
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
	}
	if err := s.repo.CreateTransaction(transaction); err != nil {
		return nil, errors.New("failed to create transaction")
	}

	return snapResp, nil
}

func (s *SubscriptionServiceImpl) ProcessSubscriptionWebhook(notification *MidtransNotification) error {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	Sign := notification.OrderID + notification.StatusCode + notification.GrossAmount + serverKey

	hash := sha512.New()
	hash.Write([]byte(Sign))
	expectedSign := hex.EncodeToString(hash.Sum(nil))

	if notification.SignatureKey != expectedSign {
		return errors.New("Invalid signature")
	}

	parts := strings.Split(notification.OrderID, "_")
	if len(parts) != 3 {
		return errors.New("Invalid Subscription")
	}

	transaction, err := s.repo.FincTransactionByOrderId(notification.OrderID)

	if err != nil {
		return errors.New("Transaction not found")
	}

	if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {
		transaction.Status = "success"
		transaction.PaymentMethod = notification.PaymentType

		if err := s.repo.UpdateTransaction(transaction); err != nil {
			return errors.New("Something wrong, please try again later")
		}

		orgIdStr := transaction.OrganizationID.String()

		subs, err := s.repo.FindSubscriptionByOrgId(orgIdStr)
		if err != nil {
			return errors.New("Your plan not found")
		}

		subs.Plan = transaction.PlanType
		subs.Status = "active"

		now := time.Now()
		if subs.CurrentPeriodEnd != nil && subs.CurrentPeriodEnd.After(now) {
			newEnd := subs.CurrentPeriodEnd.AddDate(0, 1, 0)
			subs.CurrentPeriodEnd = &newEnd
		} else {
			newEnd := now.AddDate(0, 1, 0)
			subs.CurrentPeriodEnd = &newEnd
		}

		if err := s.repo.UpdateSubscription(subs); err != nil {
			return errors.New("failed to update subscription")
		}

		if s.auditService != nil {
			// Find owner or user from org to record log user_id if needed, or system/org owner
			_ = s.auditService.Record(audit.RecordAuditLogParams{
				OrganizationID: orgIdStr,
				UserID:         transaction.OrganizationID.String(), // system fallback ID or org UUID
				Action:         "SUBSCRIPTION_UPGRADE",
				EntityType:     "subscription",
				Description:    fmt.Sprintf("Plan upgraded to %s", transaction.PlanType),
			})
		}

	} else if notification.TransactionStatus == "deny" || notification.TransactionStatus == "expire" || notification.TransactionStatus == "failure" {
		transaction.Status = "failed"
		if err := s.repo.UpdateTransaction(transaction); err != nil {
			return errors.New("Something wrong, please try again later")
		}
	}

	return nil

}

func mapSubscriptionToResponse(sub model.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		OrganizationID: sub.OrganizationID.String(),
		Plan:           sub.Plan,
		Status:         sub.Status,
		Organization: SubscriptionOrgInfo{
			Name: sub.Organization.Name,
		},
	}
	if sub.CurrentPeriodEnd != nil {
		resp.CurrentPeriodEnd = sub.CurrentPeriodEnd.Format(time.RFC3339)
	}
	return resp
}

func (s *SubscriptionServiceImpl) GetAllSubsByUserId(userId string) ([]SubscriptionResponse, error) {
	ctx := context.Background()
	cacheKey := "subs:" + userId

	cacheData, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cacheData != "" {
		var subsResponse []SubscriptionResponse

		if err := json.Unmarshal([]byte(cacheData), &subsResponse); err == nil {
			return subsResponse, nil
		}
	}

	subs, err := s.repo.GetAllSubsByUserId(userId)

	if err != nil {
		return nil, errors.New("Failed to get user subscriptions")
	}

	var response []SubscriptionResponse
	for _, sub := range subs {
		response = append(response, mapSubscriptionToResponse(sub))
	}

	if jsonData, err := json.Marshal(response); err == nil {
		s.rdb.Set(ctx, cacheKey, jsonData, 60*time.Minute)
	}

	return response, nil
}
