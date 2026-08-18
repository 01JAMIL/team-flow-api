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
	getUserByIdFunc    func(ctx context.Context, id string) (repo.User, error)
	registerFunc       func(ctx context.Context, arg repo.RegisterParams) (repo.User, error)
}

func (f *mockAuthRepository) GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error) {
	panic("implement me")
}

func (f *mockAuthRepository) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return f.getUserByEmailFunc(ctx, email)
}

func (f *mockAuthRepository) Register(ctx context.Context, arg repo.RegisterParams) (repo.User, error) {
	return f.registerFunc(ctx, arg)
}

func TestRegister_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	expectedEmail := "johndoe@example.com"
	expectedUserID := uuid.New()

	expectedUser := repo.User{
		ID:        pgtype.UUID{Bytes: expectedUserID, Valid: true},
		FirstName: "John",
		LastName:  "Doe",
		Email:     expectedEmail,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	fakeRepo := &mockAuthRepository{
		getUserByEmailFunc: func(
			ctx context.Context,
			email string,
		) (repo.User, error) {
			return repo.User{}, pgx.ErrNoRows
		},

		registerFunc: func(
			ctx context.Context,
			arg repo.RegisterParams,
		) (repo.User, error) {
			return repo.User{
				ID:        pgtype.UUID{Bytes: expectedUserID, Valid: true},
				FirstName: arg.FirstName,
				LastName:  arg.LastName,
				Email:     arg.Email,
			}, nil
		},
	}

	service := &svc{
		repo: fakeRepo,
	}

	payload := registerPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Password:  "password123",
	}

	// Act
	result, err := service.Register(ctx, payload)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.User.ID != expectedUser.ID.String() {
		t.Errorf(
			"expected user ID %s, got %s",
			expectedUser.ID.String(),
			result.User.ID,
		)
	}

	if result.User.FirstName == "" || result.User.FirstName != expectedUser.FirstName {
		t.Errorf(
			"expected first name %s, got %s",
			expectedUser.FirstName,
			result.User.FirstName,
		)
	}

	if result.User.LastName == "" || result.User.LastName != expectedUser.LastName {
		t.Errorf(
			"expected last name %s, got %s",
			expectedUser.LastName,
			result.User.LastName,
		)
	}

	if result.User.Email == "" || result.User.Email != expectedUser.Email {
		t.Errorf(
			"expected email %s, got %s",
			expectedUser.Email,
			result.User.Email,
		)
	}

	if result.Token == "" {
		t.Error("expected token to be generated")
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	// Arrange
	ctx := context.Background()

	userID := uuid.New()

	fakeUser := repo.User{
		ID:        pgtype.UUID{Bytes: userID, Valid: true},
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	fakeRepo := &mockAuthRepository{
		// Simulate: user already exists
		getUserByEmailFunc: func(
			ctx context.Context,
			email string,
		) (repo.User, error) {
			return fakeUser, nil
		},

		// This should NEVER be called
		registerFunc: func(
			ctx context.Context,
			arg repo.RegisterParams,
		) (repo.User, error) {
			t.Fatal("register should not be called when user already exists")
			return repo.User{}, nil
		},
	}

	service := &svc{
		repo: fakeRepo,
	}

	payload := registerPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Password:  "password123",
	}

	// Act
	_, err := service.Register(ctx, payload)

	// Assert
	if err == nil {
		t.Fatal("expected error when user already exists")
	}
}
