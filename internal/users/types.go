package users

import "gin-api-1/internal/auth"

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type getUsersResponse struct {
	Users      []auth.UserResponse `json:"users"`
	Pagination paginationResponse  `json:"pagination"`
}

type updateUserProfilePayload struct {
	FirstName *string `json:"firstName,omitempty" binding:"omitempty,min=1,max=255"`
	LastName  *string `json:"lastName,omitempty" binding:"omitempty,min=1,max=255"`
}
