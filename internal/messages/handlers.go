package messages

import (
	"gin-api-1/internal/auth"
	"gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewMessagesHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateMessage(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload CreateMessagePayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	message, err := h.service.CreateMessage(c, loggedUser.ID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message created successfully",
		"data":    message,
	})

}

func (h *handler) GetMessagesBetweenUsers(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)
	otherUserID := c.Param("userId")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	response, err := h.service.GetMessagesBetweenUsers(c, loggedUser.ID, otherUserID, page, pageSize)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
