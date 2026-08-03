package main

import (
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"gin-api-1/internal/workspace"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	r := gin.Default()

	authService := auth.NewAuthService(repo.New(app.db), app.db)
	authHandler := auth.NewAuthHandler(authService)

	workspaceService := workspace.NewWorkspaceService(repo.New(app.db), app.db)
	workspaceHandler := workspace.NewWorkspaceHandler(workspaceService)

	/* Public routes */
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

		/* Workspaces routes */
		authGroup.GET("/workspaces", workspaceHandler.GetUserWorkspaces)
		authGroup.GET("/workspaces/:id", workspaceHandler.GetUserWorkspaceByID)
		authGroup.POST("/workspaces", workspaceHandler.CreateWorkspace)
		authGroup.PATCH("/workspaces/:id", workspaceHandler.UpdateWorkspace)
		authGroup.DELETE("/workspaces/:id", workspaceHandler.DeleteWorkspace)
	}

	return r
}
