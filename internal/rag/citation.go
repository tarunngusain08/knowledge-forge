package rag

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

func CitationFromHit(hit RetrievalHit, excerpt string) Citation {
	metadata := hit.Chunk.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return Citation{
		ChunkID:      hit.Chunk.ID,
		DocumentID:   hit.Chunk.DocumentID,
		Document:     stringValue(metadata, "filename"),
		RepositoryID: uuidValue(metadata, "repository_id"),
		SnapshotID:   uuidValue(metadata, "snapshot_id"),
		BranchName:   stringValue(metadata, "branch_name"),
		CommitSHA:    stringValue(metadata, "commit_sha"),
		Path:         stringValue(metadata, "path"),
		StartLine:    intValue(metadata, "start_line"),
		EndLine:      intValue(metadata, "end_line"),
		PageNumber:   hit.Chunk.PageNumber,
		Excerpt:      excerpt,
		DenseScore:   hit.DenseScore,
		LexicalRank:  hit.LexicalRank,
		FusedRank:    hit.FusedRank,
		RerankScore:  hit.RerankScore,
		Metadata:     metadata,
	}
}

func stringValue(metadata map[string]any, key string) string {
	value := fmt.Sprint(metadata[key])
	if value == "<nil>" {
		return ""
	}
	return value
}

func uuidValue(metadata map[string]any, key string) uuid.UUID {
	value, err := uuid.Parse(stringValue(metadata, key))
	if err != nil {
		return uuid.Nil
	}
	return value
}

func intValue(metadata map[string]any, key string) int {
	switch value := metadata[key].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, _ := strconv.Atoi(value)
		return parsed
	default:
		return 0
	}
}
