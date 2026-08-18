package auth

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockAuthRepository struct {
	getUserByEmailFunc func(ctx context.Context, email string) (repo.User, error)
	getUserByIdFunc    func(ctx context.Context, id pgtype.UUID) (repo.User, error)
	registerFunc       func(ctx context.Context, arg repo.RegisterParams) (repo.User, error)
}

func (f *mockAuthRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	return f.getUserByIdFunc(ctx, id)
}

func (f *mockAuthRepository) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return f.getUserByEmailFunc(ctx, email)
}

func (f *mockAuthRepository) Register(ctx context.Context, arg repo.RegisterParams) (repo.User, error) {
	return f.registerFunc(ctx, arg)
}

func newTestUser(userID uuid.UUID, email string) repo.User {
	return repo.User{
		ID:        pgtype.UUID{Bytes: userID, Valid: true},
		FirstName: "John",
		LastName:  "Doe",
		Email:     email,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		payload     registerPayload
		repo        *mockAuthRepository
		wantErr     bool
		wantToken   bool
		wantIDMatch bool
	}{
		{
			name: "success",
			payload: registerPayload{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "johndoe@example.com",
				Password:  "password123",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
				registerFunc: func(ctx context.Context, arg repo.RegisterParams) (repo.User, error) {
					return newTestUser(uuid.New(), arg.Email), nil
				},
			},
			wantErr:     false,
			wantToken:   true,
			wantIDMatch: true,
		},
		{
			name: "user already exists",
			payload: registerPayload{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "johndoe@example.com",
				Password:  "password123",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					return newTestUser(uuid.New(), email), nil
				},
				registerFunc: func(ctx context.Context, arg repo.RegisterParams) (repo.User, error) {
					t.Fatal("register should not be called when user already exists")
					return repo.User{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "repo register failure",
			payload: registerPayload{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "johndoe@example.com",
				Password:  "password123",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
				registerFunc: func(ctx context.Context, arg repo.RegisterParams) (repo.User, error) {
					return repo.User{}, pgx.ErrTxClosed
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.Register(ctx, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantToken && result.Token == "" {
				t.Error("expected token to be generated")
			}

			if result.User.Email != tt.payload.Email {
				t.Errorf("email = %s, want %s", result.User.Email, tt.payload.Email)
			}

			if result.User.FirstName != tt.payload.FirstName {
				t.Errorf("first name = %s, want %s", result.User.FirstName, tt.payload.FirstName)
			}

			if result.User.LastName != tt.payload.LastName {
				t.Errorf("last name = %s, want %s", result.User.LastName, tt.payload.LastName)
			}
		})
	}
}

func TestGetUserById(t *testing.T) {
	ctx := context.Background()
	validID := uuid.New()

	tests := []struct {
		name    string
		id      string
		repo    *mockAuthRepository
		wantErr bool
	}{
		{
			name: "success",
			id:   validID.String(),
			repo: &mockAuthRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return newTestUser(validID, "john@example.com"), nil
				},
			},
			wantErr: false,
		},
		{
			name: "invalid uuid",
			id:   "not-a-uuid",
			repo: &mockAuthRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "user not found",
			id:   validID.String(),
			repo: &mockAuthRepository{
				getUserByIdFunc: func(ctx context.Context, id pgtype.UUID) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.GetUserById(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ID.String() != validID.String() {
				t.Errorf("user ID = %s, want %s", result.ID, validID)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	hashedPassword, err := hashPassword("correctpassword")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	tests := []struct {
		name      string
		payload   loginPayload
		repo      *mockAuthRepository
		wantErr   bool
		wantToken bool
	}{
		{
			name: "success",
			payload: loginPayload{
				Email:    "john@example.com",
				Password: "correctpassword",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					user := newTestUser(userID, email)
					user.Password = hashedPassword
					return user, nil
				},
			},
			wantErr:   false,
			wantToken: true,
		},
		{
			name: "user not found",
			payload: loginPayload{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					return repo.User{}, pgx.ErrNoRows
				},
			},
			wantErr: true,
		},
		{
			name: "wrong password",
			payload: loginPayload{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			repo: &mockAuthRepository{
				getUserByEmailFunc: func(ctx context.Context, email string) (repo.User, error) {
					user := newTestUser(userID, email)
					user.Password = hashedPassword
					return user, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &svc{repo: tt.repo}
			result, err := service.Login(ctx, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantToken && result.Token == "" {
				t.Error("expected token to be generated")
			}

			if result.User.Email != tt.payload.Email {
				t.Errorf("email = %s, want %s", result.User.Email, tt.payload.Email)
			}
		})
	}
}
