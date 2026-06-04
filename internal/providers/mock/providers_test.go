package mock

import (
	"context"
	"testing"
)

func TestEmbeddingsAreDeterministic(t *testing.T) {
	embedder := Embeddings{Dimension: 8}
	first, err := embedder.EmbedQuery(context.Background(), "remote work policy")
	if err != nil {
		t.Fatalf("embed first: %v", err)
	}
	second, err := embedder.EmbedQuery(context.Background(), "remote work policy")
	if err != nil {
		t.Fatalf("embed second: %v", err)
	}
	if len(first.Vector) != 8 || len(second.Vector) != 8 {
		t.Fatalf("unexpected dimensions: %d %d", len(first.Vector), len(second.Vector))
	}
	for i := range first.Vector {
		if first.Vector[i] != second.Vector[i] {
			t.Fatalf("embedding mismatch at %d: %v != %v", i, first.Vector[i], second.Vector[i])
		}
	}
}
