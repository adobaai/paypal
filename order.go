package paypal

import (
	"context"
	"net/http"
	"time"
)

// OrderIndent is the intent to either capture payment immediately
// or authorize a payment for an order after order creation.
//
// See https://developer.paypal.com/docs/api/orders/v2/#orders_create!path=intent&t=request.
type OrderIntent string

const (
	OICapture   OrderIntent = "CAPTURE"
	OIAuthorize OrderIntent = "AUTHORIZE"
)

type OrderStatus string

const (
	// OSCreated indicates the order was created with the specified context.
	OSCreated OrderStatus = "CREATED"

	// OSSaved indicates the order was saved and persisted.
	OSSaved OrderStatus = "SAVED"

	// OSApproved indicates the customer approved the payment through the PayPal wallet
	// or another form of guest or unbranded payment.
	// For example, a card, bank account, or so on.
	OSApproved OrderStatus = "APPROVED"

	// OSVoided indicates all purchase units in the order are voided.
	OSVoided OrderStatus = "VOIDED"

	// OSCompleted indicates the payment was authorized
	// or the authorized payment was captured for the order.
	OSCompleted OrderStatus = "COMPLETED"

	// OSPayerActionRequired indicates the order requires an action from the payer
	// (e.g. 3DS authentication).
	OSPayerActionRequired OrderStatus = "PAYER_ACTION_REQUIRED"
)

// PurchaseUnit represents either a full or partial order
// that the payer intends to purchase from the payee.
//
// See https://developer.paypal.com/docs/api/orders/v2/#orders_create!path=purchase_units&t=request.
type PurchaseUnit struct {
	Amount *Amount `json:"amount"` // Requried

	// Description is the purchase description.
	//
	// The maximum length of the character is dependent on the type of characters used.
	// The character length is specified assuming a US ASCII character.
	// Depending on type of character; (e.g. accented character, Japanese characters)
	// the number of characters that can be specified as input
	// might not equal the permissible max length.
	Description string `json:"description"`
}

// Amount is the total order amount with an optional breakdown that provides details,
// such as the total item amount, total tax amount, shipping, handling, insurance,
// and discounts, if any.
//
// See https://developer.paypal.com/docs/api/orders/v2/#orders_create!path=purchase_units/amount&t=request.
type Amount struct {
	// CurrencyCode is the three-character ISO-4217 currency code that identifies the currency.
	//
	// Required.
	CurrencyCode string `json:"currency_code"`
	// Required.
	Value string `json:"value"`
}

// PaymentSource is the payment source.
//
// See https://developer.paypal.com/docs/api/orders/v2/#definition-payment_source.
type PaymentSource struct {
}

// Order is the PayPal order.
//
// See https://developer.paypal.com/docs/api/orders/v2/#definition-order.
type Order struct {
	ID            string          `json:"id,omitempty"`
	Intent        OrderIntent     `json:"intent,omitempty"`         // Required
	PurchaseUnits []*PurchaseUnit `json:"purchase_units,omitempty"` // Required
	Status        OrderStatus     `json:"status,omitempty"`
	CreateTime    time.Time       `json:"create_time,omitempty"`
	UpdateTime    time.Time       `json:"update_time,omitempty"`
	Links         []*Link         `json:"links,omitempty"`
}

type CreateOrderReq struct {
	*Order
}

// CreateOrder creates an order.
//
// See https://developer.paypal.com/docs/api/orders/v2/#orders_create.
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderReq) (res *Order, err error) {
	ctx = WithOperation(ctx, "CreateOrder")
	return JSON[Order](ctx, c, http.MethodPost, "/v2/checkout/orders", req)
}

type CaptureOrderReq struct {
	ID            string         `json:"id"`
	PaymentSource *PaymentSource `json:"payment_source"`
}

// CaptureOrder captures payment for an order.
//
// See https://developer.paypal.com/docs/api/orders/v2/#orders_capture.
func (c *Client) CaptureOrder(ctx context.Context, req *CaptureOrderReq) (res *Order, err error) {
	ctx = WithOperation(ctx, "CaptureOrder")
	return JSON[Order](ctx, c, http.MethodPost, "/v2/checkout/orders/"+req.ID+"/capture", req)
}
