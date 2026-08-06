package auth

import (
	codeerror "gin-api-1/internal/codeerror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewAuthHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Register(c *gin.Context) {
	var payload registerPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	user, err := h.service.Register(c, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *handler) Login(c *gin.Context) {
	var payload loginPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	user, err := h.service.Login(c, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged successfully",
		"user":    user,
	})
}

func (h *handler) GetMe(c *gin.Context) {
	user := c.MustGet("user").(UserResponse)
	c.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"user":    user,
	})
}
