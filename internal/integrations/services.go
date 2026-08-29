package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	repo "gin-api-1/internal/adapters/postgresql/sqlc"
	codeerror "gin-api-1/internal/codeerror"
	"gin-api-1/internal/env"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9-_.]+/[A-Za-z0-9-_.]+$`)

type Service interface {
	ConnectRepository(ctx context.Context, projectID, loggedUserID string, payload connectRepositoryPayload) (connectRepositoryResponse, error)
	GetProjectIntegration(ctx context.Context, projectID string) (projectIntegrationResponse, error)
	RegenerateSecret(ctx context.Context, projectID, loggedUserID string) (regenerateSecretResponse, error)
	CreateIntegrationTask(ctx context.Context, body []byte, signature string, payload createIntegrationTaskParams) (repo.IntegrationTask, error)
	UpdateIntegrationTaskStatus(ctx context.Context, body []byte, signature string, payload updateIntegrationTaskStatusParams) (repo.IntegrationTask, error)
}

type svc struct {
	repo   *repo.Queries
	db     *pgx.Conn
	github *gitHubClient
}

func NewIntegrationsService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo:   repo,
		db:     db,
		github: newGitHubClient(),
	}
}

func (s *svc) ConnectRepository(ctx context.Context, projectID, loggedUserID string, payload connectRepositoryPayload) (connectRepositoryResponse, error) {
	if !repositoryPattern.MatchString(payload.Repository) {
		return connectRepositoryResponse{}, codeerror.New(codeerror.InvalidRepositoryFormat, "Repository must be in the format owner/repo_name")
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return connectRepositoryResponse{}, codeerror.New(codeerror.InvalidUUID, "Invalid project ID")
	}

	project, err := s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return connectRepositoryResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	if err := s.ensureAdminMember(ctx, project.WorkspaceID.String(), loggedUserID); err != nil {
		return connectRepositoryResponse{}, err
	}

	_, err = s.repo.GetProjectIntegrationByProjectID(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err == nil {
		return connectRepositoryResponse{}, codeerror.New(codeerror.IntegrationAlreadyExists, "Project is already connected to a repository")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return connectRepositoryResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to check project integration", err)
	}

	owner, repositoryName := splitRepository(payload.Repository)

	if err := s.github.RepositoryExists(ctx, owner, repositoryName); err != nil {
		if errors.Is(err, errRepositoryNotFound) {
			return connectRepositoryResponse{}, codeerror.New(codeerror.RepositoryNotFound, "Repository not found")
		}
		return connectRepositoryResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to verify repository", err)
	}

	webhookSecret, err := generateWebhookSecret()
	if err != nil {
		return connectRepositoryResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to generate webhook secret", err)
	}

	integration, err := s.repo.CreateProjectIntegration(ctx, repo.CreateProjectIntegrationParams{
		ID:              pgtype.UUID{Bytes: uuid.New(), Valid: true},
		ProjectID:       pgtype.UUID{Bytes: projectUUID, Valid: true},
		Provider:        "github",
		RepositoryOwner: owner,
		RepositoryName:  repositoryName,
		WebhookSecret:   webhookSecret,
	})
	if err != nil {
		return connectRepositoryResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to save integration", err)
	}

	baseURL := env.GetEnvString("APP_BASE_URL", "http://localhost:3700")
	webhookURL := baseURL + "/api/v1/webhooks/github"

	return connectRepositoryResponse{
		Provider:      integration.Provider,
		Repository:    integration.RepositoryOwner + "/" + integration.RepositoryName,
		WebhookURL:    webhookURL,
		WebhookSecret: integration.WebhookSecret,
	}, nil
}

func (s *svc) GetProjectIntegration(ctx context.Context, projectID string) (projectIntegrationResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return projectIntegrationResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	integration, err := s.repo.GetProjectIntegrationByProjectID(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return projectIntegrationResponse{}, codeerror.New(codeerror.ProjectNotFound, "No integration found for this project")
	}

	return projectIntegrationResponse{
		ID:              integration.ID.String(),
		ProjectID:       integration.ProjectID.String(),
		Provider:        integration.Provider,
		RepositoryOwner: integration.RepositoryOwner,
		RepositoryName:  integration.RepositoryName,
		IsActive:        integration.IsActive,
		CreatedAt:       integration.CreatedAt,
		UpdatedAt:       integration.UpdatedAt,
	}, nil
}

func (s *svc) RegenerateSecret(ctx context.Context, projectID, loggedUserID string) (regenerateSecretResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return regenerateSecretResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	integration, err := s.repo.GetProjectIntegrationByProjectID(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return regenerateSecretResponse{}, codeerror.New(codeerror.ProjectNotFound, "No integration found for this project")
	}

	project, err := s.repo.GetProjectById(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return regenerateSecretResponse{}, codeerror.New(codeerror.ProjectNotFound, "Project not found")
	}

	if err := s.ensureAdminMember(ctx, project.WorkspaceID.String(), loggedUserID); err != nil {
		return regenerateSecretResponse{}, err
	}

	webhookSecret, err := generateWebhookSecret()
	if err != nil {
		return regenerateSecretResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to generate webhook secret", err)
	}

	updated, err := s.repo.UpdateProjectIntegrationWebhookSecret(ctx, repo.UpdateProjectIntegrationWebhookSecretParams{
		ID:            integration.ID,
		WebhookSecret: webhookSecret,
	})
	if err != nil {
		return regenerateSecretResponse{}, codeerror.Wrap(codeerror.StatusInternalServerError, "Failed to regenerate webhook secret", err)
	}

	baseURL := env.GetEnvString("APP_BASE_URL", "http://localhost:3700")
	webhookURL := baseURL + "/api/v1/webhooks/github"

	return regenerateSecretResponse{
		Provider:      updated.Provider,
		Repository:    updated.RepositoryOwner + "/" + updated.RepositoryName,
		WebhookURL:    webhookURL,
		WebhookSecret: updated.WebhookSecret,
	}, nil
}

func (s *svc) ensureAdminMember(ctx context.Context, workspaceID, loggedUserID string) error {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return codeerror.New(codeerror.InvalidUUID, "Invalid workspace ID")
	}

	userUUID, err := uuid.Parse(loggedUserID)
	if err != nil {
		return codeerror.New(codeerror.UserNotFound, "User not found")
	}

	member, err := s.repo.GetMemberFromWorkspace(ctx, repo.GetMemberFromWorkspaceParams{
		UserID:      pgtype.UUID{Bytes: userUUID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceUUID, Valid: true},
	})
	if err != nil {
		return codeerror.New(codeerror.MemberNotFound, "Member not found")
	}

	if member.UserRole != "ADMIN" {
		return codeerror.New(codeerror.StatusForbidden, "Only ADMIN can connect a repository")
	}

	return nil
}

func splitRepository(repository string) (string, string) {
	parts := strings.SplitN(repository, "/", 2)
	return parts[0], parts[1]
}

func generateWebhookSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return hex.EncodeToString(secret), nil
}

func (s *svc) CreateIntegrationTask(ctx context.Context, body []byte, signature string, payload createIntegrationTaskParams) (repo.IntegrationTask, error) {
	integration, err := s.resolveIntegration(ctx, payload.Provider, payload.RepositoryName, body, signature)
	if err != nil {
		return repo.IntegrationTask{}, err
	}

	pk := uuid.New()

	integrationTask, err := s.repo.CreateIntegrationTask(ctx, repo.CreateIntegrationTaskParams{
		ID:             pgtype.UUID{Bytes: pk, Valid: true},
		Provider:       payload.Provider,
		ResourceType:   payload.ResourceType,
		ExternalID:     payload.ExternalID,
		RepositoryName: payload.RepositoryName,
		IssueNumber:    payload.IssueNumber,
		Title:          payload.Title,
		Description:    pgtype.Text{String: payload.Description, Valid: true},
		Status:         payload.Status,
		AssigneeID:     payload.AssigneeID,
		Payload:        payload.Payload,
		ProjectID:      integration.ProjectID,
	})

	if err != nil {
		return repo.IntegrationTask{}, err
	}

	return integrationTask, nil
}

func (s *svc) UpdateIntegrationTaskStatus(ctx context.Context, body []byte, signature string, payload updateIntegrationTaskStatusParams) (repo.IntegrationTask, error) {
	integration, err := s.resolveIntegration(ctx, payload.Provider, payload.RepositoryName, body, signature)
	if err != nil {
		return repo.IntegrationTask{}, err
	}

	integrationTask, err := s.repo.UpdateIntegrationTaskStatus(ctx, repo.UpdateIntegrationTaskStatusParams{
		ExternalID: payload.ExternalID,
		ProjectID:  integration.ProjectID,
		Status:     payload.Status,
	})
	if err != nil {
		return repo.IntegrationTask{}, codeerror.New(codeerror.ProjectNotFound, "Integration task not found")
	}

	return integrationTask, nil
}

func (s *svc) resolveIntegration(ctx context.Context, provider, repositoryName string, body []byte, signature string) (repo.ProjectIntegration, error) {
	owner, name := splitRepository(repositoryName)

	integration, err := s.repo.GetProjectIntegrationByRepository(ctx, repo.GetProjectIntegrationByRepositoryParams{
		Provider:        provider,
		RepositoryOwner: owner,
		RepositoryName:  name,
	})
	if err != nil {
		return repo.ProjectIntegration{}, codeerror.New(codeerror.ProjectNotFound, "No project connected to this repository")
	}

	if !verifyWebhookSignature(body, signature, integration.WebhookSecret) {
		return repo.ProjectIntegration{}, codeerror.New(codeerror.StatusUnauthorized, "Invalid webhook signature")
	}

	return integration, nil
}

func verifyWebhookSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
