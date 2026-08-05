package projects

import "github.com/jackc/pgx/v5/pgtype"

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
