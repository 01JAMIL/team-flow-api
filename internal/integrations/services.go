package integrations

import (
	"context"
	"crypto/rand"
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
	CreateIntegrationTask(ctx context.Context, payload createIntegrationTaskParams) (repo.IntegrationTask, error)
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

func (s *svc) CreateIntegrationTask(ctx context.Context, payload createIntegrationTaskParams) (repo.IntegrationTask, error) {
	pk := uuid.New()

	projectUUID := pgtype.UUID{}
	if payload.ProjectID != "" {
		parsed, err := uuid.Parse(payload.ProjectID)
		if err != nil {
			return repo.IntegrationTask{}, codeerror.New(codeerror.InvalidUUID, "Invalid project ID")
		}
		projectUUID = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		owner, repositoryName := splitRepository(payload.RepositoryName)

		integration, err := s.repo.GetProjectIntegrationByRepository(ctx, repo.GetProjectIntegrationByRepositoryParams{
			Provider:        payload.Provider,
			RepositoryOwner: owner,
			RepositoryName:  repositoryName,
		})
		if err != nil {
			return repo.IntegrationTask{}, codeerror.New(codeerror.ProjectNotFound, "No project connected to this repository")
		}

		projectUUID = integration.ProjectID
	}

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
		ProjectID:      projectUUID,
	})

	if err != nil {
		return repo.IntegrationTask{}, err
	}

	return integrationTask, nil
}
