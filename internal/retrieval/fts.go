package retrieval

import (
	"context"

	"github.com/tarunngusain08/RAG-bot/internal/db"
	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type FTSStore interface {
	SearchChunksFTS(ctx context.Context, arg db.SearchChunksFTSParams) ([]db.SearchChunksFTSRow, error)
}

type PostgresFTS struct {
	store FTSStore
}

func NewPostgresFTS(store FTSStore) *PostgresFTS {
	return &PostgresFTS{store: store}
}

func (p *PostgresFTS) Search(ctx context.Context, query string, topK int) ([]rag.RetrievalHit, error) {
	if topK <= 0 {
		topK = 20
	}
	rows, err := p.store.SearchChunksFTS(ctx, db.SearchChunksFTSParams{
		WebsearchToTsquery: query,
		Limit:              int32(topK),
	})
	if err != nil {
		return nil, err
	}
	hits := make([]rag.RetrievalHit, 0, len(rows))
	for i, row := range rows {
		hits = append(hits, rag.RetrievalHit{
			Chunk:       chunkFromFTSRow(row),
			LexicalRank: i + 1,
			Source:      "lexical",
			Reasons:     []string{"postgres_fts"},
		})
	}
	return hits, nil
}
