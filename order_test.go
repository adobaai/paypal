package paypal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobaai/paypal/ptesting"
)

func TestOrder(t *testing.T) {
	c := NewTestClient()
	ctx := context.Background()
	t.Run("Create", func(t *testing.T) {
		order := &Order{
			Intent: OICapture,
			PurchaseUnits: []*PurchaseUnit{
				{
					Amount: &Amount{
						CurrencyCode: "USD",
						Value:        "12.12",
					},
				},
			},
		}
		ptesting.R(c.CreateOrder(ctx, &CreateOrderReq{Order: order})).
			NoError(t).
			Do(func(t *testing.T, it *Order) {
				assert.NotZero(t, it.ID)
				assert.Equal(t, it.Status, OSCreated)
			})
	})
}
