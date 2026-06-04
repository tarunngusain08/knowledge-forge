package langchain

import (
	"context"
	"testing"
)

func TestExtractorReadsText(t *testing.T) {
	text, err := (Extractor{}).Extract(context.Background(), "handbook.md", []byte("# Handbook"))
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if text != "# Handbook" {
		t.Fatalf("unexpected text: %q", text)
	}
}
