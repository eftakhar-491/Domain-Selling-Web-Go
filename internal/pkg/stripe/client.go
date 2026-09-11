package stripe

import (
	"fmt"

	"project-setup/internal/config"

	stripe "github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/webhook"
)

// StripeService wraps all Stripe API interactions
type StripeService struct {
	secretKey     string
	webhookSecret string
}

// NewStripeService creates a StripeService from the application config
func NewStripeService() *StripeService {
	env := config.GetEnv()
	stripe.Key = env.StripeSecretKey

	return &StripeService{
		secretKey:     env.StripeSecretKey,
		webhookSecret: env.StripeWebhookSecret,
	}
}

// CreatePaymentIntent creates a Stripe PaymentIntent for the given amount
// amount is in the major currency unit (e.g. 9.99 USD), it will be converted to cents
func (s *StripeService) CreatePaymentIntent(amount float64, currency string, metadata map[string]string) (*stripe.PaymentIntent, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("payment amount must be greater than zero")
	}

	// Stripe expects amounts in the smallest currency unit (cents for USD)
	amountInCents := int64(amount * 100)

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amountInCents),
		Currency:           stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	// Attach metadata (order number, user id, etc.)
	if metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range metadata {
			params.Metadata[k] = v
		}
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return pi, nil
}

// GetPaymentIntent retrieves a PaymentIntent by its ID
func (s *StripeService) GetPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}
	return pi, nil
}

// ConstructWebhookEvent validates and parses a Stripe webhook event
func (s *StripeService) ConstructWebhookEvent(payload []byte, sigHeader string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
}
