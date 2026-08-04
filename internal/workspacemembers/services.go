package workspacemembers

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
	AddWorkspaceMember(ctx context.Context, workspaceID string, payload addWorkspaceMemberPayload) (repo.WorkspaceMember, error)
	GetWorkspaceMembers(ctx context.Context, workspaceID string, page, pageSize int) (getWorkspaceMembersResponse, error)
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

func (s *svc) GetWorkspaceMembers(ctx context.Context, workspaceID string, page, pageSize int) (getWorkspaceMembersResponse, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return getWorkspaceMembersResponse{}, err
	}

	workspace, err := s.repo.GetWorkspaceById(ctx, pgtype.UUID{Bytes: workspaceUUID, Valid: true})
	if err != nil {
		return getWorkspaceMembersResponse{}, ErrWorkspaceNotFound
	}

	rows, err := s.repo.GetWorkspaceMembers(ctx, repo.GetWorkspaceMembersParams{
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
		Limit:       int32(pageSize),
		Offset:      int32((page - 1) * pageSize),
	})
	if err != nil {
		return getWorkspaceMembersResponse{}, err
	}

	members := make([]workspaceMemberResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		members = append(members, workspaceMemberResponse{
			ID:          row.MemberID.String(),
			UserID:      row.UserID.String(),
			WorkspaceID: row.WorkspaceID.String,
			UserRole:    row.UserRole,
			CreatedAt:   row.MemberCreatedAt,
			User: memberUserResponse{
				ID:        row.UserID.String(),
				FirstName: row.FirstName,
				LastName:  row.LastName,
				Email:     row.Email,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return getWorkspaceMembersResponse{
		Workspace: workspaceResponse{
			ID:            workspace.ID.String(),
			WorkspaceName: workspace.WorkspaceName,
			Description:   workspace.Description,
			UserID:        workspace.UserID.String,
			CreatedAt:     workspace.CreatedAt,
			UpdatedAt:     workspace.UpdatedAt,
		},
		Members: members,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
