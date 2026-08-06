package tasks

import (
	codeerror "gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	task, err := h.service.CreateTask(c, projectID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
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
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

func (h *handler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")

	var payload updateTaskPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	task, err := h.service.UpdateTask(c, taskID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task":    task,
	})
}

func (h *handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.service.DeleteTask(c, taskID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
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
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
