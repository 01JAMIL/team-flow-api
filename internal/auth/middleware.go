package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthenticationMiddleware(service Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is empty",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token is invalid",
			})
			c.Abort()
			return
		}

		claims, err := parseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		userID := claims.Subject
		user, err := service.GetUserById(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to access this resource",
			})
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
