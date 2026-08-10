package payment

import (
	"context"
	"gin-api-1/internal/env"

	"github.com/stripe/stripe-go/v86"
)

type Service struct {
	client *stripe.Client
}

func NewStripeService(client *stripe.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) CreateStripeCustomer(name string) (*stripe.Customer, error) {
	params := &stripe.CustomerCreateParams{
		Name: stripe.String(name),
	}

	return s.client.V1Customers.Create(
		context.Background(),
		params,
	)
}

func (s *Service) CreateCheckoutSession(customerID string, priceID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionCreateParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),

		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: new(int64(1)),
			},
		},

		SuccessURL: stripe.String(env.GetEnvString("STRIPE_SUCCESS_URL", "http://localhost:3000/payment/success")),
		CancelURL:  stripe.String(env.GetEnvString("STRIPE_CANCEL_URL", "http://localhost:3000/payment/cancel")),
	}

	return s.client.V1CheckoutSessions.Create(
		context.Background(),
		params,
	)
}
