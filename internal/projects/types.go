package projects

import "github.com/jackc/pgx/v5/pgtype"

const DefaultPageSize = 10

type createProjectPayload struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type projectResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	WorkspaceID string             `json:"workspaceId"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type getWorkspaceProjectsResponse struct {
	Projects   []projectResponse  `json:"projects"`
	Pagination paginationResponse `json:"pagination"`
}
