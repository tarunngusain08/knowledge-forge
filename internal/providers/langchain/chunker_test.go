package langchain

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

func TestRecursiveChunkerUsesLangChainGoSplitter(t *testing.T) {
	chunker := RecursiveChunker{ChunkSize: 40, ChunkOverlap: 5}
	chunks, err := chunker.Split(context.Background(), rag.ChunkInput{
		DocumentID: uuid.New(),
		Filename:   "policy.md",
		Content:    strings.Repeat("paid time off policy ", 20),
		Metadata:   map[string]any{"kind": "policy"},
	})
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	if chunks[0].VectorID == "" {
		t.Fatal("expected stable vector id")
	}
	if chunks[0].Metadata["filename"] != "policy.md" {
		t.Fatalf("expected filename metadata, got %+v", chunks[0].Metadata)
	}
}
