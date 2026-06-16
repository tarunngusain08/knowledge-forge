package gitdeliverable

import (
	"bytes"
	"strings"
	"testing"
)

func TestEditPlanSupportsMessageEditingRegroupingAndAccept(t *testing.T) {
	plan := Plan{Deliverables: []Deliverable{
		{
			Message: "feat(repo): add repository service",
			Type:    "feat",
			Scope:   "repo",
			Files:   []Change{{Path: "internal/repositories/service.go"}},
		},
		{
			Message: "docs(repo): document architecture",
			Type:    "docs",
			Scope:   "repo",
			Files:   []Change{{Path: "docs/repository.md"}},
		},
	}}
	input := strings.NewReader("e 1 feat(repo): add repository service and docs\nm repository.md 1\n\n")
	var output bytes.Buffer

	edited, accepted, err := EditPlan(input, &output, plan)
	if err != nil {
		t.Fatalf("edit plan: %v", err)
	}
	if !accepted {
		t.Fatalf("expected plan to be accepted")
	}
	if len(edited.Deliverables) != 1 {
		t.Fatalf("deliverables = %#v", edited.Deliverables)
	}
	if edited.Deliverables[0].Message != "feat(repo): add repository service and docs" {
		t.Fatalf("message = %q", edited.Deliverables[0].Message)
	}
	if len(edited.Deliverables[0].Files) != 2 {
		t.Fatalf("files = %#v", edited.Deliverables[0].Files)
	}
	if !strings.Contains(output.String(), "Interactive commands:") {
		t.Fatalf("expected command help in output, got:\n%s", output.String())
	}
}

func TestEditPlanCanCreateNewDeliverableAndQuit(t *testing.T) {
	plan := Plan{Deliverables: []Deliverable{{
		Message: "feat(repo): add repository service",
		Type:    "feat",
		Scope:   "repo",
		Files:   []Change{{Path: "internal/repositories/service.go"}},
	}}}
	input := strings.NewReader("n docs(repo): document repository service\nq\n")

	edited, accepted, err := EditPlan(input, &bytes.Buffer{}, plan)
	if err != nil {
		t.Fatalf("edit plan: %v", err)
	}
	if accepted {
		t.Fatalf("expected quit to decline plan")
	}
	if len(edited.Deliverables) != 2 {
		t.Fatalf("deliverables = %#v", edited.Deliverables)
	}
	if edited.Deliverables[1].Message != "docs(repo): document repository service" {
		t.Fatalf("new deliverable = %#v", edited.Deliverables[1])
	}
}

func TestConfirmDefaultsToNo(t *testing.T) {
	confirmed, err := Confirm(strings.NewReader("\n"), &bytes.Buffer{}, "Create commits?")
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed {
		t.Fatalf("empty confirmation should default to no")
	}
}
