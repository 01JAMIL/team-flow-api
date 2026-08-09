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
