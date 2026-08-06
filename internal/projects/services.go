package projects

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	WorkspaceExists(ctx context.Context, workspaceID string) error
	CreateProject(ctx context.Context, workspaceID string, payload createProjectPayload) (projectResponse, error)
	GetWorkspaceProjects(ctx context.Context, workspaceID string, page, pageSize int) (getWorkspaceProjectsResponse, error)
	GetProjectByID(ctx context.Context, projectID string) (projectResponse, error)
	UpdateProject(ctx context.Context, projectID string, payload updateProjectPayload) (projectResponse, error)
	DeleteProject(ctx context.Context, projectID string) error
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewProjectsService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) WorkspaceExists(ctx context.Context, workspaceID string) error {
	_, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}
	return nil
}

func (s *svc) GetWorkspaceProjects(ctx context.Context, workspaceID string, page, pageSize int) (getWorkspaceProjectsResponse, error) {
	_, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return getWorkspaceProjectsResponse{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	rows, err := s.repo.GetWorkspaceProjects(ctx, repo.GetWorkspaceProjectsParams{
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
		Limit:       int32(pageSize),
		Offset:      int32((page - 1) * pageSize),
	})
	if err != nil {
		return getWorkspaceProjectsResponse{}, err
	}

	projects := make([]projectResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		projects = append(projects, projectResponse{
			ID:          row.ID.String(),
			Name:        row.Name,
			Description: row.Description,
			WorkspaceID: row.WorkspaceID.String,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return getWorkspaceProjectsResponse{
		Projects: projects,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *svc) CreateProject(ctx context.Context, workspaceID string, payload createProjectPayload) (projectResponse, error) {
	_, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return projectResponse{}, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found")
	}

	pk := uuid.New()

	project, err := s.repo.CreateProject(ctx, repo.CreateProjectParams{
		ID:          pgtype.UUID{Bytes: pk, Valid: true},
		Name:        payload.Name,
		Description: payload.Description,
		WorkspaceID: pgtype.Text{String: workspaceID, Valid: true},
	})
	if err != nil {
		return projectResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create project", err)
	}

	return toProjectResponse(project), nil
}

func (s *svc) GetProjectByID(ctx context.Context, projectID string) (projectResponse, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return projectResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	project, err := s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return projectResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	return toProjectResponse(project), nil
}

func (s *svc) UpdateProject(ctx context.Context, projectID string, payload updateProjectPayload) (projectResponse, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return projectResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return projectResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	project, err := s.repo.UpdateProject(ctx, repo.UpdateProjectParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		return projectResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to update project", err)
	}

	return toProjectResponse(project), nil
}

func (s *svc) DeleteProject(ctx context.Context, projectID string) error {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	_, err = s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	if err := s.repo.DeleteProject(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		return codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to delete project", err)
	}

	return nil
}

func toProjectResponse(project repo.Project) projectResponse {
	return projectResponse{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		WorkspaceID: project.WorkspaceID.String,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}
