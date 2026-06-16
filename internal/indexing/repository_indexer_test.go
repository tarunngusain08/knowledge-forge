package indexing

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
)

func TestChunkFileAddsRepositoryProvenance(t *testing.T) {
	repoID := uuid.New()
	snapshot := codeintel.Snapshot{
		ID:           uuid.New(),
		RepositoryID: repoID,
		BranchName:   "main",
		CommitSHA:    "abc123",
	}
	file := codeintel.FileVersion{
		ID:       uuid.New(),
		Path:     "internal/auth/service.go",
		Language: codeintel.LanguageGo,
	}
	content := strings.Repeat("package auth\n", maxChunkLines+5)

	chunks := chunkFile(repoID, snapshot, file, content)
	if len(chunks) < 2 {
		t.Fatalf("expected overlapping chunks, got %d", len(chunks))
	}
	first := chunks[0]
	if first.Path != file.Path || first.StartLine != 1 || first.EndLine != maxChunkLines {
		t.Fatalf("unexpected first chunk range: %#v", first)
	}
	if first.Metadata["repository_id"] != repoID.String() || first.Metadata["commit_sha"] != "abc123" {
		t.Fatalf("missing repository provenance metadata: %#v", first.Metadata)
	}
}

func TestDetectLanguageRestrictsMVPFileTypes(t *testing.T) {
	if lang, ok := detectLanguage("main.go"); !ok || lang != codeintel.LanguageGo {
		t.Fatalf("go language = %q, %v", lang, ok)
	}
	if _, ok := detectLanguage("image.png"); ok {
		t.Fatal("expected binary-looking extension to be skipped")
	}
}

func TestBinaryDetectionSkipsNullBytes(t *testing.T) {
	if !isBinary([]byte("hello\x00world")) {
		t.Fatal("expected null byte content to be binary")
	}
	if isBinary([]byte("package main\n")) {
		t.Fatal("expected source text to be non-binary")
	}
}
