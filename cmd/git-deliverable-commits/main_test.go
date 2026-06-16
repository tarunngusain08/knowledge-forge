package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDryRunPrintsDeliverablePlan(t *testing.T) {
	repo := newCLIRepo(t)
	writeCLIFile(t, repo, "internal/repositories/service.go", "package repositories\n\nfunc CreateRepository() {}\n")
	writeCLIFile(t, repo, "internal/retrieval/dense.go", "package retrieval\n\nfunc DenseSearch() {}\n")
	writeCLIFile(t, repo, "internal/retrieval/dense_test.go", "package retrieval\n\nfunc TestDenseSearch() {}\n")
	writeCLIFile(t, repo, "docs/repository-intelligence.md", "# Repository architecture\n")

	cmd := goRunCLI(t, "--repo", repo, "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, string(out))
	}
	output := string(out)
	for _, want := range []string{
		"Deliverable Plan",
		"feat(repo): add repository service",
		"feat(retrieval): add dense retrieval",
		"test(retrieval): add tests",
		"docs(repo): document architecture",
		"Dry run only. Re-run with --execute to create commits.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestCLIExecuteCreatesCommitsAfterConfirmation(t *testing.T) {
	repo := newCLIRepo(t)
	writeCLIFile(t, repo, "internal/repositories/service.go", "package repositories\n\nfunc CreateRepository() {}\n")
	writeCLIFile(t, repo, "docs/repository-intelligence.md", "# Repository architecture\n")

	cmd := goRunCLI(t, "--repo", repo, "--execute")
	cmd.Stdin = strings.NewReader("y\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execute: %v\n%s", err, string(out))
	}
	output := string(out)
	for _, want := range []string{
		"Create these commits now? [y/N]:",
		"Commit Results",
		"feat(repo): add repository service",
		"docs(repo): document architecture",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}

	subjects := cliGitLogSubjects(t, repo)
	wantSubjects := []string{
		"chore: initial commit",
		"feat(repo): add repository service",
		"docs(repo): document architecture",
	}
	if strings.Join(subjects, "\n") != strings.Join(wantSubjects, "\n") {
		t.Fatalf("subjects = %#v, want %#v", subjects, wantSubjects)
	}
}

func TestCLIExecuteAbortsWithoutConfirmation(t *testing.T) {
	repo := newCLIRepo(t)
	writeCLIFile(t, repo, "internal/repositories/service.go", "package repositories\n\nfunc CreateRepository() {}\n")

	cmd := goRunCLI(t, "--repo", repo, "--execute")
	cmd.Stdin = strings.NewReader("n\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execute abort: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "Aborted.") {
		t.Fatalf("expected abort output, got:\n%s", string(out))
	}
	if subjects := cliGitLogSubjects(t, repo); strings.Join(subjects, "\n") != "chore: initial commit" {
		t.Fatalf("subjects = %#v", subjects)
	}
	if status := runCLIGit(t, repo, "status", "--short"); !strings.Contains(status, "?? internal/") {
		t.Fatalf("expected untracked file to remain after abort, got:\n%s", status)
	}
}

func TestCLIRejectsDryRunAndExecuteTogether(t *testing.T) {
	repo := newCLIRepo(t)
	cmd := goRunCLI(t, "--repo", repo, "--dry-run", "--execute")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected command to fail, got output:\n%s", string(out))
	}
	if !strings.Contains(string(out), "--dry-run and --execute cannot be used together") {
		t.Fatalf("unexpected output:\n%s", string(out))
	}
}

func goRunCLI(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	fullArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", fullArgs...)
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	return cmd
}

func newCLIRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runCLIGit(t, repo, "init")
	runCLIGit(t, repo, "config", "user.email", "knowledge-forge-cli-tests@example.com")
	runCLIGit(t, repo, "config", "user.name", "Knowledge Forge CLI Tests")
	writeCLIFile(t, repo, "README.md", "# CLI test repository\n")
	runCLIGit(t, repo, "add", "README.md")
	runCLIGit(t, repo, "commit", "-m", "chore: initial commit")
	return repo
}

func writeCLIFile(t *testing.T, repo, path, contents string) {
	t.Helper()
	fullPath := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runCLIGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func cliGitLogSubjects(t *testing.T, repo string) []string {
	t.Helper()
	out := runCLIGit(t, repo, "log", "--format=%s", "--reverse")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}
