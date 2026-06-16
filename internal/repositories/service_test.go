package repositories

import "testing"

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
