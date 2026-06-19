package git

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
)

func TestResolveWorktreeRejectsLocalPathWhenDisabled(t *testing.T) {
	provider := Provider{Policy: codeintel.NewRepositoryPolicy(false, nil)}
	_, err := provider.ResolveWorktree(context.Background(), codeintel.Repository{
		ID:        uuid.New(),
		LocalPath: t.TempDir(),
	}, "main", "")
	if err == nil || !strings.Contains(err.Error(), "local repository paths are disabled") {
		t.Fatalf("expected disabled local path error, got %v", err)
	}
}

func TestResolveWorktreeRejectsUnsafeRemoteBeforeClone(t *testing.T) {
	provider := Provider{Policy: codeintel.NewRepositoryPolicy(false, nil)}
	_, err := provider.ResolveWorktree(context.Background(), codeintel.Repository{
		ID:        uuid.New(),
		RemoteURL: "http://169.254.169.254/latest/meta-data",
	}, "main", "")
	if err == nil || !strings.Contains(err.Error(), "remote_url must use https") {
		t.Fatalf("expected remote url validation error, got %v", err)
	}
}
