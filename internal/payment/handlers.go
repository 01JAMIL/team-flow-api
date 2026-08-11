package payment

import (
	"encoding/json"
	"gin-api-1/internal/env"
	"gin-api-1/internal/subscriptions"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
)

type handler struct {
	service              Svc
	subscriptionsService subscriptions.Service
}

func NewPaymentHandler(service Svc, subscriptionsService subscriptions.Service) *handler {
	return &handler{
		service:              service,
		subscriptionsService: subscriptionsService,
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
		var checkoutSession stripe.CheckoutSession

		if err := json.Unmarshal(
			event.Data.Raw,
			&checkoutSession,
		); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		workspaceID := checkoutSession.Metadata["workspace_id"]
		workspaceUUID, err := uuid.Parse(workspaceID)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		subscriptionID := checkoutSession.Subscription.ID

		subscription, err := h.subscriptionsService.GetSubscription(
			subscriptionID,
		)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		priceID := subscription.Items.Data[0].Price.ID

		periodStart := time.Now().UTC()
		periodEnd := periodStart.AddDate(0, 1, 0)
		_, err = h.subscriptionsService.CreateSubscription(c, subscriptions.CreateSubscriptionPayload{
			WorkspaceID:          pgtype.UUID{Bytes: workspaceUUID, Valid: true},
			StripeSubscriptionID: subscriptionID,
			StripePriceID:        priceID,
			Status:               "ACTIVE",
			Plan:                 "PRO",
			CurrentPeriodStart: pgtype.Timestamptz{
				Time:  periodStart,
				Valid: true,
			},
			CurrentPeriodEnd: pgtype.Timestamptz{
				Time:  periodEnd,
				Valid: true,
			},
		})
	}

	c.Status(http.StatusOK)
}
