package auth

import (
	"strings"

	codeerror "gin-api-1/internal/codeerror"

	"github.com/gin-gonic/gin"
)

func AuthenticationMiddleware(service Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := ""
		if authHeader != "" {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				codeerror.HandleError(c, codeerror.New(codeerror.InvalidToken, "Bearer token is invalid"))
				c.Abort()
				return
			}
		} else {
			// Browsers cannot set custom headers on WebSocket connections, so
			// WebSocket clients may authenticate via ?token=<JWT> instead.
			token = c.Query("token")
		}

		if token == "" {
			codeerror.HandleError(c, codeerror.New(codeerror.MissingToken, "Authorization header is missing"))
			c.Abort()
			return
		}

		claims, err := parseToken(token)
		if err != nil {
			codeerror.HandleError(c, codeerror.New(codeerror.InvalidToken, "Invalid token"))
			c.Abort()
			return
		}

		userID := claims.Subject
		user, err := service.GetUserById(c.Request.Context(), userID)

		if err != nil {
			codeerror.HandleError(c, codeerror.New(codeerror.StatusUnauthorized, "You are not authorized to access this resource"))
			c.Abort()
			return
		}

		c.Set("user", UserResponse{
			ID:        userID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})

		c.Next()
	}
}
