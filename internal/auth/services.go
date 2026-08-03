package auth

import (
	"context"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	GetUserByEmail(ctx context.Context, email string) (repo.User, error)
	GetUserById(ctx context.Context, id string) (repo.User, error)
	Register(ctx context.Context, payload registerPayload) (authResponse, error)
	Login(ctx context.Context, payload loginPayload) (authResponse, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewAuthService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *svc) GetUserById(ctx context.Context, id string) (repo.User, error) {
	return s.repo.GetUserById(ctx, id)
}

func (s *svc) Register(ctx context.Context, payload registerPayload) (authResponse, error) {

	_, err := s.GetUserByEmail(ctx, payload.Email)
	if err == nil {
		return authResponse{}, errors.New("user with this email already exists")
	}

	pk := uuid.New().String()

	hashedPassword, err := hashPassword(payload.Password)
	if err != nil {
		return authResponse{}, err
	}

	user, err := s.repo.Register(ctx, repo.RegisterParams{
		ID:        pk,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})

	if err != nil {
		return authResponse{}, err
	}

	jwt, err := createToken(user)

	if err != nil {
		return authResponse{}, err
	}

	return authResponse{
		User: UserResponse{
			ID:        user.ID,
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
		return authResponse{}, err
	}

	isValidPassword := checkPasswordHash(payload.Password, user.Password)
	if !isValidPassword {
		return authResponse{}, errors.New("invalid password")
	}

	jwt, err := createToken(user)
	if err != nil {
		return authResponse{}, err
	}

	return authResponse{
		User: UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token: jwt,
	}, nil
}
