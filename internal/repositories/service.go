package repositories

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
)

type Service struct {
	store  *Store
	policy codeintel.RepositoryPolicy
}

type CreateInput struct {
	OwnerUserID   uuid.UUID `json:"owner_user_id"`
	Name          string    `json:"name"`
	RemoteURL     string    `json:"remote_url"`
	LocalPath     string    `json:"local_path"`
	DefaultBranch string    `json:"default_branch"`
}

type CreateIngestionInput struct {
	UserID       uuid.UUID `json:"user_id"`
	RepositoryID uuid.UUID `json:"repository_id"`
	BranchName   string    `json:"branch_name"`
	CommitSHA    string    `json:"commit_sha"`
}

func NewService(store *Store) *Service {
	return NewServiceWithPolicy(store, codeintel.NewRepositoryPolicy(false, nil))
}

func NewServiceWithPolicy(store *Store, policy codeintel.RepositoryPolicy) *Service {
	return &Service{store: store, policy: policy}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (codeintel.Repository, error) {
	if input.OwnerUserID == uuid.Nil {
		return codeintel.Repository{}, errors.New("owner user id is required")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = repoName(input.LocalPath, input.RemoteURL)
	}
	if name == "" {
		return codeintel.Repository{}, errors.New("repository name is required")
	}
	if err := s.policy.ValidateRegistration(input.RemoteURL, input.LocalPath); err != nil {
		return codeintel.Repository{}, err
	}
	if strings.TrimSpace(input.DefaultBranch) == "" {
		input.DefaultBranch = "main"
	}
	return s.store.CreateRepository(ctx, CreateRepositoryInput{
		OwnerUserID:   input.OwnerUserID,
		Name:          name,
		RemoteURL:     input.RemoteURL,
		LocalPath:     input.LocalPath,
		DefaultBranch: input.DefaultBranch,
	})
}

func (s *Service) Get(ctx context.Context, ownerUserID, repositoryID uuid.UUID) (codeintel.Repository, error) {
	if ownerUserID == uuid.Nil {
		return codeintel.Repository{}, errors.New("owner user id is required")
	}
	return s.store.GetRepositoryForUser(ctx, ownerUserID, repositoryID)
}

func (s *Service) CreateIngestion(ctx context.Context, input CreateIngestionInput) (codeintel.IngestionJob, error) {
	if input.RepositoryID == uuid.Nil {
		return codeintel.IngestionJob{}, errors.New("repository id is required")
	}
	repo, err := s.store.GetRepositoryForUser(ctx, input.UserID, input.RepositoryID)
	if err != nil {
		return codeintel.IngestionJob{}, fmt.Errorf("get repository: %w", err)
	}
	branch := strings.TrimSpace(input.BranchName)
	if branch == "" {
		branch = repo.DefaultBranch
	}
	return s.store.CreateIngestionJob(ctx, input.RepositoryID, branch, strings.TrimSpace(input.CommitSHA), input.UserID)
}

func (s *Service) GetIngestion(ctx context.Context, jobID uuid.UUID) (codeintel.IngestionJob, error) {
	return s.store.GetIngestionJob(ctx, jobID)
}

func (s *Service) GetIngestionForUser(ctx context.Context, userID, jobID uuid.UUID) (codeintel.IngestionJob, error) {
	if userID == uuid.Nil {
		return codeintel.IngestionJob{}, errors.New("owner user id is required")
	}
	return s.store.GetIngestionJobForUser(ctx, jobID, userID)
}

func repoName(localPath, remoteURL string) string {
	value := strings.TrimSpace(localPath)
	if value == "" {
		value = strings.TrimSuffix(strings.TrimSpace(remoteURL), ".git")
	}
	value = strings.TrimRight(value, "/")
	if value == "" {
		return ""
	}
	base := filepath.Base(value)
	return strings.TrimSuffix(base, ".git")
}
