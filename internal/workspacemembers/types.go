package workspacemembers

import (
	"github.com/jackc/pgx/v5/pgtype"
)

const DefaultPageSize = 10

type addWorkspaceMemberPayload struct {
	UserID   string `json:"userId" binding:"required"`
	UserRole string `json:"userRole" binding:"required,oneof=ADMIN MEMBER"`
}

type memberUserResponse struct {
	ID        string             `json:"id"`
	FirstName string             `json:"firstName"`
	LastName  string             `json:"lastName"`
	Email     string             `json:"email"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt pgtype.Timestamptz `json:"updatedAt"`
}

type workspaceMemberResponse struct {
	ID          string             `json:"id"`
	UserID      string             `json:"userId"`
	WorkspaceID string             `json:"workspaceId"`
	UserRole    string             `json:"userRole"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	User        memberUserResponse `json:"user"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type workspaceResponse struct {
	ID            string             `json:"id"`
	WorkspaceName string             `json:"workspaceName"`
	Description   string             `json:"description"`
	UserID        string             `json:"userId"`
	CreatedAt     pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt     pgtype.Timestamptz `json:"updatedAt"`
}

type getWorkspaceMembersResponse struct {
	Workspace  workspaceResponse         `json:"workspace"`
	Members    []workspaceMemberResponse `json:"members"`
	Pagination paginationResponse        `json:"pagination"`
}
