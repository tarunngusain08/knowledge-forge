package gitdeliverable

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGitChangesDetectsStagedUnstagedAndUntrackedFiles(t *testing.T) {
	repo := newTestRepo(t)
	writeFile(t, repo, "internal/retrieval/service.go", "package retrieval\n\nfunc Search() {}\n")
	runGit(t, repo, "add", "internal/retrieval/service.go")
	runGit(t, repo, "commit", "-m", "feat: add retrieval service")

	writeFile(t, repo, "internal/retrieval/service.go", "package retrieval\n\nfunc Search() string { return \"staged\" }\n")
	runGit(t, repo, "add", "internal/retrieval/service.go")
	writeFile(t, repo, "internal/retrieval/service.go", "package retrieval\n\nfunc Search() string { return \"unstaged\" }\n")
	writeFile(t, repo, "internal/repositories/service.go", "package repositories\n\nfunc CreateRepository() {}\n")
	writeFile(t, repo, "docs/repository-intelligence.md", "# Repository intelligence\n")

	changes, err := (Git{Dir: repo}).Changes(context.Background())
	if err != nil {
		t.Fatalf("changes: %v", err)
	}

	byPath := changesByPath(changes)
	retrieval := byPath["internal/retrieval/service.go"]
	if !retrieval.Staged || !retrieval.Unstaged || retrieval.Untracked {
		t.Fatalf("retrieval status = %+v", retrieval)
	}
	if !strings.Contains(retrieval.Diff, "staged") || !strings.Contains(retrieval.Diff, "unstaged") {
		t.Fatalf("expected staged and unstaged diffs, got:\n%s", retrieval.Diff)
	}

	repoService := byPath["internal/repositories/service.go"]
	if !repoService.Untracked || !repoService.Unstaged {
		t.Fatalf("repo service status = %+v", repoService)
	}
	if !strings.Contains(repoService.ContentHint, "CreateRepository") {
		t.Fatalf("expected content hint for untracked file, got %q", repoService.ContentHint)
	}

	docs := byPath["docs/repository-intelligence.md"]
	if !docs.Untracked {
		t.Fatalf("docs status = %+v", docs)
	}

	plan := BuildPlan(changes)
	if len(plan.Warnings) != 1 || !strings.Contains(plan.Warnings[0], "staged and unstaged") {
		t.Fatalf("warnings = %#v", plan.Warnings)
	}
}

func TestGitExecutePlanCreatesDeliverableCommits(t *testing.T) {
	repo := newTestRepo(t)
	writeFile(t, repo, "internal/repositories/service.go", "package repositories\n\nfunc CreateRepository() {}\n")
	writeFile(t, repo, "internal/retrieval/dense.go", "package retrieval\n\nfunc DenseSearch() {}\n")
	writeFile(t, repo, "internal/retrieval/dense_test.go", "package retrieval\n\nfunc TestDenseSearch() {}\n")
	writeFile(t, repo, "docs/git-deliverable-commits.md", "# git-deliverable-commits\n")

	git := Git{Dir: repo}
	changes, err := git.Changes(context.Background())
	if err != nil {
		t.Fatalf("changes: %v", err)
	}
	plan := BuildPlan(changes)

	results, err := git.ExecutePlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 commit results, got %#v", results)
	}
	for _, result := range results {
		if result.Skipped {
			t.Fatalf("did not expect skipped result: %#v", result)
		}
		if result.Hash == "" {
			t.Fatalf("missing hash for result: %#v", result)
		}
	}

	subjects := gitLogSubjects(t, repo)
	want := []string{
		"chore: initial commit",
		"feat(repo): add repository service",
		"feat(retrieval): add dense retrieval",
		"test(retrieval): add tests",
		"docs(tooling): document deliverable commit tool",
	}
	if !reflect.DeepEqual(subjects, want) {
		t.Fatalf("subjects = %#v, want %#v", subjects, want)
	}

	status := runGit(t, repo, "status", "--short")
	if strings.TrimSpace(status) != "" {
		t.Fatalf("expected clean worktree, got:\n%s", status)
	}
}

func TestGitExecutePlanSkipsEmptyDeliverables(t *testing.T) {
	repo := newTestRepo(t)
	results, err := (Git{Dir: repo}).ExecutePlan(context.Background(), Plan{
		Deliverables: []Deliverable{{Message: "docs: empty deliverable"}},
	})
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if len(results) != 1 || !results[0].Skipped || results[0].Reason != "no files" {
		t.Fatalf("results = %#v", results)
	}
	if subjects := gitLogSubjects(t, repo); !reflect.DeepEqual(subjects, []string{"chore: initial commit"}) {
		t.Fatalf("subjects = %#v", subjects)
	}
}

func TestGitChangesHandlesRenamedFiles(t *testing.T) {
	repo := newTestRepo(t)
	writeFile(t, repo, "docs/old.md", "# Old\n")
	runGit(t, repo, "add", "docs/old.md")
	runGit(t, repo, "commit", "-m", "docs: add old file")
	runGit(t, repo, "mv", "docs/old.md", "docs/new.md")

	changes, err := (Git{Dir: repo}).Changes(context.Background())
	if err != nil {
		t.Fatalf("changes: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("changes = %#v", changes)
	}
	change := changes[0]
	if !change.Renamed || change.Path != "docs/new.md" || change.OrigPath != "docs/old.md" {
		t.Fatalf("rename change = %+v", change)
	}
	if got, want := change.StagePaths(), []string{"docs/old.md", "docs/new.md"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stage paths = %#v, want %#v", got, want)
	}
}

func newTestRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "knowledge-forge-tests@example.com")
	runGit(t, repo, "config", "user.name", "Knowledge Forge Tests")
	writeFile(t, repo, "README.md", "# Test repository\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "chore: initial commit")
	return repo
}

func writeFile(t *testing.T, repo, path, contents string) {
	t.Helper()
	fullPath := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func changesByPath(changes []Change) map[string]Change {
	out := make(map[string]Change, len(changes))
	for _, change := range changes {
		out[change.DisplayPath()] = change
	}
	return out
}

func gitLogSubjects(t *testing.T, repo string) []string {
	t.Helper()
	out := runGit(t, repo, "log", "--format=%s", "--reverse")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}
