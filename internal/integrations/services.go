package integrations

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateIntegrationTask(ctx context.Context, payload createIntegrationTaskParams) (repo.IntegrationTask, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewIntegrationsService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) CreateIntegrationTask(ctx context.Context, payload createIntegrationTaskParams) (repo.IntegrationTask, error) {
	pk := uuid.New()

	integrationTask, err := s.repo.CreateIntegrationTask(ctx, repo.CreateIntegrationTaskParams{
		ID:             pgtype.UUID{Bytes: pk, Valid: true},
		Provider:       payload.Provider,
		ResourceType:   payload.ResourceType,
		ExternalID:     payload.ExternalID,
		RepositoryName: payload.RepositoryName,
		IssueNumber:    payload.IssueNumber,
		Title:          payload.Title,
		Description:    pgtype.Text{String: payload.Description, Valid: true},
		Status:         payload.Status,
		AssigneeID:     payload.AssigneeID,
		Payload:        payload.Payload,
	})

	if err != nil {
		return repo.IntegrationTask{}, err
	}

	return integrationTask, nil
}
