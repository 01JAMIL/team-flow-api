package subscriptions

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v86"
)

type Service interface {
	CreateSubscription(ctx context.Context, payload CreateSubscriptionPayload) (repo.Subscription, error)
	UpdateSubscription(ctx context.Context, payload UpdateSubscriptionPayload) (repo.Subscription, error)
	GetSubscription(subscriptionID string) (*stripe.Subscription, error)
	DeactivateSubscription(ctx context.Context, stripeSubscriptionID string) (repo.Subscription, error)
	GetWorkspaceSubscription(ctx context.Context, workspaceID, userID string) (subscriptionResponse, error)
}

type svc struct {
	repo   *repo.Queries
	db     *pgx.Conn
	client *stripe.Client
}

func NewSubscriptionsService(repo *repo.Queries, db *pgx.Conn, client *stripe.Client) Service {
	return &svc{
		repo:   repo,
		db:     db,
		client: client,
	}
}

func (s *svc) CreateSubscription(ctx context.Context, payload CreateSubscriptionPayload) (repo.Subscription, error) {

	pk := uuid.New()

	workspaceUUID, err := uuid.Parse(payload.WorkspaceID.String())
	if err != nil {
		return repo.Subscription{}, codeerror.New(codeerror.InvalidUUID, "Workspace ID is not a valid UUID")
	}

	return s.repo.CreateSubscription(ctx, repo.CreateSubscriptionParams{
		ID:                   pgtype.UUID{Bytes: pk, Valid: true},
		WorkspaceID:          pgtype.UUID{Bytes: workspaceUUID, Valid: true},
		StripeSubscriptionID: payload.StripeSubscriptionID,
		StripePriceID:        payload.StripePriceID,
		Status:               payload.Status,
		Plan:                 payload.Plan,
		CurrentPeriodStart:   pgtype.Timestamptz{Time: payload.CurrentPeriodStart.Time, Valid: true},
		CurrentPeriodEnd:     pgtype.Timestamptz{Time: payload.CurrentPeriodEnd.Time, Valid: true},
	})
}

func (s *svc) UpdateSubscription(ctx context.Context, payload UpdateSubscriptionPayload) (repo.Subscription, error) {
	_, err := s.repo.GetSubscriptionByStripeSubscription(ctx, payload.StripeSubscriptionID)
	if err != nil {
		return repo.Subscription{}, codeerror.New(codeerror.SubscriptionNotFound, "Subscription not found")
	}

	return s.repo.UpdateSubscription(ctx, repo.UpdateSubscriptionParams{
		StripeSubscriptionID: payload.StripeSubscriptionID,
		StripePriceID:        payload.StripePriceID,
		Status:               payload.Status,
		CurrentPeriodStart:   pgtype.Timestamptz{Time: payload.CurrentPeriodStart.Time, Valid: true},
		CurrentPeriodEnd:     pgtype.Timestamptz{Time: payload.CurrentPeriodEnd.Time, Valid: true},
	})
}

func (s *svc) DeactivateSubscription(ctx context.Context, stripeSubscriptionID string) (repo.Subscription, error) {
	_, err := s.repo.GetSubscriptionByStripeSubscription(ctx, stripeSubscriptionID)
	if err != nil {
		return repo.Subscription{}, codeerror.New(codeerror.SubscriptionNotFound, "Subscription not found")
	}

	subscription, err := s.repo.DeactivateSubscription(ctx, stripeSubscriptionID)

	if err != nil {
		return repo.Subscription{}, codeerror.New(codeerror.FailedToDeactivateSubscription, "Failed to deactivate subscription")
	}

	return subscription, nil
}

func (s *svc) GetSubscription(
	subscriptionID string,
) (*stripe.Subscription, error) {
	return s.client.V1Subscriptions.Retrieve(
		context.Background(),
		subscriptionID,
		nil,
	)
}

func (s *svc) GetWorkspaceSubscription(ctx context.Context, workspaceID, userID string) (subscriptionResponse, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return subscriptionResponse{}, codeerror.New(codeerror.InvalidUUID, "Workspace ID is not a valid UUID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return subscriptionResponse{}, codeerror.New(codeerror.InvalidUUID, "User ID is not a valid UUID")
	}

	_, err = s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID:     pgtype.UUID{Bytes: workspaceUUID, Valid: true},
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err != nil {
		return subscriptionResponse{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	subscription, err := s.repo.GetWorkspaceActiveSubscription(ctx, pgtype.UUID{Bytes: workspaceUUID, Valid: true})
	if err != nil {
		return subscriptionResponse{}, codeerror.New(codeerror.SubscriptionNotFound, "Subscription not found")
	}

	return subscriptionResponse{
		ID:                   subscription.ID.String(),
		WorkspaceID:          subscription.WorkspaceID.String(),
		StripeSubscriptionID: subscription.StripeSubscriptionID,
		StripePriceID:        subscription.StripePriceID,
		Status:               subscription.Status,
		Plan:                 subscription.Plan,
		CurrentPeriodStart:   subscription.CurrentPeriodStart,
		CurrentPeriodEnd:     subscription.CurrentPeriodEnd,
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
	}, nil
}
