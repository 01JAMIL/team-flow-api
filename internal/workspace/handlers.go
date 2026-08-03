package workspace

import (
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func (h *handler) CreateWorkspace(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload createWorkspacePayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			errs := make(map[string]string)

			for _, fieldErr := range validationErrors {
				errs[fieldErr.Field()] = fieldErr.Error()
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Validation failed",
				"errors":  errs,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"errors": err.Error(),
		})
		return
	}

	workspace, err := h.service.CreateWorkspace(c, createWorkspacePayload{
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
		UserID:        loggedUser.ID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_CREATE_WORKSPACE",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Workspace created successfully",
		"workspace": workspace,
	})
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

func (h *handler) UpdateWorkspace(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload updateWorkspacePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	workspace, err := h.service.UpdateWorkspace(c, updateWorkspacePayload{
		ID:            id,
		UserID:        loggedUser.ID,
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update workspace",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Workspace updated successfully",
		"workspace": workspace,
	})
}

func (h *handler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)
	err := h.service.DeleteWorkspace(c, repo.DeleteWorkspaceParams{
		ID: id,
		UserID: pgtype.Text{
			String: loggedUser.ID,
			Valid:  true,
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete workspace",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace deleted successfully",
	})
}
