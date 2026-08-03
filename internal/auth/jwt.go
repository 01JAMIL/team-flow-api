package auth

import (
	"errors"
	"fmt"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/env"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret string = env.GetEnvString("JWT_SECRET", "secret_123456")

func createToken(user repo.User) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(jwtSecret))
}

func parseToken(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtClaims{},
		func(token *jwt.Token) (any, error) {
			// Ensure we're using the expected signing algorithm
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
