package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type jwtClaims struct {
	jwt.RegisteredClaims
}
type registerPayload struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

type loginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
	ID        string             `json:"id"`
	FirstName string             `json:"firstName"`
	LastName  string             `json:"lastName"`
	Email     string             `json:"email"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt pgtype.Timestamptz `json:"updatedAt"`
}
type authResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}
