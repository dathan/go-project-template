package stripe

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

// IStripeClient defines the interface for the Stripe client.
type IStripeClient interface {
	CreatePaymentIntent(amount int64, currency string) (*stripe.PaymentIntent, error)
	ConfirmPaymentIntent(id string) (*stripe.PaymentIntent, error)
}

// Client is a client for interacting with the Stripe API.
type Client struct {
	APIKey string
}

// NewClient creates a new Stripe client.
func NewClient(apiKey string) *Client {
	stripe.Key = apiKey
	return &Client{APIKey: apiKey}
}

// CreatePaymentIntent creates a new payment intent.
func (c *Client) CreatePaymentIntent(amount int64, currency string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// ConfirmPaymentIntent confirms a payment intent.
func (c *Client) ConfirmPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Confirm(id, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}
