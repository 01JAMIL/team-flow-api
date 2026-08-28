package integrations

import "github.com/jackc/pgx/v5/pgtype"

type createIntegrationTaskParams struct {
	Provider       string      `json:"provider" binding:"required"`
	ResourceType   string      `json:"resourceType" binding:"required"`
	ExternalID     string      `json:"externalId" binding:"required"`
	RepositoryName string      `json:"repositoryName" binding:"required"`
	IssueNumber    int32       `json:"issueNumber" binding:"required"`
	Title          string      `json:"title" binding:"required"`
	Description    string      `json:"description" binding:"required"`
	Status         string      `json:"status" binding:"required"`
	AssigneeID     pgtype.UUID `json:"assigneeId" binding:"required"`
	Payload        []byte      `json:"payload" binding:"required"`
	ProjectID      string      `json:"projectId"`
}

type connectRepositoryPayload struct {
	Repository string `json:"repository" binding:"required"`
}

type connectRepositoryResponse struct {
	Provider      string `json:"provider"`
	Repository    string `json:"repository"`
	WebhookURL    string `json:"webhookUrl"`
	WebhookSecret string `json:"webhookSecret"`
}

type projectIntegrationResponse struct {
	ID              string             `json:"id"`
	ProjectID       string             `json:"projectId"`
	Provider        string             `json:"provider"`
	RepositoryOwner string             `json:"repositoryOwner"`
	RepositoryName  string             `json:"repositoryName"`
	WebhookSecret   string             `json:"webhookSecret"`
	IsActive        bool               `json:"isActive"`
	CreatedAt       pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt       pgtype.Timestamptz `json:"updatedAt"`
}

type gitHubIssueWebhookPayload struct {
	Action string `json:"action"`

	Issue struct {
		ID     int64  `json:"id"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
	} `json:"issue"`

	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type gitHubRepoResponse struct {
	FullName string `json:"full_name"`
}
