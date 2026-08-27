package integrations

import (
	"context"
	"gin-api-1/internal/codeerror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewIntegrationsHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateIntegrationTask(c *gin.Context) {
	var payload createIntegrationTaskParams

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	integrationTask, err := h.service.CreateIntegrationTask(context.Background(), payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "integration task created",
		"integrationTask": integrationTask,
	})
}
