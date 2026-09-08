package dashboard

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	GetKPIs(ctx context.Context, userID string) (kpiResponse, error)
}

type svc struct {
	repo *repo.Queries
}

func NewDashboardService(repo *repo.Queries) Service {
	return &svc{
		repo: repo,
	}
}

func (s *svc) GetKPIs(ctx context.Context, userID string) (kpiResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return kpiResponse{}, codeerror.New(codeerror.InvalidUUID, "Invalid user ID")
	}

	row, err := s.repo.GetUserKPIs(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		return kpiResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to fetch dashboard KPIs", err)
	}

	return kpiResponse{
		Workspaces:     row.TotalWorkspaces,
		ActiveProjects: row.ActiveProjects,
		OpenTasks:      row.OpenTasks,
		TeamMembers:    row.TeamMembers,
	}, nil
}
