package retrieval

import (
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

func TestReciprocalRankFusionMergesDenseAndLexical(t *testing.T) {
	docID := uuid.New()
	shared := rag.RetrievalHit{Chunk: rag.Chunk{ID: uuid.New(), DocumentID: docID, VectorID: docID.String() + ":0"}, DenseScore: 0.9}
	denseOnly := rag.RetrievalHit{Chunk: rag.Chunk{ID: uuid.New(), VectorID: uuid.NewString() + ":1"}, DenseScore: 0.8}
	lexicalOnly := rag.RetrievalHit{Chunk: rag.Chunk{ID: uuid.New(), VectorID: uuid.NewString() + ":2"}, LexicalRank: 2}

	fused := ReciprocalRankFusion(
		[]rag.RetrievalHit{shared, denseOnly},
		[]rag.RetrievalHit{shared, lexicalOnly},
		10,
	)
	if len(fused) != 3 {
		t.Fatalf("expected 3 fused hits, got %d", len(fused))
	}
	if fused[0].Chunk.VectorID != shared.Chunk.VectorID {
		t.Fatalf("expected shared hit to rank first, got %+v", fused[0])
	}
	if fused[0].FusedRank != 1 {
		t.Fatalf("expected fused rank 1, got %d", fused[0].FusedRank)
	}
}
