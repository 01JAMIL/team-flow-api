package projects

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type handler struct {
	service Service
}

func NewProjectsHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateProject(c *gin.Context) {
	workspaceID := c.Param("id")

	if err := h.service.WorkspaceExists(c, workspaceID); err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_CHECK_WORKSPACE",
			"message": err.Error(),
		})
		return
	}

	var payload createProjectPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Request body is required",
			})
			return
		}

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
			"code":    "VALIDATION_ERROR",
			"message": "Invalid request body",
		})
		return
	}

	project, err := h.service.CreateProject(c, workspaceID, payload)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_CREATE_PROJECT",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": project,
	})
}
