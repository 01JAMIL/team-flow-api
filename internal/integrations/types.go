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
}
