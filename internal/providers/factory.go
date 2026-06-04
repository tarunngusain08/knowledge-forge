package providers

import (
	"context"
	"fmt"

	"github.com/tarunngusain08/RAG-bot/internal/config"
	"github.com/tarunngusain08/RAG-bot/internal/providers/langchain"
	"github.com/tarunngusain08/RAG-bot/internal/providers/mock"
	"github.com/tarunngusain08/RAG-bot/internal/providers/pinecone"
	"github.com/tarunngusain08/RAG-bot/internal/providers/vertex"
	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type IndexingProviders struct {
	Chunker  rag.ChunkingProvider
	Embedder rag.EmbeddingProvider
	Vector   rag.VectorStoreProvider
}

func NewIndexingProviders(ctx context.Context, cfg config.Config) (IndexingProviders, error) {
	chunker := langchain.RecursiveChunker{ChunkSize: 900, ChunkOverlap: 120}
	if cfg.ProviderMode == "cloud" {
		embedder, err := vertex.NewEmbeddings(ctx, cfg.GoogleProjectID, cfg.GoogleLocation, cfg.VertexEmbedModel)
		if err != nil {
			return IndexingProviders{}, err
		}
		vector := &pinecone.VectorStore{
			Host:      cfg.PineconeHost,
			APIKey:    cfg.PineconeAPIKey,
			Namespace: cfg.PineconeNamespace,
		}
		return IndexingProviders{Chunker: chunker, Embedder: embedder, Vector: vector}, nil
	}
	if cfg.ProviderMode != "mock" {
		return IndexingProviders{}, fmt.Errorf("unsupported PROVIDER_MODE %q", cfg.ProviderMode)
	}
	return IndexingProviders{
		Chunker:  chunker,
		Embedder: mock.Embeddings{Dimension: 3072, Model: "mock-embedding"},
		Vector:   &mock.VectorStore{},
	}, nil
}
