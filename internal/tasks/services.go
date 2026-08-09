package tasks

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

type Service interface {
	CreateTask(ctx context.Context, projectID string, payload createTaskPayload) (taskResponse, error)
	GetTaskByID(ctx context.Context, taskID string) (taskResponse, error)
	GetProjectTasks(ctx context.Context, projectID string, page, pageSize int) (getProjectTasksResponse, error)
	UpdateTask(ctx context.Context, taskID string, payload updateTaskPayload) (taskResponse, error)
	DeleteTask(ctx context.Context, taskID string) error
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewTasksService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) CreateTask(ctx context.Context, projectID string, payload createTaskPayload) (taskResponse, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	var assigneeID pgtype.UUID
	if payload.AssigneeID != nil {
		userUUID, err := uuid.Parse(*payload.AssigneeID)
		if err != nil {
			return taskResponse{}, codeerror.New(codeerror.UserNotFound, "User not found")
		}

		_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			return taskResponse{}, codeerror.New(codeerror.UserNotFound, "User not found")
		}

		assigneeID = pgtype.UUID{Bytes: userUUID, Valid: true}
	}

	startDate, err := parseDate(payload.StartDate)
	if err != nil {
		return taskResponse{}, err
	}

	endDate, err := parseDate(payload.EndDate)
	if err != nil {
		return taskResponse{}, err
	}

	task, err := s.repo.CreateTask(ctx, repo.CreateTaskParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:        payload.Name,
		Description: payload.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      payload.Status,
		Priority:    payload.Priority,
		ProjectID:   pgtype.UUID{Bytes: id, Valid: true},
		AssigneeID:  assigneeID,
	})
	if err != nil {
		return taskResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create task", err)
	}

	return toTaskResponse(task), nil
}

func (s *svc) GetTaskByID(ctx context.Context, taskID string) (taskResponse, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	task, err := s.repo.GetTaskById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	return toTaskResponse(task), nil
}

func (s *svc) UpdateTask(ctx context.Context, taskID string, payload updateTaskPayload) (taskResponse, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	_, err = s.repo.GetTaskById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return taskResponse{}, codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	var assigneeID pgtype.UUID
	if payload.AssigneeID != nil {
		assigneeUUID, err := uuid.Parse(*payload.AssigneeID)
		if err != nil {
			return taskResponse{}, codeerror.New(codeerror.UserNotFound, "User not found")
		}

		_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: assigneeUUID, Valid: true})
		if err != nil {
			return taskResponse{}, codeerror.New(codeerror.UserNotFound, "User not found")
		}

		assigneeID = pgtype.UUID{Bytes: assigneeUUID, Valid: true}
	}

	startDate, err := parseDatePtr(payload.StartDate)
	if err != nil {
		return taskResponse{}, err
	}

	endDate, err := parseDatePtr(payload.EndDate)
	if err != nil {
		return taskResponse{}, err
	}

	task, err := s.repo.UpdateTask(ctx, repo.UpdateTaskParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        stringPtr(payload.Name),
		Description: stringPtr(payload.Description),
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      stringPtr(payload.Status),
		Priority:    stringPtr(payload.Priority),
		AssigneeID:  assigneeID,
	})
	if err != nil {
		return taskResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to update task", err)
	}

	return toTaskResponse(task), nil
}

func (s *svc) DeleteTask(ctx context.Context, taskID string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	_, err = s.repo.GetTaskById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return codeerror.New(codeerror.TaskNotFound, "Task not found")
	}

	if err := s.repo.DeleteTask(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		return codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to delete task", err)
	}

	return nil
}

func (s *svc) GetProjectTasks(ctx context.Context, projectID string, page, pageSize int) (getProjectTasksResponse, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return getProjectTasksResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return getProjectTasksResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	rows, err := s.repo.GetProjectTasks(ctx, repo.GetProjectTasksParams{
		ProjectID: pgtype.UUID{Bytes: id, Valid: true},
		Limit:     int32(pageSize),
		Offset:    int32((page - 1) * pageSize),
	})
	if err != nil {
		return getProjectTasksResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to get project tasks", err)
	}

	tasks := make([]taskResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		tasks = append(tasks, taskResponse{
			ID:          row.ID.String(),
			Name:        row.Name,
			Description: row.Description,
			StartDate:   row.StartDate,
			EndDate:     row.EndDate,
			Status:      row.Status,
			Priority:    row.Priority,
			ProjectID:   row.ProjectID.String(),
			AssigneeID:  pgtype.Text{String: row.AssigneeID.String(), Valid: row.AssigneeID.Valid},
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return getProjectTasksResponse{
		Tasks: tasks,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func parseDate(value string) (pgtype.Date, error) {
	if value == "" {
		return pgtype.Date{}, nil
	}

	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return pgtype.Date{}, codeerror.New(codeerror.InvalidDate, "Invalid date format, expected YYYY-MM-DD")
	}

	return pgtype.Date{Time: t, Valid: true}, nil
}

func parseDatePtr(value *string) (pgtype.Date, error) {
	if value == nil {
		return pgtype.Date{}, nil
	}

	return parseDate(*value)
}

func stringPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{String: *value, Valid: true}
}

func toTaskResponse(task repo.Task) taskResponse {
	return taskResponse{
		ID:          task.ID.String(),
		Name:        task.Name,
		Description: task.Description,
		StartDate:   task.StartDate,
		EndDate:     task.EndDate,
		Status:      task.Status,
		Priority:    task.Priority,
		ProjectID:   task.ProjectID.String(),
		AssigneeID:  pgtype.Text{String: task.AssigneeID.String(), Valid: task.AssigneeID.Valid},
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
