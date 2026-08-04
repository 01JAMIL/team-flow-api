package workspace

import (
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	"net/http"
	"strconv"

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

	var pgUUID pgtype.UUID
	copy(pgUUID.Bytes[:], id[:])

	workspace, err := h.service.GetUserWorkspaceByID(c, repo.GetUserWorkspaceByIDParams{
		UserID: pgtype.Text{
			String: loggedUser.ID,
			Valid:  true,
		},
		ID: pgUUID,
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

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}

	response, err := h.service.GetUserWorkspaces(c, pgtype.Text{
		String: loggedUser.ID,
		Valid:  true,
	}, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
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

	var pgUUID pgtype.UUID
	copy(pgUUID.Bytes[:], id[:])

	err := h.service.DeleteWorkspace(c, repo.DeleteWorkspaceParams{
		ID: pgUUID,
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
