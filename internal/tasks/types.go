package tasks

import "github.com/jackc/pgx/v5/pgtype"

const DefaultPageSize = 10

type createTaskPayload struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	StartDate   string  `json:"startDate"`
	EndDate     string  `json:"endDate"`
	Status      string  `json:"status" binding:"required,oneof=TODO IN_PROGRESS TESTING DONE"`
	Priority    string  `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH URGENT"`
	AssigneeID  *string `json:"assigneeId"`
}

type updateTaskPayload struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=TODO IN_PROGRESS TESTING DONE"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
}

type taskResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	StartDate   pgtype.Date        `json:"startDate"`
	EndDate     pgtype.Date        `json:"endDate"`
	Status      string             `json:"status"`
	Priority    string             `json:"priority"`
	ProjectID   string             `json:"projectId"`
	AssigneeID  pgtype.Text        `json:"assigneeId"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type getProjectTasksResponse struct {
	Tasks      []taskResponse     `json:"tasks"`
	Pagination paginationResponse `json:"pagination"`
}
