package main

import (
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"gin-api-1/internal/email"
	"gin-api-1/internal/messages"
	"gin-api-1/internal/payment"
	"gin-api-1/internal/projects"
	"gin-api-1/internal/subscriptions"
	"gin-api-1/internal/tasks"
	"gin-api-1/internal/websocket"
	"gin-api-1/internal/workspace"
	"gin-api-1/internal/workspacemembers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	r := gin.Default()

	emailService := email.NewEmailService(app.resend)
	paymentService := payment.NewStripeService(app.stripe, repo.New(app.db))
	subscriptionsService := subscriptions.NewSubscriptionsService(repo.New(app.db), app.db, app.stripe)
	paymentHandler := payment.NewPaymentHandler(*paymentService, subscriptionsService)

	authService := auth.NewAuthService(repo.New(app.db), app.db)
	authHandler := auth.NewAuthHandler(authService, *emailService)

	workspaceService := workspace.NewWorkspaceService(repo.New(app.db), app.db, *paymentService)
	workspaceHandler := workspace.NewWorkspaceHandler(workspaceService)

	workspaceMembersService := workspacemembers.NewWorkspaceMembersService(repo.New(app.db), app.db)
	workspaceMembersHandler := workspacemembers.NewWorkspaceMembersHandler(workspaceMembersService)

	projectsService := projects.NewProjectsService(repo.New(app.db), app.db)
	projectsHandler := projects.NewProjectsHandler(projectsService)

	tasksService := tasks.NewTasksService(repo.New(app.db), app.db)
	tasksHandler := tasks.NewTasksHandler(tasksService)

	messageService := messages.NewMessagesService(repo.New(app.db), app.db)
	messagesHandler := messages.NewMessagesHandler(messageService)

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

		// Webhooks
		v1.POST("/webhooks/stripe", paymentHandler.HandleWebhook)
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
		authGroup.POST("/workspaces/:id/checkout", workspaceHandler.CreateCheckoutSession)

		/* Workspace Members routes */
		authGroup.GET("/workspaces/:id/members", workspaceMembersHandler.GetWorkspaceMembers)
		authGroup.POST("/workspaces/:id/members", workspaceMembersHandler.AddWorkspaceMember)
		authGroup.DELETE("/workspaces/:id/members/:userId", workspaceMembersHandler.RemoveWorkspaceMember)

		/* Projects routes */
		authGroup.GET("/workspaces/:id/projects", projectsHandler.GetWorkspaceProjects)
		authGroup.POST("/workspaces/:id/projects", projectsHandler.CreateProject)
		authGroup.GET("/projects/:projectID", projectsHandler.GetProjectByID)
		authGroup.PATCH("/projects/:projectID", projectsHandler.UpdateProject)
		authGroup.DELETE("/projects/:projectID", projectsHandler.DeleteProject)

		/* Tasks routes */
		authGroup.POST("/projects/:projectID/tasks", tasksHandler.CreateTask)
		authGroup.GET("/projects/:projectID/tasks", tasksHandler.GetProjectTasks)
		authGroup.GET("/tasks/:id", tasksHandler.GetTaskByID)
		authGroup.PATCH("/tasks/:id", tasksHandler.UpdateTask)
		authGroup.DELETE("/tasks/:id", tasksHandler.DeleteTask)

		/* Messages routes */
		authGroup.GET("/messages/:userId", messagesHandler.GetMessagesBetweenUsers)

		// WebSocket Connection
		authGroup.GET("/ws", func(c *gin.Context) {
			websocket.Handler(c, messageService)
		})
	}

	return r
}
