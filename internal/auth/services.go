package auth

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	GetUserByEmail(ctx context.Context, email string) (repo.User, error)
	GetUserById(ctx context.Context, id string) (repo.User, error)
	Register(ctx context.Context, payload registerPayload) (authResponse, error)
	Login(ctx context.Context, payload loginPayload) (authResponse, error)
}

// Interface for the database dependency.
type authRepository interface {
	GetUserByEmail(ctx context.Context, email string) (repo.User, error)
	GetUserById(ctx context.Context, id pgtype.UUID) (repo.User, error)
	Register(ctx context.Context, arg repo.RegisterParams) (repo.User, error)
}

type svc struct {
	repo authRepository
	db   *pgx.Conn
}

func NewAuthService(repo authRepository, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *svc) GetUserById(ctx context.Context, id string) (repo.User, error) {
	value, err := uuid.Parse(id)
	if err != nil {
		return repo.User{}, codeerror.New(codeerror.UserNotFound, "User not found")
	}
	return s.repo.GetUserById(ctx, pgtype.UUID{
		Bytes: value,
		Valid: true,
	})
}

func (s *svc) Register(ctx context.Context, payload registerPayload) (authResponse, error) {

	_, err := s.GetUserByEmail(ctx, payload.Email)
	if err == nil {
		return authResponse{}, codeerror.New(codeerror.UserAlreadyExist, "User with this email already exists")
	}

	pk := uuid.New()

	hashedPassword, err := hashPassword(payload.Password)
	if err != nil {
		return authResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create user", err)
	}

	user, err := s.repo.Register(ctx, repo.RegisterParams{
		ID: pgtype.UUID{
			Bytes: pk,
			Valid: true,
		},
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})

	if err != nil {
		return authResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create user", err)
	}

	jwt, err := createToken(user)

	if err != nil {
		return authResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create session token", err)
	}

	return authResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token: jwt,
	}, nil
}

func (s *svc) Login(ctx context.Context, payload loginPayload) (authResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return authResponse{}, codeerror.New(codeerror.InvalidCredentials, "Invalid email or password")
	}

	isValidPassword := checkPasswordHash(payload.Password, user.Password)
	if !isValidPassword {
		return authResponse{}, codeerror.New(codeerror.InvalidCredentials, "Invalid email or password")
	}

	jwt, err := createToken(user)
	if err != nil {
		return authResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to create session token", err)
	}

	return authResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token: jwt,
	}, nil
}
