package workspacemembers

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockWorkspaceMemberRepository struct {
	getAccessibleWorkspaceByIDFunc func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error)
	getWorkspaceMembersFunc        func(ctx context.Context, arg repo.GetWorkspaceMembersParams) ([]repo.GetWorkspaceMembersRow, error)
}

func (m *mockWorkspaceMemberRepository) AddWorkspaceMember(ctx context.Context, arg repo.AddWorkspaceMemberParams) (repo.WorkspaceMember, error) {
	return repo.WorkspaceMember{}, nil
}

func (m *mockWorkspaceMemberRepository) DeleteMemberFromWorkspace(ctx context.Context, arg repo.DeleteMemberFromWorkspaceParams) error {
	return nil
}

func (m *mockWorkspaceMemberRepository) GetAccessibleWorkspaceByID(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
	if m.getAccessibleWorkspaceByIDFunc == nil {
		return repo.Workspace{}, nil
	}
	return m.getAccessibleWorkspaceByIDFunc(ctx, arg)
}

func (m *mockWorkspaceMemberRepository) GetMemberFromWorkspace(ctx context.Context, arg repo.GetMemberFromWorkspaceParams) (repo.WorkspaceMember, error) {
	return repo.WorkspaceMember{}, nil
}

func (m *mockWorkspaceMemberRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	return repo.User{}, nil
}

func (m *mockWorkspaceMemberRepository) GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (repo.Workspace, error) {
	return repo.Workspace{}, nil
}

func (m *mockWorkspaceMemberRepository) GetWorkspaceMembers(ctx context.Context, arg repo.GetWorkspaceMembersParams) ([]repo.GetWorkspaceMembersRow, error) {
	if m.getWorkspaceMembersFunc == nil {
		return nil, nil
	}
	return m.getWorkspaceMembersFunc(ctx, arg)
}

func TestGetWorkspaceMembers(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	workspaceID := uuid.New()
	memberRowUserID := uuid.New()

	workspace := repo.Workspace{
		ID:            pgtype.UUID{Bytes: workspaceID, Valid: true},
		WorkspaceName: "Test Workspace",
		Description:   "A workspace",
		UserID:        pgtype.UUID{Bytes: ownerID, Valid: true},
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	memberRows := []repo.GetWorkspaceMembersRow{
		{
			TotalCount:      1,
			MemberID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			WorkspaceID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
			UserRole:        "MEMBER",
			MemberCreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UserID:          pgtype.UUID{Bytes: memberRowUserID, Valid: true},
			FirstName:       "Jane",
			LastName:        "Doe",
			Email:           "jane@example.com",
			CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	tests := []struct {
		name        string
		workspaceID string
		loggedUser  string
		repo        *mockWorkspaceMemberRepository
		wantErr     bool
		wantLen     int
	}{
		{
			name:        "owner can list workspace members",
			workspaceID: workspaceID.String(),
			loggedUser:  ownerID.String(),
			repo: &mockWorkspaceMemberRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return workspace, nil
				},
				getWorkspaceMembersFunc: func(ctx context.Context, arg repo.GetWorkspaceMembersParams) ([]repo.GetWorkspaceMembersRow, error) {
					return memberRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:        "member can list workspace members",
			workspaceID: workspaceID.String(),
			loggedUser:  memberID.String(),
			repo: &mockWorkspaceMemberRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					if arg.UserID.Bytes != memberID {
						t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, memberID)
					}
					return workspace, nil
				},
				getWorkspaceMembersFunc: func(ctx context.Context, arg repo.GetWorkspaceMembersParams) ([]repo.GetWorkspaceMembersRow, error) {
					return memberRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:        "neither owner nor member cannot list workspace members",
			workspaceID: workspaceID.String(),
			loggedUser:  nonMemberID.String(),
			repo: &mockWorkspaceMemberRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return repo.Workspace{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetWorkspaceMembers(ctx, tt.workspaceID, tt.loggedUser, 1, 10)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Members) != tt.wantLen {
				t.Errorf("members count = %d, want %d", len(result.Members), tt.wantLen)
			}

			if result.Workspace.ID != workspaceID.String() {
				t.Errorf("workspace ID = %s, want %s", result.Workspace.ID, workspaceID)
			}
		})
	}
}
