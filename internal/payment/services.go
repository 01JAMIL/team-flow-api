package payment

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/env"

	"github.com/stripe/stripe-go/v86"
)

type Svc struct {
	repo   *repo.Queries
	client *stripe.Client
}

func NewStripeService(client *stripe.Client, repo *repo.Queries) *Svc {
	return &Svc{
		client: client,
		repo:   repo,
	}
}

func (s *Svc) CreateStripeCustomer(name string) (*stripe.Customer, error) {
	params := &stripe.CustomerCreateParams{
		Name: stripe.String(name),
	}

	return s.client.V1Customers.Create(
		context.Background(),
		params,
	)
}

func (s *Svc) CreateCheckoutSession(workspaceID string, customerID string, priceID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionCreateParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),

		Metadata: map[string]string{
			"workspace_id": workspaceID,
		},

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
