package workspace

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockWorkspaceRepository struct {
	getAccessibleWorkspaceByIDFunc func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error)
	getUserWorkspacesFunc          func(ctx context.Context, arg repo.GetUserWorkspacesParams) ([]repo.GetUserWorkspacesRow, error)
}

func (m *mockWorkspaceRepository) AddWorkspaceMember(ctx context.Context, arg repo.AddWorkspaceMemberParams) (repo.WorkspaceMember, error) {
	return repo.WorkspaceMember{}, nil
}

func (m *mockWorkspaceRepository) CountUserWorkspaces(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return 0, nil
}

func (m *mockWorkspaceRepository) CreateWorkspace(ctx context.Context, arg repo.CreateWorkspaceParams) (repo.Workspace, error) {
	return repo.Workspace{}, nil
}

func (m *mockWorkspaceRepository) DeleteWorkspace(ctx context.Context, arg repo.DeleteWorkspaceParams) error {
	return nil
}

func (m *mockWorkspaceRepository) GetAccessibleWorkspaceByID(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
	if m.getAccessibleWorkspaceByIDFunc == nil {
		return repo.Workspace{}, nil
	}
	return m.getAccessibleWorkspaceByIDFunc(ctx, arg)
}

func (m *mockWorkspaceRepository) GetUserActiveProSubscription(ctx context.Context, userID pgtype.UUID) (repo.Subscription, error) {
	return repo.Subscription{}, pgx.ErrNoRows
}

func (m *mockWorkspaceRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	return repo.User{}, nil
}

func (m *mockWorkspaceRepository) GetUserWorkspaceByID(ctx context.Context, arg repo.GetUserWorkspaceByIDParams) (repo.Workspace, error) {
	return repo.Workspace{}, nil
}

func (m *mockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, arg repo.GetUserWorkspacesParams) ([]repo.GetUserWorkspacesRow, error) {
	if m.getUserWorkspacesFunc == nil {
		return nil, nil
	}
	return m.getUserWorkspacesFunc(ctx, arg)
}

func (m *mockWorkspaceRepository) UpdateUserStripeCustomer(ctx context.Context, arg repo.UpdateUserStripeCustomerParams) (repo.User, error) {
	return repo.User{}, nil
}

func (m *mockWorkspaceRepository) UpdateWorkspace(ctx context.Context, arg repo.UpdateWorkspaceParams) (repo.Workspace, error) {
	return repo.Workspace{}, nil
}

func (m *mockWorkspaceRepository) WithTx(tx pgx.Tx) *repo.Queries {
	return nil
}

func assertErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *codeerror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *codeerror.Error, got %T", err)
	}

	if appErr.Code != code {
		t.Errorf("error code = %s, want %s", appErr.Code, code)
	}
}

func newTestWorkspace(id, userID uuid.UUID) repo.Workspace {
	return repo.Workspace{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		WorkspaceName: "Test Workspace",
		Description:   "A workspace",
		UserID:        pgtype.UUID{Bytes: userID, Valid: true},
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestGetUserWorkspaceByID(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	workspaceID := uuid.New()
	workspace := newTestWorkspace(workspaceID, ownerID)

	tests := []struct {
		name        string
		arg         repo.GetUserWorkspaceByIDParams
		repo        *mockWorkspaceRepository
		wantErr     bool
		wantCode    string
		wantMatchID bool
	}{
		{
			name: "owner can access workspace",
			arg: repo.GetUserWorkspaceByIDParams{
				ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
				UserID: pgtype.UUID{Bytes: ownerID, Valid: true},
			},
			repo: &mockWorkspaceRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					if arg.ID.Bytes != workspaceID {
						t.Errorf("workspace ID = %v, want %v", arg.ID.Bytes, workspaceID)
					}
					if arg.UserID.Bytes != ownerID {
						t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, ownerID)
					}
					return workspace, nil
				},
			},
			wantErr:     false,
			wantMatchID: true,
		},
		{
			name: "member can access workspace",
			arg: repo.GetUserWorkspaceByIDParams{
				ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
				UserID: pgtype.UUID{Bytes: memberID, Valid: true},
			},
			repo: &mockWorkspaceRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					if arg.UserID.Bytes != memberID {
						t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, memberID)
					}
					return workspace, nil
				},
			},
			wantErr:     false,
			wantMatchID: true,
		},
		{
			name: "non-member cannot access workspace",
			arg: repo.GetUserWorkspaceByIDParams{
				ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
				UserID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			},
			repo: &mockWorkspaceRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return repo.Workspace{}, pgx.ErrNoRows
				},
			},
			wantErr:  true,
			wantCode: codeerror.WorkspaceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetUserWorkspaceByID(ctx, tt.arg)

			if tt.wantErr {
				assertErrorCode(t, err, tt.wantCode)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantMatchID && result.ID.String() != workspaceID.String() {
				t.Errorf("workspace ID = %s, want %s", result.ID, workspaceID)
			}
		})
	}
}

func TestGetUserWorkspaces(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ownedID := uuid.New()
	memberID := uuid.New()
	now := time.Now()

	rows := []repo.GetUserWorkspacesRow{
		{
			TotalCount:    2,
			ID:            pgtype.UUID{Bytes: ownedID, Valid: true},
			WorkspaceName: "Owned Workspace",
			Description:   "Owned",
			UserID:        pgtype.UUID{Bytes: userID, Valid: true},
			CreatedAt:     pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			TotalCount:    2,
			ID:            pgtype.UUID{Bytes: memberID, Valid: true},
			WorkspaceName: "Member Workspace",
			Description:   "Member",
			UserID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			CreatedAt:     pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockRepo := &mockWorkspaceRepository{
		getUserWorkspacesFunc: func(ctx context.Context, arg repo.GetUserWorkspacesParams) ([]repo.GetUserWorkspacesRow, error) {
			if arg.UserID.Bytes != userID {
				t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, userID)
			}
			if arg.Limit != 10 || arg.Offset != 0 {
				t.Errorf("limit/offset = %d/%d, want 10/0", arg.Limit, arg.Offset)
			}
			return rows, nil
		},
	}

	service := &svc{repo: mockRepo}
	result, err := service.GetUserWorkspaces(ctx, pgtype.UUID{Bytes: userID, Valid: true}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Workspaces) != 2 {
		t.Fatalf("workspaces count = %d, want 2", len(result.Workspaces))
	}

	// Owned workspaces and member workspaces are returned together.
	if result.Workspaces[0].ID != ownedID.String() {
		t.Errorf("first workspace ID = %s, want %s", result.Workspaces[0].ID, ownedID)
	}

	if result.Workspaces[1].ID != memberID.String() {
		t.Errorf("second workspace ID = %s, want %s", result.Workspaces[1].ID, memberID)
	}

	if result.Pagination.Total != 2 || result.Pagination.TotalPages != 1 {
		t.Errorf("pagination = %+v, want total=2 totalPages=1", result.Pagination)
	}
}

func TestGetUserWorkspacesRepositoryFailure(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := &mockWorkspaceRepository{
		getUserWorkspacesFunc: func(ctx context.Context, arg repo.GetUserWorkspacesParams) ([]repo.GetUserWorkspacesRow, error) {
			return nil, errors.New("repo failure")
		},
	}

	service := &svc{repo: mockRepo}
	_, err := service.GetUserWorkspaces(ctx, pgtype.UUID{Bytes: userID, Valid: true}, 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
