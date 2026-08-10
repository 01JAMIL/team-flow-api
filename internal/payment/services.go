package payment

import (
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/customer"
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
	params := &stripe.CustomerParams{
		Name: stripe.String(name),
	}

	return customer.New(params)
}
