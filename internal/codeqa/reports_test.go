package codeqa

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

func TestBuildDeepDiveReportResponseIncludesEvidenceQualityAndMarkdown(t *testing.T) {
	repoID := uuid.New()
	snapshotID := uuid.New()
	sharedTraceID := uuid.New()
	targetedTraceID := uuid.New()
	shared := AskResponse{
		Answer: "Based on the indexed context: Knowledge Forge is a repository intelligence system with API and retrieval layers.",
		Citations: []rag.Citation{
			reportCitation(repoID, snapshotID, "cmd/api/main.go", 12, 40, "abc123", "api entrypoint"),
			reportCitation(repoID, snapshotID, "internal/retrieval/code_service.go", 20, 80, "abc123", "repository retrieval"),
		},
		TraceID: sharedTraceID,
		Provenance: AnswerProvenance{
			RepositoryID:      repoID,
			BranchName:        "main",
			SnapshotID:        snapshotID,
			CommitSHA:         "abc123",
			ContextTokenCount: 900,
		},
		Model: "mock-llm",
	}
	targeted := shared
	targeted.TraceID = targetedTraceID
	targeted.Answer = "Based on the indexed context: cmd/api/main.go starts the HTTP API."
	targeted.Citations = []rag.Citation{
		reportCitation(repoID, snapshotID, "cmd/api/main.go", 12, 40, "abc123", "func main"),
	}

	report := buildDeepDiveReportResponse(shared, []reportSectionAsk{
		{spec: deepDiveReportSectionSpecs()[0], ask: shared},
		{spec: deepDiveReportSectionSpecs()[1], ask: targeted, targeted: true},
		{spec: deepDiveReportSectionSpecs()[2], ask: shared, missingOnly: true},
	}, time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC))

	if report.Summary == "" || !strings.Contains(report.Summary, "abc123") {
		t.Fatalf("summary = %q", report.Summary)
	}
	if report.EvidenceQuality.FilesExamined != 2 {
		t.Fatalf("files examined = %d", report.EvidenceQuality.FilesExamined)
	}
	if report.EvidenceQuality.CitationCount != 3 {
		t.Fatalf("citation count = %d", report.EvidenceQuality.CitationCount)
	}
	if len(report.TraceIDs) != 2 {
		t.Fatalf("trace ids = %#v", report.TraceIDs)
	}
	if !strings.Contains(report.Markdown, "# Repository Deep-Dive Report") {
		t.Fatalf("markdown missing title:\n%s", report.Markdown)
	}
	if !strings.Contains(report.Markdown, "## Evidence Quality") {
		t.Fatalf("markdown missing evidence quality:\n%s", report.Markdown)
	}
	if !strings.Contains(report.Markdown, "cmd/api/main.go:12-40") {
		t.Fatalf("markdown missing citation:\n%s", report.Markdown)
	}
}

func TestBuildDeepDiveReportResponseKeepsUnsupportedSectionsInMissingContext(t *testing.T) {
	shared := AskResponse{
		Answer: "I could not find this in the indexed context.",
		Provenance: AnswerProvenance{
			BranchName: "main",
		},
		Model: "mock-llm",
	}

	report := buildDeepDiveReportResponse(shared, []reportSectionAsk{
		{spec: deepDiveReportSectionSpecs()[0], ask: shared},
	}, time.Now())

	if report.EvidenceQuality.Confidence.Label != "Low" {
		t.Fatalf("confidence = %#v", report.EvidenceQuality.Confidence)
	}
	if len(report.Sections[0].Findings) != 0 {
		t.Fatalf("unsupported section should not have findings: %#v", report.Sections[0].Findings)
	}
	if !contains(report.Sections[0].MissingContext, "No cited repository evidence was retrieved for this report section.") {
		t.Fatalf("missing context = %#v", report.Sections[0].MissingContext)
	}
	if !strings.Contains(report.Markdown, "No cited repository evidence was retrieved") {
		t.Fatalf("markdown missing refusal context:\n%s", report.Markdown)
	}
}

func reportCitation(repoID, snapshotID uuid.UUID, path string, start, end int, commit, excerpt string) rag.Citation {
	return rag.Citation{
		ChunkID:      uuid.New(),
		RepositoryID: repoID,
		SnapshotID:   snapshotID,
		BranchName:   "main",
		CommitSHA:    commit,
		Path:         path,
		StartLine:    start,
		EndLine:      end,
		Excerpt:      excerpt,
		DenseScore:   0.91,
	}
}
