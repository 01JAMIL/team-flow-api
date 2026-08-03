package main

import (
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	r := gin.Default()

	authService := auth.NewAuthService(repo.New(app.db), app.db)
	authHandler := auth.NewAuthHandler(authService)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		})

		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
	}

	authGroup := v1.Group("/")
	authGroup.Use(auth.AuthenticationMiddleware(authService))
	{
		authGroup.GET("/auth/me", authHandler.GetMe)
	}

	return r
}
