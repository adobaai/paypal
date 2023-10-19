package paypal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobaai/paypal/ptesting"
)

func TestSubscription(t *testing.T) {
	c := NewTestClient()
	ctx := context.Background()
	subID := "I-SF5PBRMX4EPF"
	ptesting.R(c.GetSubscription(ctx, &GetSubscriptionReq{ID: subID})).NoError(t).
		Do(func(t *testing.T, it *Subscription) {
			t.Log(it)
		})

	var e *Error
	ptesting.R(c.GetSubscription(ctx, &GetSubscriptionReq{ID: "I-SF6PBRMK4EPJ"})).ErrorAs(t, &e)
	assert.NotZero(t, e.DebugID)
	e.DebugID = ""
	href := "https://developer.paypal.com/docs/api/v1/billing/subscriptions#RESOURCE_NOT_FOUND"
	assert.Equal(t, &Error{
		StatusCode: 404,
		Name:       "RESOURCE_NOT_FOUND",
		Message:    "The specified resource does not exist.",
		Details: []*ErrorDetail{
			{
				Issue:       "INVALID_RESOURCE_ID",
				Description: "Requested resource ID was not found.",
			},
		},
		Links: []*Link{
			{
				HRef:   href,
				Rel:    "information_link",
				Method: "GET",
			},
		},
	}, e)
}
