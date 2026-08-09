package email

import "github.com/resend/resend-go/v3"

type Service struct {
	client *resend.Client
}

func NewEmailService(client *resend.Client) *Service {
	return &Service{
		client,
	}
}

func (s *Service) SendRegisterWelcomeEmail(to string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    "onboarding@giliwawa.tn",
		To:      []string{to},
		Subject: "Welcome to my application",
		Html:    "<h1>Hello!</h1><br/><p>Welcome to our platform! Your account has been successfully created, and we're excited to have you with us.</p>",
	})
	return err
}
