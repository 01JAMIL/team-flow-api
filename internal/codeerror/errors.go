// Package codeerror defines the application-level error codes and error type
// used across the API so every error response has a consistent shape.
//
//	{
//	  "code":    "PROJECT_NOT_FOUND",
//	  "message": "Project not found"
//	}
package codeerror

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Error codes returned in the "code" field of every error response.
// Frontends can switch on these values to handle errors per resource.
const (
	// Generic HTTP-aligned codes.
	StatusBadRequest          = "BAD_REQUEST"
	StatusUnauthorized        = "UNAUTHORIZED"
	StatusForbidden           = "FORBIDDEN"
	StatusNotFound            = "NOT_FOUND"
	StatusConflict            = "CONFLICT"
	StatusInternalServerError = "INTERNAL_SERVER_ERROR"

	// Auth codes.
	ValidationError    = "VALIDATION_ERROR"
	UserAlreadyExist   = "USER_ALREADY_EXIST"
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidToken       = "INVALID_TOKEN"
	MissingToken       = "MISSING_TOKEN"

	// Domain codes.
	WorkspaceNotFound   = "WORKSPACE_NOT_FOUND"
	ProjectNotFound     = "PROJECT_NOT_FOUND"
	TaskNotFound        = "TASK_NOT_FOUND"
	UserNotFound        = "USER_NOT_FOUND"
	MemberAlreadyExists = "MEMBER_ALREADY_EXISTS"
	InvalidDate         = "INVALID_DATE"
)

// Error is the application error carrying a machine-readable Code and a
// human-readable Message. Cause and Fields are optional.
type Error struct {
	Code    string
	Message string
	Cause   error
	Fields  map[string]string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap exposes the underlying cause for errors.Is/errors.As.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is makes errors.Is(err, &Error{Code: "..."}) match on the code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New returns an error with the given code and message.
func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap returns an error wrapping cause while exposing only code and message
// to the client.
func Wrap(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// WithFields attaches per-field validation details to e.
func (e *Error) WithFields(fields map[string]string) *Error {
	e.Fields = fields
	return e
}

// NewBindingError converts a request body binding error into a VALIDATION_ERROR.
func NewBindingError(err error) *Error {
	if errors.Is(err, io.EOF) {
		return New(ValidationError, "Request body is required")
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make(map[string]string, len(validationErrors))

		for _, fieldErr := range validationErrors {
			fields[fieldErr.Field()] = fieldErr.Error()
		}

		return New(ValidationError, "Validation failed").WithFields(fields)
	}

	return New(ValidationError, "Invalid request body")
}

// ErrorResponse is the JSON shape of every error response.
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// HandleError writes a consistent JSON error response for err.
// Errors that are not *Error are treated as unexpected internal errors (500)
// and logged so the real cause is not leaked to the client.
func HandleError(c *gin.Context, err error) {
	var appErr *Error
	if !errors.As(err, &appErr) {
		log.Printf("codeerror: unhandled error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	if appErr.Code == StatusInternalServerError && appErr.Cause != nil {
		log.Printf("codeerror: %s: %v", appErr.Message, appErr.Cause)
	}

	c.JSON(HTTPStatus(appErr), ErrorResponse{
		Code:    appErr.Code,
		Message: appErr.Message,
		Errors:  appErr.Fields,
	})
}

// HTTPStatus maps an error to its HTTP status code.
func HTTPStatus(err error) int {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return http.StatusInternalServerError
	}

	switch appErr.Code {
	case StatusUnauthorized, InvalidCredentials, InvalidToken, MissingToken:
		return http.StatusUnauthorized
	case StatusForbidden:
		return http.StatusForbidden
	case StatusNotFound, WorkspaceNotFound, ProjectNotFound, TaskNotFound, UserNotFound:
		return http.StatusNotFound
	case StatusConflict, UserAlreadyExist, MemberAlreadyExists:
		return http.StatusConflict
	case StatusBadRequest, ValidationError, InvalidDate:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
