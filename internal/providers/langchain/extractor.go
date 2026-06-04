package langchain

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
)

type Extractor struct{}

func (Extractor) Extract(ctx context.Context, filename string, content []byte) (string, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".txt", ".md":
		return string(content), nil
	case ".pdf":
		reader := bytes.NewReader(content)
		loader := documentloaders.NewPDF(reader, int64(len(content)))
		docs, err := loader.Load(ctx)
		if err != nil {
			return "", fmt.Errorf("load pdf with langchaingo: %w", err)
		}
		var b strings.Builder
		for _, doc := range docs {
			b.WriteString(doc.PageContent)
			b.WriteString("\n\n")
		}
		text := strings.TrimSpace(b.String())
		if text == "" {
			return "", fmt.Errorf("pdf produced no extractable text")
		}
		return text, nil
	default:
		return "", fmt.Errorf("unsupported file extension %q", filepath.Ext(filename))
	}
}
