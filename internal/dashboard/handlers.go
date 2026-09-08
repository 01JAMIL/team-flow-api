package dashboard

import (
	"net/http"

	"gin-api-1/internal/auth"
	codeerror "gin-api-1/internal/codeerror"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewDashboardHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetKPIs(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	kpis, err := h.service.GetKPIs(c, loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"kpis": kpis,
	})
}
