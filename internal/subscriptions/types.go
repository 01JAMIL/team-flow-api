package subscriptions

import "github.com/jackc/pgx/v5/pgtype"

type CreateSubscriptionPayload struct {
	WorkspaceID          pgtype.UUID        `json:"workspaceId" binding:"required"`
	StripeSubscriptionID string             `json:"stripeSubscriptionId" binding:"required"`
	StripePriceID        string             `json:"stripePriceId" binding:"required"`
	Status               string             `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
	Plan                 string             `json:"plan" binding:"required,oneof=PRO BUSINESS ENTERPRISE"`
	CurrentPeriodStart   pgtype.Timestamptz `json:"currentPeriodStart"`
	CurrentPeriodEnd     pgtype.Timestamptz `json:"currentPeriodEnd"`
}

type UpdateSubscriptionPayload struct {
	StripePriceID        string             `json:"stripePriceId" binding:"required"`
	Status               string             `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
	CurrentPeriodStart   pgtype.Timestamptz `json:"currentPeriodStart"`
	CurrentPeriodEnd     pgtype.Timestamptz `json:"currentPeriodEnd"`
	StripeSubscriptionID string             `json:"stripeSubscriptionId" binding:"required"`
}

type subscriptionResponse struct {
	ID                   string             `json:"id"`
	WorkspaceID          string             `json:"workspaceId"`
	StripeSubscriptionID string             `json:"stripeSubscriptionId"`
	StripePriceID        string             `json:"stripePriceId"`
	Status               string             `json:"status"`
	Plan                 string             `json:"plan"`
	CurrentPeriodStart   pgtype.Timestamptz `json:"currentPeriodStart"`
	CurrentPeriodEnd     pgtype.Timestamptz `json:"currentPeriodEnd"`
	CreatedAt            pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt            pgtype.Timestamptz `json:"updatedAt"`
}
