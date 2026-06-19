package repositories

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRepoNameFallsBackToPathOrRemote(t *testing.T) {
	tests := []struct {
		name      string
		localPath string
		remoteURL string
		want      string
	}{
		{name: "local path", localPath: "/tmp/knowledge-forge", want: "knowledge-forge"},
		{name: "remote url", remoteURL: "https://github.com/example/service.git", want: "service"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoName(tt.localPath, tt.remoteURL); got != tt.want {
				t.Fatalf("repoName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateRejectsLocalPathByDefault(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(t.Context(), CreateInput{
		OwnerUserID: uuid.New(),
		Name:        "private",
		LocalPath:   "/etc",
	})
	if err == nil || !strings.Contains(err.Error(), "local_path repository registration is disabled") {
		t.Fatalf("expected local_path rejection, got %v", err)
	}
}

func TestCreateRejectsUnsafeRemoteURL(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(t.Context(), CreateInput{
		OwnerUserID: uuid.New(),
		Name:        "metadata",
		RemoteURL:   "https://169.254.169.254/latest/meta-data",
	})
	if err == nil {
		t.Fatal("expected unsafe remote_url rejection")
	}
}
