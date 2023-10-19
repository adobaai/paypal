package paypal

import (
	"context"
	"net/http"
)

type PaymentPreferences struct {
	SetupFee *Amount `json:"setup_fee,omitempty"`
}

type SubscriptionPlan struct {
	PaymentPreferences *PaymentPreferences `json:"payment_preferences,omitempty"`
}

type SubscriptionStatus string

const (
	SSApprovalPending SubscriptionStatus = "APPROVAL_PENDING"
	SSApproved        SubscriptionStatus = "APPROVED"
	SSActive          SubscriptionStatus = "ACTIVE"
	SSSuspended       SubscriptionStatus = "SUSPENDED"
	SSCancelled       SubscriptionStatus = "CANCELLED"
	SSExpired         SubscriptionStatus = "EXPIRED"
)

type Subscription struct {
	ID       string             `json:"id,omitempty"`
	PlanID   string             `json:"plan_id,omitempty"`
	Quantity string             `json:"quantity,omitempty"`
	Status   SubscriptionStatus `json:"status,omitempty"`
	Plan     *SubscriptionPlan  `json:"plan,omitempty"`
}

type CreateSubscriptionReq struct {
	*Subscription
}

// CreateSubscription creates a subscription.
//
// See https://developer.paypal.com/docs/api/subscriptions/v1/#subscriptions_create
func (c *Client) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq,
) (res *Subscription, err error) {
	return JSON[Subscription](ctx, c, http.MethodPost, "/v1/billing/subscriptions", req)
}

type GetSubscriptionReq struct {
	ID string
}

// GetSubscription get details of a subscription.
//
// See https://developer.paypal.com/docs/api/subscriptions/v1/#subscriptions_get
func (c *Client) GetSubscription(ctx context.Context, req *GetSubscriptionReq,
) (res *Subscription, err error) {
	ctx = WithOperation(ctx, "GetSubscription")
	return JSON[Subscription](ctx, c, http.MethodGet, "/v1/billing/subscriptions/"+req.ID, nil)
}

type CancelSubscriptionReq struct {
	ID     string `json:"-"`                // The ID of the subscription
	Reason string `json:"reason,omitempty"` // The reason for the cancellation
}

// CancelSubscription cancels a subscription.
//
// See https://developer.paypal.com/docs/api/subscriptions/v1/#subscriptions_cancel
func (c *Client) CancelSubscription(ctx context.Context, req *CancelSubscriptionReq) (err error) {
	ctx = WithOperation(ctx, "CancelSubscription")
	path := "/v1/billing/subscriptions/" + req.ID + "/cancel"
	err = JSONNop(ctx, c, http.MethodPost, path, req)
	return
}
