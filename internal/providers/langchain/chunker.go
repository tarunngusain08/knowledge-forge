package langchain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/textsplitter"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type RecursiveChunker struct {
	ChunkSize    int
	ChunkOverlap int
}

func (c RecursiveChunker) Split(ctx context.Context, input rag.ChunkInput) ([]rag.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	chunkSize := c.ChunkSize
	if chunkSize == 0 {
		chunkSize = 900
	}
	overlap := c.ChunkOverlap
	if overlap == 0 {
		overlap = 120
	}
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(overlap),
	)
	parts, err := splitter.SplitText(input.Content)
	if err != nil {
		return nil, fmt.Errorf("split text with langchaingo: %w", err)
	}
	chunks := make([]rag.Chunk, 0, len(parts))
	for i, part := range parts {
		metadata := cloneMetadata(input.Metadata)
		metadata["filename"] = input.Filename
		chunks = append(chunks, rag.Chunk{
			ID:         uuid.New(),
			DocumentID: input.DocumentID,
			VectorID:   fmt.Sprintf("%s:%d", input.DocumentID.String(), i),
			Index:      i,
			Content:    part,
			TokenCount: len(part),
			Metadata:   metadata,
		})
	}
	return chunks, nil
}

func cloneMetadata(input map[string]any) map[string]any {
	metadata := make(map[string]any, len(input)+1)
	for key, value := range input {
		metadata[key] = value
	}
	return metadata
}
