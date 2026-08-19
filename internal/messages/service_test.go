package messages

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockMessagesRepository struct {
	getUserByIdFunc             func(ctx context.Context, id pgtype.UUID) (repo.User, error)
	createMessageFunc           func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error)
	getMessagesBetweenUsersFunc func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error)
}

func (m *mockMessagesRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	return m.getUserByIdFunc(ctx, id)
}

func (m *mockMessagesRepository) CreateMessage(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
	return m.createMessageFunc(ctx, arg)
}

func (m *mockMessagesRepository) GetMessagesBetweenUsers(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
	return m.getMessagesBetweenUsersFunc(ctx, arg)
}

func newTestMessage(id, senderID, receiverID uuid.UUID, content string) repo.Message {
	return repo.Message{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		SenderID:   pgtype.UUID{Bytes: senderID, Valid: true},
		ReceiverID: pgtype.UUID{Bytes: receiverID, Valid: true},
		Content:    content,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestCreateMessage(t *testing.T) {
	ctx := context.Background()
	senderID := uuid.New()
	receiverID := uuid.New()

	tests := []struct {
		name    string
		sender  string
		payload CreateMessagePayload
		repo    *mockMessagesRepository
		wantErr bool
	}{
		{
			name:   "success",
			sender: senderID.String(),
			payload: CreateMessagePayload{
				ReceiverID: receiverID.String(),
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					return newTestMessage(arg.ID.Bytes, arg.SenderID.Bytes, arg.ReceiverID.Bytes, arg.Content), nil
				},
			},
			wantErr: false,
		},
		{
			name:   "invalid sender uuid",
			sender: "not-a-uuid",
			payload: CreateMessagePayload{
				ReceiverID: receiverID.String(),
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					t.Fatal("GetUserById should not be called")
					return repo.User{}, nil
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					t.Fatal("CreateMessage should not be called")
					return repo.Message{}, nil
				},
			},
			wantErr: true,
		},
		{
			name:   "invalid receiver uuid",
			sender: senderID.String(),
			payload: CreateMessagePayload{
				ReceiverID: "not-a-uuid",
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					t.Fatal("GetUserById should not be called")
					return repo.User{}, nil
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					t.Fatal("CreateMessage should not be called")
					return repo.Message{}, nil
				},
			},
			wantErr: true,
		},
		{
			name:   "self message not allowed",
			sender: senderID.String(),
			payload: CreateMessagePayload{
				ReceiverID: senderID.String(),
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					t.Fatal("GetUserById should not be called")
					return repo.User{}, nil
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					t.Fatal("CreateMessage should not be called")
					return repo.Message{}, nil
				},
			},
			wantErr: true,
		},
		{
			name:   "receiver not found",
			sender: senderID.String(),
			payload: CreateMessagePayload{
				ReceiverID: receiverID.String(),
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					t.Fatal("CreateMessage should not be called")
					return repo.Message{}, nil
				},
			},
			wantErr: true,
		},
		{
			name:   "repo create message failure",
			sender: senderID.String(),
			payload: CreateMessagePayload{
				ReceiverID: receiverID.String(),
				Content:    "Hello!",
			},
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				createMessageFunc: func(ctx context.Context, arg repo.CreateMessageParams) (repo.Message, error) {
					return repo.Message{}, pgx.ErrTxClosed
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.CreateMessage(ctx, tt.sender, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Content != tt.payload.Content {
				t.Errorf("content = %s, want %s", result.Content, tt.payload.Content)
			}

			if result.SenderID.String() != senderID.String() {
				t.Errorf("sender ID = %s, want %s", result.SenderID, senderID)
			}

			if result.ReceiverID.String() != receiverID.String() {
				t.Errorf("receiver ID = %s, want %s", result.ReceiverID, receiverID)
			}
		})
	}
}

func TestGetMessagesBetweenUsers(t *testing.T) {
	ctx := context.Background()
	senderID := uuid.New()
	receiverID := uuid.New()

	tests := []struct {
		name           string
		loggedUserID   string
		otherUserID    string
		page           int
		pageSize       int
		repo           *mockMessagesRepository
		wantErr        bool
		wantMessages   int
		wantTotalPages int
	}{
		{
			name:         "success",
			loggedUserID: senderID.String(),
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					return []repo.GetMessagesBetweenUsersRow{
						{
							TotalCount: 1,
							ID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
							SenderID:   arg.SenderID,
							ReceiverID: arg.ReceiverID,
							Content:    "Hello!",
							CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
						},
					}, nil
				},
			},
			wantErr:        false,
			wantMessages:   1,
			wantTotalPages: 1,
		},
		{
			name:         "invalid logged user uuid",
			loggedUserID: "not-a-uuid",
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					t.Fatal("GetUserById should not be called")
					return repo.User{}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					t.Fatal("GetMessagesBetweenUsers should not be called")
					return nil, nil
				},
			},
			wantErr: true,
		},
		{
			name:         "invalid other user uuid",
			loggedUserID: senderID.String(),
			otherUserID:  "not-a-uuid",
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					t.Fatal("GetUserById should not be called")
					return repo.User{}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					t.Fatal("GetMessagesBetweenUsers should not be called")
					return nil, nil
				},
			},
			wantErr: true,
		},
		{
			name:         "other user not found",
			loggedUserID: senderID.String(),
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					t.Fatal("GetMessagesBetweenUsers should not be called")
					return nil, nil
				},
			},
			wantErr: true,
		},
		{
			name:         "empty result",
			loggedUserID: senderID.String(),
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					return []repo.GetMessagesBetweenUsersRow{}, nil
				},
			},
			wantErr:        false,
			wantMessages:   0,
			wantTotalPages: 0,
		},
		{
			name:         "pagination calculates total pages correctly",
			loggedUserID: senderID.String(),
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					return []repo.GetMessagesBetweenUsersRow{
						{
							TotalCount: 25,
							ID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
							SenderID:   arg.SenderID,
							ReceiverID: arg.ReceiverID,
							Content:    "Message 1",
							CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
						},
					}, nil
				},
			},
			wantErr:        false,
			wantMessages:   1,
			wantTotalPages: 3,
		},
		{
			name:         "repo error",
			loggedUserID: senderID.String(),
			otherUserID:  receiverID.String(),
			page:         1,
			pageSize:     10,
			repo: &mockMessagesRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{ID: id}, nil
				},
				getMessagesBetweenUsersFunc: func(ctx context.Context, arg repo.GetMessagesBetweenUsersParams) ([]repo.GetMessagesBetweenUsersRow, error) {
					return nil, pgx.ErrTxClosed
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetMessagesBetweenUsers(ctx, tt.loggedUserID, tt.otherUserID, tt.page, tt.pageSize)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Messages) != tt.wantMessages {
				t.Errorf("messages count = %d, want %d", len(result.Messages), tt.wantMessages)
			}

			if result.Pagination.Page != tt.page {
				t.Errorf("page = %d, want %d", result.Pagination.Page, tt.page)
			}

			if result.Pagination.PageSize != tt.pageSize {
				t.Errorf("page size = %d, want %d", result.Pagination.PageSize, tt.pageSize)
			}

			if result.Pagination.TotalPages != tt.wantTotalPages {
				t.Errorf("total pages = %d, want %d", result.Pagination.TotalPages, tt.wantTotalPages)
			}
		})
	}
}
