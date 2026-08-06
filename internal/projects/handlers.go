package projects

import (
	codeerror "gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		codeerror.HandleError(c, err)
		return
	}

	var payload createProjectPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	project, err := h.service.CreateProject(c, workspaceID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
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
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *handler) GetProjectByID(c *gin.Context) {
	projectID := c.Param("projectID")

	project, err := h.service.GetProjectByID(c, projectID)
	if err != nil {
		codeerror.HandleError(c, err)
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
		codeerror.HandleError(c, err)
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
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	project, err := h.service.UpdateProject(c, projectID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"project": project,
	})
}
