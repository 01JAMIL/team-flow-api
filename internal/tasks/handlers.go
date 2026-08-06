package tasks

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

func NewTasksHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateTask(c *gin.Context) {
	projectID := c.Param("projectID")

	var payload createTaskPayload

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

	task, err := h.service.CreateTask(c, projectID, payload)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Project not found",
			})
			return
		}

		if errors.Is(err, ErrInvalidDate) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			})
			return
		}

		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_CREATE_TASK",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task":    task,
	})
}

func (h *handler) GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.service.GetTaskByID(c, taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_GET_TASK",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

func (h *handler) GetProjectTasks(c *gin.Context) {
	projectID := c.Param("projectID")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	response, err := h.service.GetProjectTasks(c, projectID, page, pageSize)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_GET_PROJECT_TASKS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
