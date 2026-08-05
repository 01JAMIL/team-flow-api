package projects

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrWorkspaceNotFound = errors.New("workspace does not exist")

type Service interface {
	WorkspaceExists(ctx context.Context, workspaceID string) error
	CreateProject(ctx context.Context, workspaceID string, payload createProjectPayload) (projectResponse, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewProjectsService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) WorkspaceExists(ctx context.Context, workspaceID string) error {
	_, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return ErrWorkspaceNotFound
	}
	return nil
}

func (s *svc) CreateProject(ctx context.Context, workspaceID string, payload createProjectPayload) (projectResponse, error) {
	_, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return projectResponse{}, ErrWorkspaceNotFound
	}

	pk := uuid.New()

	project, err := s.repo.CreateProject(ctx, repo.CreateProjectParams{
		ID:          pgtype.UUID{Bytes: pk, Valid: true},
		Name:        payload.Name,
		Description: payload.Description,
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
	})
	if err != nil {
		return projectResponse{}, err
	}

	return projectResponse{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		WorkspaceID: project.WorkspaceID.String,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}
