package rag

import (
	"strings"
)

type ContextAssembly struct {
	Hits           []RetrievalHit `json:"hits"`
	TokenCount     int            `json:"token_count"`
	SkippedCount   int            `json:"skipped_count"`
	CollapsedCount int            `json:"collapsed_count"`
}

func AssembleContext(hits []RetrievalHit, maxTokens int) ContextAssembly {
	if maxTokens <= 0 {
		maxTokens = 2400
	}
	collapsed := collapseAdjacent(hits)
	out := make([]RetrievalHit, 0, len(collapsed))
	tokenCount := 0
	for _, hit := range collapsed {
		tokens := hit.Chunk.TokenCount
		if tokens <= 0 {
			tokens = estimateTokens(hit.Chunk.Content)
		}
		if tokenCount > 0 && tokenCount+tokens > maxTokens {
			continue
		}
		if tokenCount == 0 && tokens > maxTokens {
			hit.Chunk.Content = truncateByTokenEstimate(hit.Chunk.Content, maxTokens)
			hit.Chunk.TokenCount = estimateTokens(hit.Chunk.Content)
			tokens = hit.Chunk.TokenCount
		}
		tokenCount += tokens
		out = append(out, hit)
	}
	return ContextAssembly{
		Hits:           out,
		TokenCount:     tokenCount,
		SkippedCount:   len(collapsed) - len(out),
		CollapsedCount: len(hits) - len(collapsed),
	}
}

func collapseAdjacent(hits []RetrievalHit) []RetrievalHit {
	if len(hits) < 2 {
		return hits
	}
	out := make([]RetrievalHit, 0, len(hits))
	for _, hit := range hits {
		if len(out) == 0 {
			out = append(out, hit)
			continue
		}
		last := &out[len(out)-1]
		if samePath(last.Chunk.Metadata, hit.Chunk.Metadata) && adjacentLines(last.Chunk, hit.Chunk) {
			if last.Chunk.Metadata == nil {
				last.Chunk.Metadata = map[string]any{}
			}
			last.Chunk.Content = strings.TrimRight(last.Chunk.Content, "\n") + "\n" + strings.TrimLeft(hit.Chunk.Content, "\n")
			last.Chunk.TokenCount += hit.Chunk.TokenCount
			last.Chunk.Metadata["end_line"] = metadataInt(hit.Chunk.Metadata, "end_line")
			last.Reasons = appendUnique(last.Reasons, hit.Reasons...)
			if hit.RerankScore > last.RerankScore {
				last.RerankScore = hit.RerankScore
			}
			continue
		}
		out = append(out, hit)
	}
	return out
}

func samePath(a, b map[string]any) bool {
	return metadataString(a, "path") != "" && metadataString(a, "path") == metadataString(b, "path")
}

func adjacentLines(a, b Chunk) bool {
	aEnd := metadataInt(a.Metadata, "end_line")
	bStart := metadataInt(b.Metadata, "start_line")
	return aEnd > 0 && bStart > 0 && bStart <= aEnd+1
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return value
}

func metadataInt(metadata map[string]any, key string) int {
	if metadata == nil {
		return 0
	}
	switch value := metadata[key].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func estimateTokens(text string) int {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 0
	}
	return (len(fields) * 4 / 3) + 1
}

func truncateByTokenEstimate(text string, maxTokens int) string {
	if maxTokens <= 0 {
		return ""
	}
	words := strings.Fields(text)
	maxWords := maxTokens * 3 / 4
	if maxWords <= 0 || len(words) <= maxWords {
		return text
	}
	return strings.Join(words[:maxWords], " ")
}

func appendUnique(existing []string, values ...string) []string {
	seen := map[string]bool{}
	for _, value := range existing {
		seen[value] = true
	}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		existing = append(existing, value)
	}
	return existing
}
