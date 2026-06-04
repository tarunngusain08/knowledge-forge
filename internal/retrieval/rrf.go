package retrieval

import (
	"sort"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

const rrfK = 60.0

func ReciprocalRankFusion(dense, lexical []rag.RetrievalHit, limit int) []rag.RetrievalHit {
	type aggregate struct {
		hit   rag.RetrievalHit
		score float64
	}
	byID := map[string]*aggregate{}
	add := func(hits []rag.RetrievalHit, label string) {
		for i, hit := range hits {
			id := hit.Chunk.VectorID
			if id == "" {
				id = hit.Chunk.ID.String()
			}
			current, ok := byID[id]
			if !ok {
				hit.Source = "hybrid"
				hit.Reasons = append(hit.Reasons, label)
				current = &aggregate{hit: hit}
				byID[id] = current
			} else {
				current.hit.Reasons = appendMissing(current.hit.Reasons, label)
				if current.hit.DenseScore == 0 {
					current.hit.DenseScore = hit.DenseScore
				}
				if current.hit.LexicalRank == 0 {
					current.hit.LexicalRank = hit.LexicalRank
				}
			}
			current.score += 1.0 / (rrfK + float64(i+1))
		}
	}
	add(dense, "dense")
	add(lexical, "lexical")

	items := make([]aggregate, 0, len(byID))
	for _, item := range byID {
		items = append(items, *item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	fused := make([]rag.RetrievalHit, len(items))
	for i, item := range items {
		item.hit.FusedRank = i + 1
		fused[i] = item.hit
	}
	return fused
}

func appendMissing(values []string, next string) []string {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}
