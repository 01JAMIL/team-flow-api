package workspace

import (
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	"gin-api-1/internal/auth"
	codeerror "gin-api-1/internal/codeerror"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type handler struct {
	service Service
}

func NewWorkspaceHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateWorkspace(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload createWorkspacePayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	workspace, err := h.service.CreateWorkspace(c, createWorkspacePayload{
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
		UserID:        loggedUser.ID,
	})

	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Workspace created successfully",
		"workspace": workspace,
	})
}

func (h *handler) GetUserWorkspaceByID(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	uuidID, err := uuid.Parse(id)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found"))
		return
	}

	userID, err := uuid.Parse(loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.UserNotFound, "User not found"))
		return
	}

	workspace, err := h.service.GetUserWorkspaceByID(c, repo.GetUserWorkspaceByIDParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		ID:     pgtype.UUID{Bytes: uuidID, Valid: true},
	})

	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspace": workspace,
	})
}

func (h *handler) GetUserWorkspaces(c *gin.Context) {
	loggedUser := c.MustGet("user").(auth.UserResponse)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	userID, err := uuid.Parse(loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.UserNotFound, "User not found"))
		return
	}

	response, err := h.service.GetUserWorkspaces(c, pgtype.UUID{Bytes: userID, Valid: true}, page, pageSize)

	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *handler) UpdateWorkspace(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload updateWorkspacePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	workspace, err := h.service.UpdateWorkspace(c, updateWorkspacePayload{
		ID:            id,
		UserID:        loggedUser.ID,
		WorkspaceName: payload.WorkspaceName,
		Description:   payload.Description,
	})

	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Workspace updated successfully",
		"workspace": workspace,
	})
}

func (h *handler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	uuidID, err := uuid.Parse(id)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.WorkspaceNotFound, "Workspace not found"))
		return
	}

	userID, err := uuid.Parse(loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.UserNotFound, "User not found"))
		return
	}

	err = h.service.DeleteWorkspace(c, repo.DeleteWorkspaceParams{
		ID:     pgtype.UUID{Bytes: uuidID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})

	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace deleted successfully",
	})
}

func (h *handler) CreateCheckoutSession(c *gin.Context) {
	workspaceID := c.Param("id")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	uuidID, err := uuid.Parse(workspaceID)
	if err != nil {
		codeerror.HandleError(c, codeerror.New(codeerror.InvalidUUID, "Workspace ID is invalid"))
		return
	}

	session, err := h.service.CreateCheckoutSession(
		c,
		uuidID,
		loggedUser.ID,
	)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": session.URL,
	})
}
