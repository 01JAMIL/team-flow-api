package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"gin-api-1/internal/codeerror"
	"gin-api-1/internal/env"
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

func (h *handler) CreateIntegrationTask(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")

	secret := env.GetEnvString("GITHUB_WEBHOOK_SECRET", "")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		c.Status(http.StatusUnauthorized)
		return
	}

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

	if gitHubPayload.Action != "opened" {
		c.Status(http.StatusOK)
		return
	}

	payload := createIntegrationTaskParams{
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
