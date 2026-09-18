package integrations

import (
	"context"
	"encoding/json"
	"gin-api-1/internal/auth"
	"gin-api-1/internal/codeerror"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type handler struct {
	service Service
}

func NewIntegrationsHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ConnectRepository(c *gin.Context) {
	projectID := c.Param("projectID")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	var payload connectRepositoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	response, err := h.service.ConnectRepository(c, projectID, loggedUser.ID, payload)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Repository connected successfully",
		"integration": response,
	})
}

func (h *handler) GetProjectIntegration(c *gin.Context) {
	projectID := c.Param("projectID")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	integration, err := h.service.GetProjectIntegration(c, projectID, loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"integration": integration,
	})
}

func (h *handler) GetProjectIntegrationTasks(c *gin.Context) {
	projectID := c.Param("projectID")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	response, err := h.service.GetProjectIntegrationTasks(c, projectID, loggedUser.ID, page, pageSize)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *handler) RegenerateSecret(c *gin.Context) {
	projectID := c.Param("projectID")
	loggedUser := c.MustGet("user").(auth.UserResponse)

	response, err := h.service.RegenerateSecret(c, projectID, loggedUser.ID)
	if err != nil {
		codeerror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Webhook secret regenerated successfully",
		"integration": response,
	})
}

func (h *handler) CreateIntegrationTask(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	event := c.GetHeader("X-GitHub-Event")

	if event == "ping" {
		c.Status(http.StatusOK)
		return
	}

	if event != "issues" {
		c.Status(http.StatusOK)
		return
	}

	var gitHubPayload gitHubIssueWebhookPayload
	if err := json.Unmarshal(body, &gitHubPayload); err != nil {
		codeerror.HandleError(c, codeerror.NewBindingError(err))
		return
	}

	createPayload := createIntegrationTaskParams{
		Provider:       "github",
		ResourceType:   "issue",
		ExternalID:     strconv.FormatInt(gitHubPayload.Issue.ID, 10),
		RepositoryName: gitHubPayload.Repository.FullName,
		IssueNumber:    int32(gitHubPayload.Issue.Number),
		Title:          gitHubPayload.Issue.Title,
		Description:    gitHubPayload.Issue.Body,
		Status:         gitHubPayload.Issue.State,
		AssigneeID:     pgtype.UUID{},
		Payload:        body,
	}

	switch gitHubPayload.Action {
	case "opened":
		integrationTask, err := h.service.CreateIntegrationTask(context.Background(), body, signature, createPayload)
		if err != nil {
			codeerror.HandleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":         "integration task created",
			"integrationTask": integrationTask,
		})
	case "closed":
		updatePayload := updateIntegrationTaskStatusParams{
			Provider:       "github",
			RepositoryName: gitHubPayload.Repository.FullName,
			ExternalID:     strconv.FormatInt(gitHubPayload.Issue.ID, 10),
			Status:         "closed",
		}

		integrationTask, err := h.service.UpdateIntegrationTaskStatus(context.Background(), body, signature, updatePayload)
		if err != nil {
			codeerror.HandleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "integration task status updated",
			"integrationTask": integrationTask,
		})
	default:
		c.Status(http.StatusOK)
	}
}
