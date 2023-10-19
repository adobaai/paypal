package paypal

type Sale struct {
	ID                 string `json:"id,omitempty"`
	BillingAgreementId string `json:"billing_agreement_id,omitempty"` // Subscription ID
	Amount             struct {
		Total    string `json:"total,omitempty"`
		Currency string `json:"currency,omitempty"`
	} `json:"amount,omitempty"`
	Links []*Link `json:"links,omitempty"`
}
