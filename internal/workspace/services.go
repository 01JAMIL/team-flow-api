package workspace

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateWorkspace(ctx context.Context, payload createWorkspacePayload) (repo.Workspace, error)
	GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID pgtype.Text) ([]repo.Workspace, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewWorkspaceService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error) {
	return s.repo.GetUserWorkspaceByID(ctx, arg)
}

func (s *svc) GetUserWorkspaces(ctx context.Context, userID pgtype.Text) ([]repo.Workspace, error) {
	return s.repo.GetUserWorkspaces(ctx, userID)
}

func (s *svc) CreateWorkspace(ctx context.Context, payload createWorkspacePayload) (repo.Workspace, error) {
	pk := uuid.New().String()

	return s.repo.CreateWorkspace(ctx, repo.CreateWorkspaceParams{
		ID:            pk,
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
		UserID: pgtype.Text{
			String: payload.UserID,
			Valid:  true,
		},
	})
}
