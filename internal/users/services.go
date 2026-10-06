package users

import (
	"context"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"gin-api-1/internal/codeerror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	GetUsers(ctx context.Context, excludedUserID, search string, page, pageSize int) (getUsersResponse, error)
	UpdateUserProfile(ctx context.Context, userID string, payload updateUserProfilePayload) (auth.UserResponse, error)
}

type usersRepository interface {
	GetUsers(ctx context.Context, arg repo.GetUsersParams) ([]repo.GetUsersRow, error)
	UpdateUserProfile(ctx context.Context, arg repo.UpdateUserProfileParams) (repo.User, error)
}

type svc struct {
	repo usersRepository
	db   *pgxpool.Pool
}

func NewUsersService(repo usersRepository, db *pgxpool.Pool) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) GetUsers(ctx context.Context, excludedUserID, search string, page, pageSize int) (getUsersResponse, error) {
	excludedUUID, err := uuid.Parse(excludedUserID)
	if err != nil {
		return getUsersResponse{}, codeerror.New(codeerror.InvalidUUID, "User ID is not a valid UUID")
	}

	rows, err := s.repo.GetUsers(ctx, repo.GetUsersParams{
		ExcludedUserID: pgtype.UUID{Bytes: excludedUUID, Valid: true},
		Search:         pgtype.Text{String: search, Valid: true},
		PageOffset:     int32((page - 1) * pageSize),
		PageLimit:      int32(pageSize),
	})
	if err != nil {
		return getUsersResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to fetch users", err)
	}

	users := make([]auth.UserResponse, 0, len(rows))

	var total int64
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	for _, row := range rows {
		users = append(users, auth.UserResponse{
			ID:        row.ID.String(),
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Email:     row.Email,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return getUsersResponse{
		Users: users,
		Pagination: paginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		},
	}, nil
}

func (s *svc) UpdateUserProfile(ctx context.Context, userID string, payload updateUserProfilePayload) (auth.UserResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return auth.UserResponse{}, codeerror.New(codeerror.InvalidUUID, "User ID is not a valid UUID")
	}

	user, err := s.repo.UpdateUserProfile(ctx, repo.UpdateUserProfileParams{
		ID:        pgtype.UUID{Bytes: userUUID, Valid: true},
		FirstName: textPtr(payload.FirstName),
		LastName:  textPtr(payload.LastName),
	})
	if err != nil {
		return auth.UserResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to update user profile", err)
	}

	return toUserResponse(user), nil
}

func textPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{String: *value, Valid: true}
}

func toUserResponse(user repo.User) auth.UserResponse {
	return auth.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
