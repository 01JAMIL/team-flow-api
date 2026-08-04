package workspacemembers

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	AddWorkspaceMember(ctx context.Context, workspaceID string, payload addWorkspaceMemberPayload) (repo.WorkspaceMember, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewWorkspaceMembersService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) AddWorkspaceMember(ctx context.Context, workspaceID string, payload addWorkspaceMemberPayload) (repo.WorkspaceMember, error) {

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return repo.WorkspaceMember{}, err
	}

	_, err = s.repo.GetWorkspaceById(ctx, pgtype.UUID{Bytes: workspaceUUID, Valid: true})
	if err != nil {
		return repo.WorkspaceMember{}, errors.New("workspace does not exist")
	}

	userUUID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return repo.WorkspaceMember{}, err
	}
	_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		return repo.WorkspaceMember{}, errors.New("user does not exist")
	}

	_, err = s.repo.GetMemberFromWorkspace(ctx, repo.GetMemberFromWorkspaceParams{
		UserID:      pgtype.Text{String: payload.UserID, Valid: true},
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
	})

	if err == nil {
		return repo.WorkspaceMember{}, errors.New("user is already a member of this workspace")
	}

	pk := uuid.New()

	return s.repo.AddWorkspaceMember(ctx, repo.AddWorkspaceMemberParams{
		ID:          pgtype.UUID{Bytes: pk, Valid: true},
		UserID:      pgtype.Text{String: payload.UserID, Valid: true},
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
		UserRole:    payload.UserRole,
	})
}
