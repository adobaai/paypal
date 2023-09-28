package paypal

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobaai/paypal/ptesting"
)

func NewTestClient() *Client {
	return NewClient("", "", "")
}

func TestAuth(t *testing.T) {
	ctx := context.Background()
	c := NewTestClient()
	token := ptesting.R(c.Auth(ctx)).NoError(t).V()
	assert.NotZero(t, token.Scope)
	assert.NotZero(t, token.AccessToken)
	assert.Equal(t, token.TokenType, "Bearer")
	assert.NotZero(t, token.AppID)
	assert.NotZero(t, token.ExpiresIn)
	assert.NotZero(t, token.Nonce)
}

func TestError(t *testing.T) {
	ctx := context.Background()
	t.Run("Auth", func(t *testing.T) {
		c := NewClient("https://api-m.sandbox.paypal.com", "1234", "5678")
		ptesting.R(c.Auth(ctx)).EqualError(t, "invalid_client: Client Authentication failed")
	})

	t.Run("CreateOrder", func(t *testing.T) {
		c := NewTestClient()
		var e *Error
		order := &Order{
			Intent: OICapture,
		}
		ptesting.R(c.CreateOrder(ctx, &CreateOrderReq{Order: order})).ErrorAs(t, &e)
		assert.NotZero(t, e.DebugID)
		e.DebugID = ""
		assert.Equal(t, &Error{
			StatusCode: http.StatusBadRequest,
			Name:       "INVALID_REQUEST",
			Message:    "Request is not well-formed, syntactically incorrect, or violates schema.",
			Details: []*ErrorDetail{
				{
					Field:       "/purchase_units",
					Location:    "body",
					Issue:       "MISSING_REQUIRED_PARAMETER",
					Description: "A required field / parameter is missing.",
				},
			},
			Links: []*Link{
				{
					HRef:   "https://developer.paypal.com/docs/api/orders/v2/#error-MISSING_REQUIRED_PARAMETER",
					Rel:    "information_link",
					Method: "",
				},
			},
		}, e)
	})
}

type Hello struct {
	Name string
}

type HelloResp struct {
	JSON *Hello
}

func TestJSON(t *testing.T) {
	c := &http.Client{}
	ctx := context.Background()
	name := "Фёдор Миха́йлович Достое́вский"
	url := "https://httpbin.org/anything"
	hreq := ptesting.R(NewJSONRequest(ctx, http.MethodPost, url, Hello{Name: name})).NoError(t).V()
	hres := ptesting.R(c.Do(hreq)).NoError(t).V()
	ptesting.R(RespJSON[HelloResp](hres)).NoError(t).Do(func(t *testing.T, it *HelloResp) {
		assert.Equal(t, name, it.JSON.Name)
	})
}
