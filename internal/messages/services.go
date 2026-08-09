package messages

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateMessage(ctx context.Context, senderID string, payload CreateMessagePayload) (repo.Message, error)
	GetMessagesBetweenUsers(ctx context.Context, loggedUserID, otherUserID string, page, pageSize int) (getMessagesResponse, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewMessagesService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) CreateMessage(ctx context.Context, senderID string, payload CreateMessagePayload) (repo.Message, error) {
	uuidSenderId, err := uuid.Parse(senderID)
	if err != nil {
		return repo.Message{}, codeerror.New(codeerror.InvalidUUID, "Sender ID is not a valid UUID")
	}

	receiverID, err := uuid.Parse(payload.ReceiverID)
	if err != nil {
		return repo.Message{}, codeerror.New(codeerror.InvalidUUID, "Receiver ID is not a valid UUID")
	}

	if uuidSenderId == receiverID {
		return repo.Message{}, codeerror.New(codeerror.SelfMessageNotAllowed, "You cannot send a message to yourself")
	}

	_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: receiverID, Valid: true})
	if err != nil {
		return repo.Message{}, codeerror.New(codeerror.UserNotFound, "User not found")
	}

	pk := uuid.New()

	return s.repo.CreateMessage(ctx, repo.CreateMessageParams{
		ID:         pgtype.UUID{Bytes: pk, Valid: true},
		SenderID:   pgtype.UUID{Bytes: uuidSenderId, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: receiverID, Valid: true},
		Content:    payload.Content,
	})
}

func (s *svc) GetMessagesBetweenUsers(ctx context.Context, loggedUserID, otherUserID string, page, pageSize int) (getMessagesResponse, error) {
	senderUUID, err := uuid.Parse(loggedUserID)
	if err != nil {
		return getMessagesResponse{}, codeerror.New(codeerror.InvalidUUID, "Sender ID is not a valid UUID")
	}

	otherUUID, err := uuid.Parse(otherUserID)
	if err != nil {
		return getMessagesResponse{}, codeerror.New(codeerror.InvalidUUID, "Other user ID is not a valid UUID")
	}

	_, err = s.repo.GetUserById(ctx, pgtype.UUID{Bytes: otherUUID, Valid: true})
	if err != nil {
		return getMessagesResponse{}, codeerror.New(codeerror.UserNotFound, "User not found")
	}

	rows, err := s.repo.GetMessagesBetweenUsers(ctx, repo.GetMessagesBetweenUsersParams{
		SenderID:   pgtype.UUID{Bytes: senderUUID, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: otherUUID, Valid: true},
		Limit:      int32(pageSize),
		Offset:     int32((page - 1) * pageSize),
	})
	if err != nil {
		return getMessagesResponse{}, err
	}

	messages := make([]MessageResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		messages = append(messages, MessageResponse{
			ID:         row.ID.String(),
			SenderID:   row.SenderID.String(),
			ReceiverID: row.ReceiverID.String(),
			Content:    row.Content,
			CreatedAt:  row.CreatedAt.Time,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return getMessagesResponse{
		Messages: messages,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
