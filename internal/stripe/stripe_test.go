package stripe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
)

// MockStripeClient is a mock of the Stripe client.
type MockStripeClient struct{}

// CreatePaymentIntent is a mock of the CreatePaymentIntent method.
func (m *MockStripeClient) CreatePaymentIntent(amount int64, currency string) (*stripe.PaymentIntent, error) {
	pi := &stripe.PaymentIntent{}
	pi.ID = "pi_123"
	pi.Amount = 1000
	pi.Currency = "usd"
	return pi, nil
}

// ConfirmPaymentIntent is a mock of the ConfirmPaymentIntent method.
func (m *MockStripeClient) ConfirmPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	return &stripe.PaymentIntent{
		ID:     id,
		Status: stripe.PaymentIntentStatusSucceeded,
	}, nil
}

func TestCreatePaymentIntent(t *testing.T) {
	client := &MockStripeClient{}
	amount := int64(1000)
	currency := "usd"

	pi, err := client.CreatePaymentIntent(amount, currency)

	assert.NoError(t, err)
	assert.NotNil(t, pi)
	assert.Equal(t, "pi_123", pi.ID)
	assert.Equal(t, int64(1000), pi.Amount)
	assert.Equal(t, "usd", pi.Currency)
}

func TestConfirmPaymentIntent(t *testing.T) {
	client := &MockStripeClient{}
	id := "pi_123"

	pi, err := client.ConfirmPaymentIntent(id)

	assert.NoError(t, err)
	assert.NotNil(t, pi)
	assert.Equal(t, id, pi.ID)
	assert.Equal(t, stripe.PaymentIntentStatusSucceeded, pi.Status)
}
