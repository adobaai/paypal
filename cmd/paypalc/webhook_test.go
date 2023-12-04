package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adobaai/paypal/ptesting"
)

func Test_parseTitle(t *testing.T) {
	s := `
<h2 style="position: relative">
  <a href="#payments" class="anchor before">
    <svg viewBox="0 0 16 16" width="16"></svg>
  </a>
  <div class="hidden-anchor" id="payments"></div>
  Billing plans and agreements
</h2>
`
	ptesting.R(parseTitle([]byte(s))).NoError(t).Equal([]byte("Billing plans and agreements"))
}

func Test_parseDescription(t *testing.T) {
	s := `
</h2>
  <p>
    The webhooks for orders correspond to both supported versions of the Orders
    API:
</p>
`

	expected := []byte("The webhooks for orders correspond to both supported versions of the Orders API")
	ptesting.R(parseDescription([]byte(s))).NoError(t).Equal(expected)
}

func Test_parseVersion(t *testing.T) {
	s := `
<h3 style="position: relative">
  <a href="#v2-1" aria-label="v2 1 permalink" class="anchor before">
    <svg aria-hidden="true" width="16">
      <path fill-rule="evenodd"></path>
    </svg>
  </a>
  <div class="hidden-anchor" id="v2-1"></div>
  V2
</h3>
`
	ptesting.R(getFirstMatch(reVersion, []byte(s))).NoError(t).Equal([]byte("V2"))
}

func Test_parseWebhook(t *testing.T) {
	s := `
<tr>
  <td>
    <code class="language-text">PAYMENT.AUTHORIZATION.VOIDED</code>
  </td>
  <td>
    A payment authorization is voided either due to authorization reaching it’s
    30 day validity period or authorization was manually voided using the Void
    Authorized Payment API.
  </td>
  <td>
    <a href="/docs/api/payments/v2/#authorizations_get">
      Show details for authorized payment
    </a>
    with response
    <code class="language-text">status</code> of
    <code class="language-text">voided</code>.
  </td>
</tr>
`
	ptesting.R(parseWebhook([]byte(s))).NoError(t).Equal(Webhook{
		ID:    "PaymentAuthorizationVoided",
		Event: "PAYMENT.AUTHORIZATION.VOIDED",
		Trigger: Comment{
			Content: []byte(`A payment authorization is voided either due to authorization reaching it’s 30 day validity period or authorization was manually voided using the Void Authorized Payment API.`),
		},
		RelatedMethod: Comment{
			Content: []byte("[Show details for authorized payment] with response `status` of `voided`."),
			Links: []Link{
				{
					"Show details for authorized payment",
					"https://developer.paypal.com/docs/api/payments/v2/#authorizations_get",
				},
			},
		},
	})
}

func Test_removeWhitespaces(t *testing.T) {
	s := `
    The webhooks for orders correspond to both supported versions
    of the Orders API:
  `
	expected := []byte("The webhooks for orders correspond to both supported versions of the Orders API:")
	assert.Equal(t, expected, NewParser([]byte(s)).RemoveWhitespaces().Bytes())
}

func Test_wrap(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		s := []byte("The webhooks for authorizing and capturing payments correspond...")
		expected := [][]byte{
			[]byte("The webhooks for authorizing and"),
			[]byte("capturing payments correspond..."),
		}
		assert.Equal(t, expected, wrap(s, 40))
	})

	t.Run("Corner", func(t *testing.T) {
		s := []byte("The webhooks for authorizing and capturing")
		expected := [][]byte{
			[]byte("The webhooks for authorizing and"),
			[]byte("capturing"),
		}
		assert.Equal(t, expected, wrap(s, 40))
	})
}
