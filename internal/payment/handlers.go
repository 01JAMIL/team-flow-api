package payment

import (
	"encoding/json"
	"fmt"
	"gin-api-1/internal/email"
	"gin-api-1/internal/env"
	"gin-api-1/internal/subscriptions"
	"io"
	"log"
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
	emailService         email.Service
}

func NewPaymentHandler(service Svc, subscriptionsService subscriptions.Service, emailService email.Service) *handler {
	return &handler{
		service:              service,
		subscriptionsService: subscriptionsService,
		emailService:         emailService,
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

	switch event.Type {
	case "checkout.session.completed":
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

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

	case "customer.subscription.updated":
		var subscription stripe.Subscription

		if err := json.Unmarshal(
			event.Data.Raw,
			&subscription,
		); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		if len(subscription.Items.Data) == 0 {
			c.Status(http.StatusBadRequest)
			return
		}

		priceID := subscription.Items.Data[0].Price.ID
		status := mapStripeSubscriptionStatus(subscription.Status)

		periodStart := time.Now().UTC()
		periodEnd := periodStart.AddDate(0, 1, 0)

		_, err := h.subscriptionsService.UpdateSubscription(c, subscriptions.UpdateSubscriptionPayload{
			StripePriceID: priceID,
			Status:        status,
			CurrentPeriodStart: pgtype.Timestamptz{
				Time:  periodStart,
				Valid: true,
			},
			CurrentPeriodEnd: pgtype.Timestamptz{
				Time:  periodEnd,
				Valid: true,
			},
			StripeSubscriptionID: subscription.ID,
		})

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

	case "customer.subscription.deleted":
		var subscription stripe.Subscription

		if err := json.Unmarshal(
			event.Data.Raw,
			&subscription,
		); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		_, err := h.subscriptionsService.DeactivateSubscription(
			c,
			subscription.ID,
		)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

	case "invoice.payment_failed":
		var invoice stripe.Invoice

		if err := json.Unmarshal(
			event.Data.Raw,
			&invoice,
		); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		fmt.Println("invoice.Customer.ID : ", invoice.Customer.ID)

		user, err := h.service.repo.GetUserByStripeCustomerID(c, pgtype.Text{String: invoice.Customer.ID, Valid: true})
		if err != nil {
			log.Printf("failed to get user: %v", err)
		}

		fmt.Println("User email : ", user.Email)

		if user.Email != "" {
			if err := h.emailService.PaymentFailedEmail(user.Email); err != nil {
				log.Printf("failed to send payment failed email: %v", err)
			}
		}

		c.Status(http.StatusOK)

	default:
		fmt.Println("Unhandled event:", event.Type)
		c.Status(http.StatusBadRequest)
	}

	c.Status(http.StatusOK)
}

func mapStripeSubscriptionStatus(
	status stripe.SubscriptionStatus,
) string {
	switch status {
	case stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing:
		return "ACTIVE"

	case stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusUnpaid,
		stripe.SubscriptionStatusIncompleteExpired:
		return "INACTIVE"

	default:
		return "INACTIVE"
	}
}
