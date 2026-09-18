package tasks

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

type mockTaskRepository struct {
	getAccessibleWorkspaceByIDFunc func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error)
	getProjectByIdFunc             func(ctx context.Context, id pgtype.UUID) (repo.Project, error)
	getProjectTasksFunc            func(ctx context.Context, arg repo.GetProjectTasksParams) ([]repo.GetProjectTasksRow, error)
	getTaskByIdFunc                func(ctx context.Context, id pgtype.UUID) (repo.Task, error)
}

func (m *mockTaskRepository) CreateTask(ctx context.Context, arg repo.CreateTaskParams) (repo.Task, error) {
	return repo.Task{}, nil
}

func (m *mockTaskRepository) DeleteTask(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (m *mockTaskRepository) GetAccessibleWorkspaceByID(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
	if m.getAccessibleWorkspaceByIDFunc == nil {
		return repo.Workspace{}, nil
	}
	return m.getAccessibleWorkspaceByIDFunc(ctx, arg)
}

func (m *mockTaskRepository) GetProjectById(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
	if m.getProjectByIdFunc == nil {
		return repo.Project{}, nil
	}
	return m.getProjectByIdFunc(ctx, id)
}

func (m *mockTaskRepository) GetProjectTasks(ctx context.Context, arg repo.GetProjectTasksParams) ([]repo.GetProjectTasksRow, error) {
	if m.getProjectTasksFunc == nil {
		return nil, nil
	}
	return m.getProjectTasksFunc(ctx, arg)
}

func (m *mockTaskRepository) GetTaskById(ctx context.Context, id pgtype.UUID) (repo.Task, error) {
	if m.getTaskByIdFunc == nil {
		return repo.Task{}, nil
	}
	return m.getTaskByIdFunc(ctx, id)
}

func (m *mockTaskRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	return repo.User{}, nil
}

func (m *mockTaskRepository) UpdateTask(ctx context.Context, arg repo.UpdateTaskParams) (repo.Task, error) {
	return repo.Task{}, nil
}

func assertTaskServiceError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func newTestTask(id, projectID uuid.UUID) repo.Task {
	return repo.Task{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        "Test Task",
		Description: "A task",
		Status:      "TODO",
		Priority:    "MEDIUM",
		ProjectID:   pgtype.UUID{Bytes: projectID, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func accessibleWorkspace(workspaceID, ownerID uuid.UUID) repo.Workspace {
	return repo.Workspace{
		ID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID: pgtype.UUID{Bytes: ownerID, Valid: true},
	}
}

func TestGetProjectTasks(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	workspaceID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()

	project := newTestProject(projectID, workspaceID)
	workspace := accessibleWorkspace(workspaceID, ownerID)

	taskRows := []repo.GetProjectTasksRow{
		{
			TotalCount:  1,
			ID:          pgtype.UUID{Bytes: taskID, Valid: true},
			Name:        "Test Task",
			Description: "A task",
			Status:      "TODO",
			Priority:    "MEDIUM",
			ProjectID:   pgtype.UUID{Bytes: projectID, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	tests := []struct {
		name       string
		projectID  string
		loggedUser string
		repo       *mockTaskRepository
		wantErr    bool
		wantLen    int
	}{
		{
			name:       "owner can list project tasks",
			projectID:  projectID.String(),
			loggedUser: ownerID.String(),
			repo: &mockTaskRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return project, nil
				},
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					return workspace, nil
				},
				getProjectTasksFunc: func(ctx context.Context, arg repo.GetProjectTasksParams) ([]repo.GetProjectTasksRow, error) {
					if arg.ProjectID.Bytes != projectID {
						t.Errorf("project ID = %v, want %v", arg.ProjectID.Bytes, projectID)
					}
					return taskRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:       "member can list project tasks",
			projectID:  projectID.String(),
			loggedUser: memberID.String(),
			repo: &mockTaskRepository{
				getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
					return project, nil
				},
				getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
					if arg.UserID.Bytes != memberID {
						t.Errorf("user ID = %v, want %v", arg.UserID.Bytes, memberID)
					}
					return workspace, nil
				},
				getProjectTasksFunc: func(ctx context.Context, arg repo.GetProjectTasksParams) ([]repo.GetProjectTasksRow, error) {
					return taskRows, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:       "neither owner nor member cannot list project tasks",
			projectID:  projectID.String(),
			loggedUser: nonMemberID.String(),
			repo: &mockTaskRepository{
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
			name:       "invalid project id",
			projectID:  "not-a-uuid",
			loggedUser: ownerID.String(),
			repo:       &mockTaskRepository{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetProjectTasks(ctx, tt.projectID, tt.loggedUser, 1, 10)

			if tt.wantErr {
				assertTaskServiceError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Tasks) != tt.wantLen {
				t.Errorf("tasks count = %d, want %d", len(result.Tasks), tt.wantLen)
			}
		})
	}
}

func TestGetTaskByID(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	workspaceID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()

	project := newTestProject(projectID, workspaceID)
	workspace := accessibleWorkspace(workspaceID, ownerID)
	task := newTestTask(taskID, projectID)

	tests := []struct {
		name       string
		taskID     string
		loggedUser string
		repo       *mockTaskRepository
		wantErr    bool
	}{
		{
			name:       "owner can fetch task",
			taskID:     taskID.String(),
			loggedUser: ownerID.String(),
			repo: &mockTaskRepository{
				getTaskByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Task, error) {
					return task, nil
				},
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
			name:       "member can fetch task",
			taskID:     taskID.String(),
			loggedUser: memberID.String(),
			repo: &mockTaskRepository{
				getTaskByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Task, error) {
					return task, nil
				},
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
			name:       "neither owner nor member cannot fetch task",
			taskID:     taskID.String(),
			loggedUser: nonMemberID.String(),
			repo: &mockTaskRepository{
				getTaskByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Task, error) {
					return task, nil
				},
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
			name:       "task not found",
			taskID:     taskID.String(),
			loggedUser: ownerID.String(),
			repo: &mockTaskRepository{
				getTaskByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Task, error) {
					return repo.Task{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			_, err := service.GetTaskByID(ctx, tt.taskID, tt.loggedUser)

			if tt.wantErr {
				assertTaskServiceError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetProjectTasksRepositoryFailure(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	workspaceID := uuid.New()
	projectID := uuid.New()

	project := newTestProject(projectID, workspaceID)
	workspace := accessibleWorkspace(workspaceID, ownerID)

	mockRepo := &mockTaskRepository{
		getProjectByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.Project, error) {
			return project, nil
		},
		getAccessibleWorkspaceByIDFunc: func(ctx context.Context, arg repo.GetAccessibleWorkspaceByIDParams) (repo.Workspace, error) {
			return workspace, nil
		},
		getProjectTasksFunc: func(ctx context.Context, arg repo.GetProjectTasksParams) ([]repo.GetProjectTasksRow, error) {
			return nil, errors.New("repo failure")
		},
	}

	service := &svc{repo: mockRepo}
	_, err := service.GetProjectTasks(ctx, projectID.String(), ownerID.String(), 1, 10)
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
