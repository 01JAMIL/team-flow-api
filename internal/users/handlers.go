package users

import (
	"gin-api-1/internal/auth"
	"gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const DefaultPageSize = 10

type handler struct {
	service Service
}

func NewUsersHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetUsers(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	search := c.Query("search")

	response, err := h.service.GetUsers(c, loggedUser.ID, search, page, pageSize)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
