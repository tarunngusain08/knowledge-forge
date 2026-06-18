package rag

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func CitationFromHit(hit RetrievalHit, excerpt string) Citation {
	metadata := hit.Chunk.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata = enrichedCitationMetadata(metadata)
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

func enrichedCitationMetadata(input map[string]any) map[string]any {
	metadata := map[string]any{}
	for key, value := range input {
		metadata[key] = value
	}
	var aliases []string
	for _, key := range []string{"symbol", "symbol_name", "qualified_name"} {
		symbol := stringValue(metadata, key)
		aliases = append(aliases, symbolAliases(symbol)...)
	}
	if len(aliases) > 0 {
		metadata["symbol_aliases"] = uniqueMetadataStrings(aliases)
	}
	return metadata
}

func symbolAliases(symbol string) []string {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil
	}
	parts := splitIdentifierWords(symbol)
	if len(parts) < 3 {
		return nil
	}
	var aliases []string
	for start := 0; start < len(parts); start++ {
		for end := start + 2; end <= len(parts); end++ {
			if start == 0 && end == len(parts) {
				continue
			}
			aliases = append(aliases, strings.Join(parts[start:end], ""))
		}
	}
	return aliases
}

func splitIdentifierWords(value string) []string {
	var parts []string
	var current strings.Builder
	var previous rune
	for _, r := range value {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			previous = 0
			continue
		}
		if current.Len() > 0 && unicode.IsLower(previous) && unicode.IsUpper(r) {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteRune(r)
		previous = r
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func uniqueMetadataStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
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
