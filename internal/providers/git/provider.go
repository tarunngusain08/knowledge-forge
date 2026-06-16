package git

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
)

type Provider struct {
	CloneRoot    string
	CloneTimeout time.Duration
}

func (p Provider) ResolveWorktree(ctx context.Context, repo codeintel.Repository, branchName string, commitSHA string) (codeintel.Worktree, error) {
	if branchName == "" {
		branchName = repo.DefaultBranch
	}
	if branchName == "" {
		branchName = "main"
	}
	if repo.LocalPath != "" {
		abs, err := filepath.Abs(repo.LocalPath)
		if err != nil {
			return codeintel.Worktree{}, fmt.Errorf("resolve local repository path: %w", err)
		}
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			resolved = abs
		}
		sha := commitSHA
		if sha == "" {
			sha = gitOutput(ctx, resolved, "rev-parse", "HEAD")
		}
		return codeintel.Worktree{Path: resolved, Branch: branchName, CommitSHA: sha, Cleanup: func() error { return nil }}, nil
	}
	if repo.RemoteURL == "" {
		return codeintel.Worktree{}, fmt.Errorf("repository %s has no local path or remote url", repo.ID)
	}
	root := p.CloneRoot
	if root == "" {
		root = os.TempDir()
	}
	dir, err := os.MkdirTemp(root, "knowledge-forge-repo-*")
	if err != nil {
		return codeintel.Worktree{}, fmt.Errorf("create clone dir: %w", err)
	}
	cloneCtx := ctx
	cancel := func() {}
	timeout := p.CloneTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	cloneCtx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()
	args := []string{"clone", "--depth", "50"}
	if branchName != "" {
		args = append(args, "--branch", branchName)
	}
	args = append(args, repo.RemoteURL, dir)
	if out, err := exec.CommandContext(cloneCtx, "git", args...).CombinedOutput(); err != nil {
		_ = os.RemoveAll(dir)
		return codeintel.Worktree{}, fmt.Errorf("git clone: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if commitSHA != "" {
		if out, err := exec.CommandContext(ctx, "git", "-C", dir, "checkout", commitSHA).CombinedOutput(); err != nil {
			_ = os.RemoveAll(dir)
			return codeintel.Worktree{}, fmt.Errorf("git checkout %s: %w: %s", commitSHA, err, strings.TrimSpace(string(out)))
		}
	}
	sha := gitOutput(ctx, dir, "rev-parse", "HEAD")
	return codeintel.Worktree{
		Path:      dir,
		Branch:    branchName,
		CommitSHA: firstNonEmpty(commitSHA, sha),
		Cleanup: func() error {
			return os.RemoveAll(dir)
		},
	}, nil
}

func (Provider) RecentCommits(ctx context.Context, worktree codeintel.Worktree, limit int) ([]codeintel.GitCommit, error) {
	if limit <= 0 {
		limit = 50
	}
	format := "%H%x1f%P%x1f%an%x1f%ae%x1f%aI%x1f%s"
	cmd := exec.CommandContext(ctx, "git", "-C", worktree.Path, "log", fmt.Sprintf("-%d", limit), "--date=iso-strict", "--format="+format, "--name-status")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseGitLog(out), nil
}

func parseGitLog(out []byte) []codeintel.GitCommit {
	var commits []codeintel.GitCommit
	var current *codeintel.GitCommit
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		if len(parts) == 6 {
			if current != nil {
				commits = append(commits, *current)
			}
			committedAt, _ := time.Parse(time.RFC3339, parts[4])
			current = &codeintel.GitCommit{
				SHA:             parts[0],
				ParentSHAs:      strings.Fields(parts[1]),
				AuthorName:      parts[2],
				AuthorEmailHash: hashEmail(parts[3]),
				CommittedAt:     committedAt,
				Message:         parts[5],
			}
			continue
		}
		if current == nil {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			current.Files = append(current.Files, codeintel.GitChangedFile{
				ChangeType: fields[0],
				Path:       filepath.ToSlash(fields[len(fields)-1]),
			})
		}
	}
	if current != nil {
		commits = append(commits, *current)
	}
	return commits
}

func gitOutput(ctx context.Context, dir string, args ...string) string {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func hashEmail(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
