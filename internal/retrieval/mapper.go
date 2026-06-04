package retrieval

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/RAG-bot/internal/db"
	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

func chunkFromVectorRow(row db.GetChunkByVectorIDRow) rag.Chunk {
	metadata := decodeMetadata(row.Metadata)
	metadata["filename"] = row.Filename
	return rag.Chunk{
		ID:         row.ID,
		DocumentID: row.DocumentID,
		VectorID:   vectorID(row.DocumentID, row.ChunkIndex),
		Index:      int(row.ChunkIndex),
		Content:    row.Content,
		PageNumber: intPtr(row.PageNumber),
		TokenCount: int(row.TokenCount),
		Metadata:   metadata,
	}
}

func chunkFromFTSRow(row db.SearchChunksFTSRow) rag.Chunk {
	metadata := decodeMetadata(row.Metadata)
	metadata["filename"] = row.Filename
	return rag.Chunk{
		ID:         row.ID,
		DocumentID: row.DocumentID,
		VectorID:   vectorID(row.DocumentID, row.ChunkIndex),
		Index:      int(row.ChunkIndex),
		Content:    row.Content,
		PageNumber: intPtr(row.PageNumber),
		TokenCount: int(row.TokenCount),
		Metadata:   metadata,
	}
}

func vectorID(documentID uuid.UUID, chunkIndex int32) string {
	return fmt.Sprintf("%s:%d", documentID.String(), chunkIndex)
}

func decodeMetadata(raw []byte) map[string]any {
	metadata := map[string]any{}
	if len(raw) == 0 {
		return metadata
	}
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func intPtr(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	i := int(value.Int32)
	return &i
}
