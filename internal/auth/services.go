package auth

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	GetUserByEmail(ctx context.Context, email string) (repo.User, error)
	GetUserById(ctx context.Context, id string) (repo.User, error)
	Register(ctx context.Context, payload registerPayload) (repo.User, error)
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

func (s *svc) Register(ctx context.Context, payload registerPayload) (repo.User, error) {

	pk := uuid.New().String()

	hashedPassword, err := hashPassword(payload.Password)
	if err != nil {
		return repo.User{}, err
	}

	return s.repo.Register(ctx, repo.RegisterParams{
		ID:        pk,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})
}
