package auth

import (
	"fmt"
	codeerror "gin-api-1/internal/codeerror"
	"gin-api-1/internal/email"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service      Service
	emailService email.Service
}

func NewAuthHandler(service Service, emailService email.Service) *handler {
	return &handler{
		service:      service,
		emailService: emailService,
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

	err = h.emailService.SendRegisterWelcomeEmail(user.User.Email)
	if err != nil {
		fmt.Println(err)
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
