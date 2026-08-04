package workspace

import "github.com/jackc/pgx/v5/pgtype"

const DefaultPageSize = 10

type createWorkspacePayload struct {
	WorkspaceName string `json:"workspaceName" binding:"required"`
	Description   string `json:"description" binding:"required"`
	UserID        string `json:"userID,omitempty"`
}

type updateWorkspacePayload struct {
	ID            string `json:"id,omitempty"`
	UserID        string `json:"userID,omitempty"`
	WorkspaceName string `json:"workspaceName,omitempty"`
	Description   string `json:"description,omitempty"`
}

type workspaceResponse struct {
	ID            string             `json:"id"`
	WorkspaceName string             `json:"workspaceName"`
	Description   string             `json:"description"`
	UserID        string             `json:"userId"`
	CreatedAt     pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt     pgtype.Timestamptz `json:"updatedAt"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type getUserWorkspacesResponse struct {
	Workspaces []workspaceResponse `json:"workspaces"`
	Pagination paginationResponse  `json:"pagination"`
}
