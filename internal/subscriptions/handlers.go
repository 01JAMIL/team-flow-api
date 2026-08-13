package subscriptions

import (
	"gin-api-1/internal/auth"
	"gin-api-1/internal/codeerror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewSubscriptionsHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetWorkspaceSubscription(c *gin.Context) {
	workspaceID := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	subscription, err := h.service.GetWorkspaceSubscription(c, workspaceID, loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}
