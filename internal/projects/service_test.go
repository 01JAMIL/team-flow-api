package projects

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockProjectRepository struct {
	getAccessibleWorkspaceByIDFunc func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error)
	getProjectByIdFunc             func(ctx context.Context, id pgtype.UUID) (repo.Project, error)
	getWorkspaceProjectsFunc       func(ctx context.Context, arg repo.GetWorkspaceProjectsParams) ([]repo.GetWorkspaceProjectsRow, error)
}

func (m *mockProjectRepository) CountWorkspaceProjects(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
	return 0, nil
}

func (m *mockProjectRepository) CreateProject(ctx context.Context, arg repo.CreateProjectParams) (repo.Project, error) {
	return repo.Project{}, nil
}

func (m *mockProjectRepository) DeleteProject(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (m *mockProjectRepository) GetAccessibleWorkspaceByID(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
	if m.getAccessibleWorkspaceByIDFunc == nil {
		return repo.Workspace{}, nil
	}
	return m.getAccessibleWorkspaceByIDFunc(ctx, arg)
}

func (m *mockProjectRepository) GetMemberFromWorkspace(ctx context.Context, arg repo.GetMemberFromWorkspaceParams) (repo.WorkspaceMember, error) {
	return repo.WorkspaceMember{}, nil
}

func (m *mockProjectRepository) GetProjectById(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
	if m.getProjectByIdFunc == nil {
		return repo.Project{}, nil
	}
	return m.getProjectByIdFunc(ctx, id)
}

func (m *mockProjectRepository) GetUserActiveSubscription(ctx context.Context, userID pgtype.UUID) (repo.Subscription, error) {
	return repo.Subscription{}, pgx.ErrNoRows
}

func (m *mockProjectRepository) GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (repo.Workspace, error) {
	return repo.Workspace{}, nil
}

func (m *mockProjectRepository) GetWorkspaceProjects(ctx context.Context, arg repo.GetWorkspaceProjectsParams) ([]repo.GetWorkspaceProjectsRow, error) {
	if m.getWorkspaceProjectsFunc == nil {
		return nil, nil
	}
	return m.getWorkspaceProjectsFunc(ctx, arg)
}

func (m *mockProjectRepository) UpdateProject(ctx context.Context, arg repo.UpdateProjectParams) (repo.Project, error) {
	return repo.Project{}, nil
}

func assertProjectServiceErrorCode(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func newTestProject(id, workspaceID uuid.UUID) repo.Project {
	return repo.Project{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        "Test Project",
		Description: "A project",
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestGetWorkspaceProjects(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	workspaceID := uuid.New()
	projectID := uuid.New()

	projectRows := []repo.GetWorkspaceProjectsRow{
		{
			TotalCount:  1,
			ID:          pgtype.UUID{Bytes: projectID, Valid: true},
			Name:        "Test Project",
			Description: "A project",
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	workspace := repo.Workspace{
		ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID: pgtype.UUID{Bytes: ownerID, Valid: true},
	}

	tests := []struct {
		name        string
		workspaceID string
		loggedUser  string
		repo        *mockProjectRepository
		wantErr     bool
		wantLen     int
	}{
		{
			name:        "owner can list workspace projects",
			workspaceID: workspaceID.String(),
			loggedUser:  ownerID.String(),
			repo: &mockProjectRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return workspace, nil
				},
				getWorkspaceProjectsFunc: func(ctx context.Context, arg repo.GetWorkspaceProjectsParams) ([]repo.GetWorkspaceProjectsRow, error) {
					if arg.WorkspaceID.Bytes != workspaceID {
						t.Errorf("workspace ID = %v, want %v", arg.WorkspaceID.Bytes, workspaceID)
					}
					return projectRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:        "member can list workspace projects",
			workspaceID: workspaceID.String(),
			loggedUser:  memberID.String(),
			repo: &mockProjectRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					if arg.UserID.Bytes != memberID {
						t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, memberID)
					}
					return workspace, nil
				},
				getWorkspaceProjectsFunc: func(ctx context.Context, arg repo.GetWorkspaceProjectsParams) ([]repo.GetWorkspaceProjectsRow, error) {
					return projectRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:        "neither owner nor member cannot list workspace projects",
			workspaceID: workspaceID.String(),
			loggedUser:  nonMemberID.String(),
			repo: &mockProjectRepository{
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return repo.Workspace{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
		{
			name:        "invalid workspace id",
			workspaceID: "not-a-uuid",
			loggedUser:  ownerID.String(),
			repo:        &mockProjectRepository{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetWorkspaceProjects(ctx, tt.workspaceID, tt.loggedUser, 1, 10)

			if tt.wantErr {
				assertProjectServiceErrorCode(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Projects) != tt.wantLen {
				t.Errorf("projects count = %d, want %d", len(result.Projects), tt.wantLen)
			}
		})
	}
}

func TestGetProjectByID(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	workspaceID := uuid.New()
	projectID := uuid.New()
	project := newTestProject(projectID, workspaceID)

	workspace := repo.Workspace{
		ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID: pgtype.UUID{Bytes: ownerID, Valid: true},
	}

	tests := []struct {
		name       string
		projectID  string
		loggedUser string
		repo       *mockProjectRepository
		wantErr    bool
	}{
		{
			name:       "owner can fetch project",
			projectID:  projectID.String(),
			loggedUser: ownerID.String(),
			repo: &mockProjectRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return project, nil
				},
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return workspace, nil
				},
			},
			wantErr: false,
		},
		{
			name:       "member can fetch project",
			projectID:  projectID.String(),
			loggedUser: memberID.String(),
			repo: &mockProjectRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return project, nil
				},
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return workspace, nil
				},
			},
			wantErr: false,
		},
		{
			name:       "neither owner nor member cannot fetch project",
			projectID:  projectID.String(),
			loggedUser: nonMemberID.String(),
			repo: &mockProjectRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return project, nil
				},
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return repo.Workspace{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
		{
			name:       "project not found",
			projectID:  projectID.String(),
			loggedUser: ownerID.String(),
			repo: &mockProjectRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return repo.Project{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			_, err := service.GetProjectByID(ctx, tt.projectID, tt.loggedUser)

			if tt.wantErr {
				assertProjectServiceErrorCode(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetWorkspaceProjectsRepositoryFailure(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	workspaceID := uuid.New()

	workspace := repo.Workspace{
		ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID: pgtype.UUID{Bytes: ownerID, Valid: true},
	}

	mockRepo := &mockProjectRepository{
		getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
			return workspace, nil
		},
		getWorkspaceProjectsFunc: func(ctx context.Context, arg repo.GetWorkspaceProjectsParams) ([]repo.GetWorkspaceProjectsRow, error) {
			return nil, errors.New("repo failure")
		},
	}

	service := &svc{repo: mockRepo}
	_, err := service.GetWorkspaceProjects(ctx, workspaceID.String(), ownerID.String(), 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
