package gitdeliverable

import (
	"strings"
	"testing"
)

func TestBuildPlanGroupsKnowledgeForgeDeliverables(t *testing.T) {
	changes := []Change{
		{Path: "internal/repositories/service.go", Diff: "+func CreateRepository"},
		{Path: "internal/repositories/store.go", Diff: "+type Repository"},
		{Path: "internal/retrieval/dense.go", Diff: "+func DenseSearch"},
		{Path: "internal/retrieval/retrieval_test.go", Diff: "+func TestDenseSearch"},
		{Path: "eval-runner/benchmarks/synthetic.jsonl", Diff: "+{\"question\":\"Where is auth?\"}"},
		{Path: "ui/web/src/App.tsx", Diff: "+Demo Mode"},
		{Path: "docs/repository-intelligence.md", Diff: "+Repository architecture"},
	}

	plan := BuildPlan(changes)

	messages := planMessages(plan)
	assertContains(t, messages, "feat(repo): add repository service")
	assertContains(t, messages, "feat(retrieval): add dense retrieval")
	assertContains(t, messages, "test(retrieval): add tests")
	assertContains(t, messages, "feat(eval): add benchmark corpus")
	assertContains(t, messages, "feat(ui): add demo mode")
	assertContains(t, messages, "docs(repo): document architecture")
}

func TestPlanMoveFileAndEditMessage(t *testing.T) {
	plan := Plan{Deliverables: []Deliverable{
		{Message: "feat(repo): add repository service", Files: []Change{{Path: "internal/repositories/service.go"}}},
		{Message: "docs(repo): document architecture", Files: []Change{{Path: "docs/repository.md"}}},
	}}

	if err := plan.MoveFile("repository.md", 1); err != nil {
		t.Fatalf("move file: %v", err)
	}
	if len(plan.Deliverables) != 1 {
		t.Fatalf("expected empty deliverable to be removed, got %d", len(plan.Deliverables))
	}
	if len(plan.Deliverables[0].Files) != 2 {
		t.Fatalf("expected two files in first deliverable, got %d", len(plan.Deliverables[0].Files))
	}
	if err := plan.EditMessage(1, "feat(repo): add repository docs and service"); err != nil {
		t.Fatalf("edit message: %v", err)
	}
	if plan.Deliverables[0].Message != "feat(repo): add repository docs and service" {
		t.Fatalf("message = %q", plan.Deliverables[0].Message)
	}
}

func TestBuildPlanWarnsForMixedStagedAndUnstagedFiles(t *testing.T) {
	plan := BuildPlan([]Change{{
		Path:     "internal/retrieval/service.go",
		Staged:   true,
		Unstaged: true,
	}})

	if len(plan.Warnings) != 1 || !strings.Contains(plan.Warnings[0], "staged and unstaged") {
		t.Fatalf("warnings = %#v", plan.Warnings)
	}
}

func planMessages(plan Plan) []string {
	messages := make([]string, 0, len(plan.Deliverables))
	for _, deliverable := range plan.Deliverables {
		messages = append(messages, deliverable.Message)
	}
	return messages
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("expected %q in %#v", want, values)
}
