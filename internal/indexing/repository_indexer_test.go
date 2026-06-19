package indexing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/providers/mock"
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

func TestBuildIndexSkipsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "test-secret.txt")
	if err := os.WriteFile(outside, []byte("PROJECT_FALCON must not be indexed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc Visible() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "secret.go")); err != nil {
		t.Skipf("symlink creation not supported: %v", err)
	}

	repoID := uuid.New()
	snapshot := codeintel.Snapshot{
		ID:           uuid.New(),
		RepositoryID: repoID,
		BranchName:   "main",
		CommitSHA:    "abc123",
	}
	indexer := &RepositoryIndexer{embedder: mock.Embeddings{Dimension: 8}}
	files, records, err := indexer.buildIndex(t.Context(), codeintel.Repository{ID: repoID}, snapshot, codeintel.Worktree{Path: root})
	if err != nil {
		t.Fatalf("build index: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected only the real source file to be indexed, got %d", len(files))
	}
	if files[0].Version.Path != "main.go" {
		t.Fatalf("indexed path = %q", files[0].Version.Path)
	}
	for _, record := range records {
		if strings.Contains(record.Metadata["path"].(string), "secret") {
			t.Fatalf("symlinked secret was indexed: %#v", record.Metadata)
		}
	}
}
