package messages

import "time"

const DefaultPageSize = 10

type CreateMessagePayload struct {
	ReceiverID string `json:"receiverId" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"senderId"`
	ReceiverID string    `json:"receiverId"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type getMessagesResponse struct {
	Messages   []MessageResponse  `json:"messages"`
	Pagination paginationResponse `json:"pagination"`
}
