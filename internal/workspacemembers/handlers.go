package workspacemembers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

	member, err := h.service.AddWorkspaceMember(c, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_CREATE_WORKSPACE_MEMBER",
			"message": err.Error(),
		})
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
	if pageSize > 100 {
		pageSize = 100
	}

	response, err := h.service.GetWorkspaceMembers(c, id, page, pageSize)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Workspace not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FAILED_TO_GET_WORKSPACE_MEMBERS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
