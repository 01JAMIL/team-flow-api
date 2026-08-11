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
