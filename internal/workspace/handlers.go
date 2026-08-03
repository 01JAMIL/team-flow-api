package workspace

import (
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type handler struct {
	service Service
}

func NewWorkspaceHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetUserWorkspaceByID(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	workspace, err := h.service.GetUserWorkspaceByID(c, repo.GetUserWorkspaceByIDParams{
		UserID: pgtype.Text{
			String: loggedUser.ID,
			Valid:  true,
		},
		ID: id,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Workspace not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspace": workspace,
	})
}

func (h *handler) GetUserWorkspaces(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)
	workspaces, err := h.service.GetUserWorkspaces(c, pgtype.Text{
		String: loggedUser.ID,
		Valid:  true,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if workspaces == nil {
		workspaces = []repo.Workspace{}
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
	})
}
