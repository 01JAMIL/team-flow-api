package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

	_, err := h.service.GetUserByEmail(c, payload.Email)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "USER_ALREADY_EXISTS",
			"message": "User with this email already exists",
		})
		return
	}

	user, err := h.service.Register(c, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "USER_CREATION_FAILED",
			"message": "Failed to register user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}
