package projects

import (
	"errors"
	"io"
	"net/http"
	"strconv"

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

func (h *handler) GetWorkspaceProjects(c *gin.Context) {
	workspaceID := c.Param("id")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	response, err := h.service.GetWorkspaceProjects(c, workspaceID, page, pageSize)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_GET_WORKSPACE_PROJECTS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *handler) GetProjectByID(c *gin.Context) {
	projectID := c.Param("projectID")

	project, err := h.service.GetProjectByID(c, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_GET_PROJECT",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

func (h *handler) DeleteProject(c *gin.Context) {
	projectID := c.Param("projectID")

	err := h.service.DeleteProject(c, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_DELETE_PROJECT",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

func (h *handler) UpdateProject(c *gin.Context) {
	projectID := c.Param("projectID")

	var payload updateProjectPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	project, err := h.service.UpdateProject(c, projectID, payload)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_UPDATE_PROJECT",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"project": project,
	})
}
