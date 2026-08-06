package tasks

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

var (
	ErrProjectNotFound = errors.New("project does not exist")
	ErrUserNotFound    = errors.New("user does not exist")
	ErrInvalidDate     = errors.New("invalid date format, expected YYYY-MM-DD")
)

type Service interface {
	CreateTask(ctx context.Context, projectID string, payload createTaskPayload) (taskResponse, error)
	GetProjectTasks(ctx context.Context, projectID string, page, pageSize int) (getProjectTasksResponse, error)
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
		return taskResponse{}, ErrProjectNotFound
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return taskResponse{}, ErrProjectNotFound
	}

	var assigneeID pgtype.Text
	if payload.AssigneeID != nil {
		userUUID, err := uuid.Parse(*payload.AssigneeID)
		if err != nil {
			return taskResponse{}, ErrUserNotFound
		}

		_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			return taskResponse{}, ErrUserNotFound
		}

		assigneeID = pgtype.Text{String: *payload.AssigneeID, Valid: true}
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
		return taskResponse{}, err
	}

	return toTaskResponse(task), nil
}

func (s *svc) GetProjectTasks(ctx context.Context, projectID string, page, pageSize int) (getProjectTasksResponse, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return getProjectTasksResponse{}, ErrProjectNotFound
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return getProjectTasksResponse{}, ErrProjectNotFound
	}

	rows, err := s.repo.GetProjectTasks(ctx, repo.GetProjectTasksParams{
		ProjectID: pgtype.UUID{Bytes: id, Valid: true},
		Limit:     int32(pageSize),
		Offset:    int32((page - 1) * pageSize),
	})
	if err != nil {
		return getProjectTasksResponse{}, err
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
			AssigneeID:  row.AssigneeID,
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
		return pgtype.Date{}, ErrInvalidDate
	}

	return pgtype.Date{Time: t, Valid: true}, nil
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
		AssigneeID:  task.AssigneeID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
