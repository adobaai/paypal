package paypal

import (
	"context"
	"net/http"
	"time"
)

// EventType is the type of webhook.
//
// See https://developer.paypal.com/api/rest/webhooks/event-names/
type EventType string

// TODO(ion) Use code gen
const (
	PaymentSaleCompleted EventType = "PAYMENT.SALE.COMPLETED"

	BillingSubscriptionCreated       EventType = "BILLING.SUBSCRIPTION.CREATED"
	BillingSubscriptionActivated     EventType = "BILLING.SUBSCRIPTION.ACTIVATED"
	BillingSubscriptionUpdated       EventType = "BILLING.SUBSCRIPTION.UPDATED"
	BillingSubscriptionExpired       EventType = "BILLING.SUBSCRIPTION.EXPIRED"
	BillingSubscriptionCancelled     EventType = "BILLING.SUBSCRIPTION.CANCELLED"
	BillingSubscriptionSuspended     EventType = "BILLING.SUBSCRIPTION.SUSPENDED"
	BillingSubscriptionPaymentFailed EventType = "BILLING.SUBSCRIPTION.PAYMENT.FAILED"
)

type Webhook struct {
	ID           string         `json:"id,omitempty"`
	CreateTime   time.Time      `json:"create_time,omitempty"`
	ResourceType string         `json:"resource_type,omitempty"`
	EventType    EventType      `json:"event_type,omitempty"`
	Summary      string         `json:"summary,omitempty"`
	Resource     map[string]any `json:"resource,omitempty"`
	Links        []Link         `json:"links,omitempty"`
}

type VerifyWSReq struct {
	AuthAlgo         string    `json:"auth_algo,omitempty"`
	CertURL          string    `json:"cert_url,omitempty"`
	TransmissionID   string    `json:"transmission_id,omitempty"`
	TransmissionTime time.Time `json:"transmission_time,omitempty"`
	TransmissionSig  string    `json:"transmission_sig,omitempty"`

	// WebhookID is the ID of webhook as configured in your Developer Portal account.
	WebhookID    string `json:"webhook_id,omitempty"`
	WebhookEvent any    `json:"webhook_event,omitempty"`
}

type WebhookVerification struct {
	VerificationStatus string `json:"verification_status,omitempty"`
}

// VerifyWebhookSign verifies a webhook signature.
//
// See https://developer.paypal.com/docs/api/webhooks/v1/#verify-webhook-signature_post
func (c *Client) VerifyWebhookSign(ctx context.Context, req *VerifyWSReq,
) (ok bool, err error) {
	ctx = WithOperation(ctx, "VerifyWebhookSign")
	r, err := JSON[WebhookVerification](ctx, c,
		http.MethodPost, "/v1/notifications/verify-webhook-signature", req)
	if err != nil {
		return
	}
	return r.VerificationStatus == "SUCCESS", nil
}
