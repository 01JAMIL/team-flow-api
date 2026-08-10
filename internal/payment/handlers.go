package payment

import (
	"gin-api-1/internal/env"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86/webhook"
)

type handler struct {
	service Svc
}

func NewPaymentHandler(service Svc) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	signature := c.GetHeader("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(
		payload,
		signature,
		env.GetEnvString("STRIPE_WEBHOOK_SECRET", ""),
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if event.Type == "checkout.session.completed" {
		// Handle database update logic
	}

	c.Status(http.StatusOK)
}
