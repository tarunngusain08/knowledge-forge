package gitdeliverable

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type Git struct {
	Dir string
}

type CommitResult struct {
	Message string
	Files   []string
	Hash    string
	Skipped bool
	Reason  string
}

func (g Git) Root(ctx context.Context) (string, error) {
	out, err := g.run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (g Git) Changes(ctx context.Context) ([]Change, error) {
	if _, err := g.Root(ctx); err != nil {
		return nil, fmt.Errorf("not inside a git worktree: %w", err)
	}
	status, err := g.runBytes(ctx, "status", "--porcelain=v1", "-z", "-uall")
	if err != nil {
		return nil, err
	}
	changes, err := parseStatus(status)
	if err != nil {
		return nil, err
	}
	for i := range changes {
		diff, err := g.diffFor(ctx, changes[i])
		if err != nil {
			return nil, fmt.Errorf("read diff for %s: %w", changes[i].DisplayPath(), err)
		}
		changes[i].Diff = diff
		if changes[i].Untracked && diff == "" {
			changes[i].ContentHint = g.contentHint(changes[i].DisplayPath())
		}
	}
	return changes, nil
}

func (g Git) ExecutePlan(ctx context.Context, plan Plan) ([]CommitResult, error) {
	results := make([]CommitResult, 0, len(plan.Deliverables))
	for _, deliverable := range plan.Deliverables {
		paths := deliverable.FilePaths()
		if len(paths) == 0 {
			results = append(results, CommitResult{Message: deliverable.Message, Skipped: true, Reason: "no files"})
			continue
		}
		addArgs := append([]string{"add", "-A", "--"}, paths...)
		if _, err := g.run(ctx, addArgs...); err != nil {
			return results, fmt.Errorf("stage %q: %w", deliverable.Message, err)
		}
		diffArgs := append([]string{"diff", "--cached", "--quiet", "--"}, paths...)
		err := g.runQuiet(ctx, diffArgs...)
		if err == nil {
			results = append(results, CommitResult{Message: deliverable.Message, Files: paths, Skipped: true, Reason: "no staged diff"})
			continue
		}
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return results, fmt.Errorf("check staged diff for %q: %w", deliverable.Message, err)
		}
		commitArgs := append([]string{"commit", "-m", deliverable.Message, "--"}, paths...)
		if _, err := g.run(ctx, commitArgs...); err != nil {
			return results, fmt.Errorf("commit %q: %w", deliverable.Message, err)
		}
		hash, err := g.run(ctx, "rev-parse", "--short", "HEAD")
		if err != nil {
			return results, fmt.Errorf("read commit hash: %w", err)
		}
		results = append(results, CommitResult{
			Message: deliverable.Message,
			Files:   paths,
			Hash:    strings.TrimSpace(hash),
		})
	}
	return results, nil
}

func (g Git) diffFor(ctx context.Context, change Change) (string, error) {
	var parts []string
	if change.Staged {
		diff, err := g.run(ctx, "diff", "--cached", "--", change.DisplayPath())
		if err != nil {
			return "", err
		}
		parts = append(parts, diff)
	}
	if change.Unstaged && !change.Untracked {
		diff, err := g.run(ctx, "diff", "--", change.DisplayPath())
		if err != nil {
			return "", err
		}
		parts = append(parts, diff)
	}
	return strings.Join(parts, "\n"), nil
}

func (g Git) contentHint(path string) string {
	fullPath := filepath.Join(g.Dir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil || !utf8.Valid(data) {
		return ""
	}
	if len(data) > 16_000 {
		data = data[:16_000]
	}
	return string(data)
}

func parseStatus(raw []byte) ([]Change, error) {
	records := bytes.Split(raw, []byte{0})
	var changes []Change
	for i := 0; i < len(records); i++ {
		record := string(records[i])
		if record == "" {
			continue
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("unexpected git status record %q", record)
		}
		x := record[0]
		y := record[1]
		path := strings.TrimSpace(record[3:])
		change := Change{
			Path:      path,
			Status:    string([]byte{x, y}),
			Staged:    x != ' ' && x != '?',
			Unstaged:  y != ' ',
			Untracked: x == '?' && y == '?',
			Deleted:   x == 'D' || y == 'D',
			Renamed:   x == 'R' || y == 'R',
		}
		if change.Untracked {
			change.Unstaged = true
		}
		if change.Renamed {
			i++
			if i >= len(records) || len(records[i]) == 0 {
				return nil, fmt.Errorf("rename record for %q is missing original path", path)
			}
			change.OrigPath = string(records[i])
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (g Git) run(ctx context.Context, args ...string) (string, error) {
	out, err := g.runBytes(ctx, args...)
	return string(out), err
}

func (g Git) runBytes(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.Bytes(), fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.Bytes(), nil
}

func (g Git) runQuiet(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.Dir
	return cmd.Run()
}
