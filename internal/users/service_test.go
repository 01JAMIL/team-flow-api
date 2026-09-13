package users

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockUsersRepository struct {
	getUsersFunc func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error)
}

func (m *mockUsersRepository) GetUsers(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
	return m.getUsersFunc(ctx, arg)
}

func TestGetUsers(t *testing.T) {
	ctx := context.Background()
	loggedInUserID := uuid.New()
	otherUserID := uuid.New()

	otherUser := repo.GetUsersRow{
		TotalCount: 1,
		ID:         pgtype.UUID{Bytes: otherUserID, Valid: true},
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane@example.com",
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	tests := []struct {
		name    string
		id      string
		repo    *mockUsersRepository
		wantErr bool
		wantLen int
	}{
		{
			name:    "success",
			id:      loggedInUserID.String(),
			wantLen: 1,
			repo: &mockUsersRepository{
				getUsersFunc: func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
					if arg.ExcludedUserID.Bytes != loggedInUserID {
						t.Errorf("excluded user ID = %v, want %v", arg.ExcludedUserID.Bytes, loggedInUserID)
					}
					return []repo.GetUsersRow{otherUser}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "invalid uuid",
			id:   "not-a-uuid",
			repo: &mockUsersRepository{
				getUsersFunc: func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
					t.Fatal("repo should not be called for an invalid UUID")
					return nil, nil
				},
			},
			wantErr: true,
		},
		{
			name: "empty result",
			id:   loggedInUserID.String(),
			repo: &mockUsersRepository{
				getUsersFunc: func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
					return []repo.GetUsersRow{}, nil
				},
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "repo failure",
			id:   loggedInUserID.String(),
			repo: &mockUsersRepository{
				getUsersFunc: func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
					return nil, errors.New("repo failure")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetUsers(ctx, tt.id, "", 1, 10)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Users) != tt.wantLen {
				t.Errorf("users count = %d, want %d", len(result.Users), tt.wantLen)
			}

			if tt.wantLen > 0 {
				if result.Users[0].ID != otherUserID.String() {
					t.Errorf("user ID = %s, want %s", result.Users[0].ID, otherUserID)
				}
				if result.Pagination.Total != 1 || result.Pagination.TotalPages != 1 {
					t.Errorf("pagination = %+v, want total=1 totalPages=1", result.Pagination)
				}
			}
		})
	}
}
