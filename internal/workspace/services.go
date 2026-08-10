package workspace

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"
	"gin-api-1/internal/env"
	"gin-api-1/internal/payment"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v86"
)

type Service interface {
	CreateWorkspace(ctx context.Context, payload createWorkspacePayload) (repo.Workspace, error)
	GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID pgtype.UUID, page, pageSize int) (getUserWorkspacesResponse, error)
	UpdateWorkspace(ctx context.Context, payload updateWorkspacePayload) (repo.Workspace, error)
	DeleteWorkspace(ctx context.Context, arg repo.DeleteWorkspaceParams) error
	CreateCheckoutSession(ctx context.Context, workspaceID uuid.UUID, userID string) (*stripe.CheckoutSession, error)
}

type svc struct {
	repo   *repo.Queries
	db     *pgx.Conn
	stripe payment.Service
}

func NewWorkspaceService(repo *repo.Queries, db *pgx.Conn, stripe payment.Service) Service {
	return &svc{
		repo:   repo,
		db:     db,
		stripe: stripe,
	}
}

func (s *svc) GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error) {
	workspace, err := s.repo.GetUserWorkspaceByID(ctx, arg)
	if err != nil {
		return repo.Workspace{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	return workspace, nil
}

func (s *svc) GetUserWorkspaces(ctx context.Context, userID pgtype.UUID, page, pageSize int) (getUserWorkspacesResponse, error) {
	rows, err := s.repo.GetUserWorkspaces(ctx, repo.GetUserWorkspacesParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return getUserWorkspacesResponse{}, err
	}

	workspaces := make([]workspaceResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		workspaces = append(workspaces, workspaceResponse{
			ID:            row.ID.String(),
			WorkspaceName: row.WorkspaceName,
			Description:   row.Description,
			UserID:        row.UserID.String(),
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return getUserWorkspacesResponse{
		Workspaces: workspaces,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *svc) CreateWorkspace(ctx context.Context, payload createWorkspacePayload) (repo.Workspace, error) {
	pk := uuid.New()

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return repo.Workspace{}, codeerror.New(codeerror.UserNotFound, "User not found")
	}

	workspace, err := s.repo.CreateWorkspace(ctx, repo.CreateWorkspaceParams{
		ID:            pgtype.UUID{Bytes: pk, Valid: true},
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
		UserID:        pgtype.UUID{Bytes: userID, Valid: true},
	})

	if err != nil {
		return repo.Workspace{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create workspace", err)
	}

	return workspace, nil
}

func (s *svc) UpdateWorkspace(ctx context.Context, payload updateWorkspacePayload) (repo.Workspace, error) {
	id, err := uuid.Parse(payload.ID)

	if err != nil {
		return repo.Workspace{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return repo.Workspace{}, codeerror.New(codeerror.UserNotFound, "User not found")
	}

	_, err = s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})

	if err != nil {
		return repo.Workspace{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	workspace, err := s.repo.UpdateWorkspace(ctx, repo.UpdateWorkspaceParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		UserID:        pgtype.UUID{Bytes: userID, Valid: true},
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
	})
	if err != nil {
		return repo.Workspace{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to update workspace", err)
	}

	return workspace, nil
}

func (s *svc) DeleteWorkspace(ctx context.Context, arg repo.DeleteWorkspaceParams) error {
	_, err := s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID:     arg.ID,
		UserID: arg.UserID,
	})

	if err != nil {
		return codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	if err := s.repo.DeleteWorkspace(ctx, arg); err != nil {
		return codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to delete workspace", err)
	}

	return nil
}

func (s *svc) CreateCheckoutSession(ctx context.Context, workspaceID uuid.UUID, userID string) (*stripe.CheckoutSession, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, codeerror.New(codeerror.InvalidUUID, "User ID invalid")
	}

	workspace, err := s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err != nil {
		return nil, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	if !workspace.StripeCustomerID.Valid {
		customer, err := s.stripe.CreateStripeCustomer(
			workspace.WorkspaceName,
		)

		if err != nil {
			return nil, codeerror.Wrap(
				codeerror.StatusInternalServerError,
				"Failed to create Stripe customer",
				err,
			)
		}

		workspace, err = s.repo.UpdateWorkspaceStripeCustomer(
			ctx,
			repo.UpdateWorkspaceStripeCustomerParams{
				ID: workspace.ID,
				StripeCustomerID: pgtype.Text{
					String: customer.ID,
					Valid:  true,
				},
			},
		)

		if err != nil {
			return nil, codeerror.Wrap(
				codeerror.StatusInternalServerError,
				"Failed to save Stripe customer",
				err,
			)
		}
	}

	session, err := s.stripe.CreateCheckoutSession(
		workspace.StripeCustomerID.String,
		env.GetEnvString("STRIPE_PRO_PRICE_ID", "price_xx"),
	)

	if err != nil {
		return nil, codeerror.Wrap(
			codeerror.StatusInternalServerError,
			"Failed to create checkout session",
			err,
		)
	}

	return session, nil
}
