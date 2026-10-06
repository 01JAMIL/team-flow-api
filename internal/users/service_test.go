package users

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/codeerror"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockUsersRepository struct {
	getUsersFunc          func(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error)
	updateUserProfileFunc func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error)
}

func (m *mockUsersRepository) GetUsers(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error) {
	return m.getUsersFunc(ctx, arg)
}

func (m *mockUsersRepository) UpdateUserProfile(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
	return m.updateUserProfileFunc(ctx, arg)
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

func TestUpdateUserProfile(t *testing.T) {
	ctx := context.Background()
	loggedInUserID := uuid.New()

	str := func(value string) *string { return &value }

	existingUser := repo.User{
		ID:        pgtype.UUID{Bytes: loggedInUserID, Valid: true},
		FirstName: "Old",
		LastName:  "Name",
		Email:     "jane@example.com",
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	tests := []struct {
		name      string
		userID    string
		payload   updateUserProfilePayload
		repo      *mockUsersRepository
		wantErr   bool
		wantCode  string
		wantFirst string
		wantLast  string
	}{
		{
			name:    "success",
			userID:  loggedInUserID.String(),
			payload: updateUserProfilePayload{FirstName: str("Jane"), LastName: str("Doe")},
			repo: &mockUsersRepository{
				updateUserProfileFunc: func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
					if arg.ID.Bytes != loggedInUserID {
						t.Errorf("id = %v, want %v", arg.ID.Bytes, loggedInUserID)
					}
					if !arg.FirstName.Valid || arg.FirstName.String != "Jane" {
						t.Errorf("firstName = %+v, want Jane", arg.FirstName)
					}
					if !arg.LastName.Valid || arg.LastName.String != "Doe" {
						t.Errorf("lastName = %+v, want Doe", arg.LastName)
					}

					updated := existingUser
					updated.FirstName = arg.FirstName.String
					updated.LastName = arg.LastName.String
					return updated, nil
				},
			},
			wantFirst: "Jane",
			wantLast:  "Doe",
		},
		{
			name:    "partial update keeps omitted field",
			userID:  loggedInUserID.String(),
			payload: updateUserProfilePayload{FirstName: str("Jane")},
			repo: &mockUsersRepository{
				updateUserProfileFunc: func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
					if !arg.FirstName.Valid || arg.FirstName.String != "Jane" {
						t.Errorf("firstName = %+v, want Jane", arg.FirstName)
					}
					if arg.LastName.Valid {
						t.Errorf("lastName = %+v, want invalid (kept by COALESCE)", arg.LastName)
					}

					updated := existingUser
					updated.FirstName = arg.FirstName.String
					return updated, nil
				},
			},
			wantFirst: "Jane",
			wantLast:  "Name",
		},
		{
			name:    "empty payload is a no-op",
			userID:  loggedInUserID.String(),
			payload: updateUserProfilePayload{},
			repo: &mockUsersRepository{
				updateUserProfileFunc: func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
					if arg.FirstName.Valid || arg.LastName.Valid {
						t.Errorf("params = %+v, want both fields invalid", arg)
					}
					return existingUser, nil
				},
			},
			wantFirst: "Old",
			wantLast:  "Name",
		},
		{
			name:   "invalid user id",
			userID: "not-a-uuid",
			repo: &mockUsersRepository{
				updateUserProfileFunc: func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
					t.Fatal("repo should not be called for an invalid UUID")
					return repo.User{}, nil
				},
			},
			wantErr:  true,
			wantCode: codeerror.InvalidUUID,
		},
		{
			name:   "repo failure",
			userID: loggedInUserID.String(),
			repo: &mockUsersRepository{
				updateUserProfileFunc: func(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error) {
					return repo.User{}, errors.New("repo failure")
				},
			},
			wantErr:  true,
			wantCode: codeerror.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.UpdateUserProfile(ctx, tt.userID, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, codeerror.New(tt.wantCode, "")) {
					t.Errorf("error code = %v, want %s", err, tt.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ID != loggedInUserID.String() {
				t.Errorf("user ID = %s, want %s", result.ID, loggedInUserID)
			}
			if result.FirstName != tt.wantFirst {
				t.Errorf("firstName = %s, want %s", result.FirstName, tt.wantFirst)
			}
			if result.LastName != tt.wantLast {
				t.Errorf("lastName = %s, want %s", result.LastName, tt.wantLast)
			}
		})
	}
}
