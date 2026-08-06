package workspacemembers

import (
	codeerror "gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewWorkspaceMembersHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) AddWorkspaceMember(c *gin.Context) {
	id := c.Param("id")

	var payload addWorkspaceMemberPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	member, err := h.service.AddWorkspaceMember(c, id, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added successfully",
		"member":  member,
	})
}

func (h *handler) GetWorkspaceMembers(c *gin.Context) {
	id := c.Param("id")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	response, err := h.service.GetWorkspaceMembers(c, id, page, pageSize)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *handler) RemoveWorkspaceMember(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("userId")

	err := h.service.RemoveWorkspaceMember(c, id, userID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed successfully",
	})
}
