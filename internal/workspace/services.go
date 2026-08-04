package workspace

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateWorkspace(ctx context.Context, payload createWorkspacePayload) (repo.Workspace, error)
	GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID pgtype.Text, page, pageSize int) (getUserWorkspacesResponse, error)
	UpdateWorkspace(ctx context.Context, payload updateWorkspacePayload) (repo.Workspace, error)
	DeleteWorkspace(ctx context.Context, arg repo.DeleteWorkspaceParams) error
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

func (s *svc) GetUserWorkspaces(ctx context.Context, userID pgtype.Text, page, pageSize int) (getUserWorkspacesResponse, error) {
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
			UserID:        row.UserID.String,
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

	return s.repo.CreateWorkspace(ctx, repo.CreateWorkspaceParams{
		ID:            pgtype.UUID{Bytes: pk, Valid: true},
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
		UserID: pgtype.Text{
			String: payload.UserID,
			Valid:  true,
		},
	})
}

func (s *svc) UpdateWorkspace(ctx context.Context, payload updateWorkspacePayload) (repo.Workspace, error) {

	id, err := uuid.Parse(payload.ID)

	if err != nil {
		return repo.Workspace{}, err
	}

	_, err = s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
		UserID: pgtype.Text{
			String: payload.UserID,
			Valid:  true,
		},
	})

	if err != nil {
		return repo.Workspace{}, errors.New("workspace does not exist")
	}

	return s.repo.UpdateWorkspace(ctx, repo.UpdateWorkspaceParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
		UserID: pgtype.Text{
			String: payload.UserID,
			Valid:  true,
		},
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
	})
}

func (s *svc) DeleteWorkspace(ctx context.Context, arg repo.DeleteWorkspaceParams) error {
	_, err := s.repo.GetUserWorkspaceByID(ctx, repo.GetUserWorkspaceByIDParams{
		ID:     arg.ID,
		UserID: arg.UserID,
	})

	if err != nil {
		return errors.New("workspace does not exist")
	}

	return s.repo.DeleteWorkspace(ctx, arg)
}
