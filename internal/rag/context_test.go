package rag

import (
	"testing"

	"github.com/google/uuid"
)

func TestAssembleContextCollapsesAdjacentChunks(t *testing.T) {
	fileID := uuid.New()
	hits := []RetrievalHit{
		{
			Chunk: Chunk{
				ID:         uuid.New(),
				DocumentID: fileID,
				Content:    "func Login() {",
				TokenCount: 3,
				Metadata:   map[string]any{"path": "internal/auth/service.go", "start_line": 10, "end_line": 12},
			},
			Reasons: []string{"pinecone"},
		},
		{
			Chunk: Chunk{
				ID:         uuid.New(),
				DocumentID: fileID,
				Content:    "return nil\n}",
				TokenCount: 3,
				Metadata:   map[string]any{"path": "internal/auth/service.go", "start_line": 13, "end_line": 15},
			},
			Reasons: []string{"vertex_ranking"},
		},
	}

	assembly := AssembleContext(hits, 100)

	if len(assembly.Hits) != 1 {
		t.Fatalf("hits = %d", len(assembly.Hits))
	}
	if assembly.CollapsedCount != 1 {
		t.Fatalf("collapsed = %d", assembly.CollapsedCount)
	}
	if assembly.TokenCount != 6 {
		t.Fatalf("tokens = %d", assembly.TokenCount)
	}
	if got := assembly.Hits[0].Chunk.Metadata["end_line"]; got != 15 {
		t.Fatalf("end line = %v", got)
	}
	if len(assembly.Hits[0].Reasons) != 2 {
		t.Fatalf("reasons = %#v", assembly.Hits[0].Reasons)
	}
}

func TestAssembleContextEnforcesBudget(t *testing.T) {
	hits := []RetrievalHit{
		{Chunk: Chunk{Content: "one two three", TokenCount: 3, Metadata: map[string]any{"path": "a.go", "start_line": 1, "end_line": 1}}},
		{Chunk: Chunk{Content: "four five six", TokenCount: 3, Metadata: map[string]any{"path": "b.go", "start_line": 1, "end_line": 1}}},
	}

	assembly := AssembleContext(hits, 4)

	if len(assembly.Hits) != 1 {
		t.Fatalf("hits = %d", len(assembly.Hits))
	}
	if assembly.SkippedCount != 1 {
		t.Fatalf("skipped = %d", assembly.SkippedCount)
	}
}
