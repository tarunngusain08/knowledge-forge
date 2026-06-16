package indexing

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

const (
	maxIndexedFileBytes = 1_000_000
	maxChunkLines       = 120
	chunkLineOverlap    = 20
)

var ignoredDirs = map[string]struct{}{
	".git":         {},
	".hg":          {},
	".svn":         {},
	"node_modules": {},
	"vendor":       {},
	"dist":         {},
	"build":        {},
	"target":       {},
	".next":        {},
	".turbo":       {},
	".cache":       {},
}

type RepositoryIndexer struct {
	store      *repositories.Store
	git        codeintel.GitProvider
	embedder   rag.EmbeddingProvider
	vector     rag.VectorStoreProvider
	logger     *slog.Logger
	workerName string
}

func NewRepositoryIndexer(store *repositories.Store, gitProvider codeintel.GitProvider, embedder rag.EmbeddingProvider, vector rag.VectorStoreProvider, logger *slog.Logger, workerName string) *RepositoryIndexer {
	if logger == nil {
		logger = slog.Default()
	}
	if workerName == "" {
		workerName = "repo-worker"
	}
	return &RepositoryIndexer{store: store, git: gitProvider, embedder: embedder, vector: vector, logger: logger, workerName: workerName}
}

func (i *RepositoryIndexer) Lease(ctx context.Context, limit int32) ([]codeintel.IngestionJob, error) {
	return i.store.LeaseIngestionJobs(ctx, limit, i.workerName)
}

func (i *RepositoryIndexer) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := i.store.MarkIngestionJobRunning(ctx, jobID, i.workerName)
	if err != nil {
		return fmt.Errorf("mark repository ingestion running: %w", err)
	}
	repo, err := i.store.GetRepository(ctx, job.RepositoryID)
	if err != nil {
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return fmt.Errorf("get repository: %w", err)
	}
	branch := firstNonEmpty(job.BranchName, repo.DefaultBranch, "main")
	worktree, err := i.git.ResolveWorktree(ctx, repo, branch, job.CommitSHA)
	if err != nil {
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return fmt.Errorf("resolve worktree: %w", err)
	}
	defer func() {
		if worktree.Cleanup != nil {
			_ = worktree.Cleanup()
		}
	}()
	snapshot, err := i.store.CreateSnapshot(ctx, repo.ID, branch, firstNonEmpty(job.CommitSHA, worktree.CommitSHA))
	if err != nil {
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return fmt.Errorf("create snapshot: %w", err)
	}
	indexedFiles, records, err := i.buildIndex(ctx, repo, snapshot, worktree)
	if err != nil {
		_ = i.store.MarkSnapshotFailed(ctx, snapshot.ID, err)
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return err
	}
	if err := i.store.ReplaceSnapshotIndex(ctx, snapshot, indexedFiles); err != nil {
		_ = i.store.MarkSnapshotFailed(ctx, snapshot.ID, err)
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return fmt.Errorf("store repository index: %w", err)
	}
	if err := i.vector.UpsertChunks(ctx, records); err != nil {
		_, _ = i.store.MarkIngestionJobFailed(ctx, job.ID, err)
		return fmt.Errorf("upsert repository vectors: %w", err)
	}
	if commits, err := i.git.RecentCommits(ctx, worktree, 50); err == nil {
		if err := i.store.SaveGitCommits(ctx, repo.ID, commits); err != nil {
			i.logger.Warn("save git commits", "repository_id", repo.ID, "error", err)
		}
	}
	if _, err := i.store.MarkIngestionJobSucceeded(ctx, job.ID, snapshot.ID); err != nil {
		return fmt.Errorf("mark repository ingestion succeeded: %w", err)
	}
	i.logger.Info("indexed repository", "repository_id", repo.ID, "snapshot_id", snapshot.ID, "files", len(indexedFiles), "chunks", len(records))
	return nil
}

func (i *RepositoryIndexer) buildIndex(ctx context.Context, repo codeintel.Repository, snapshot codeintel.Snapshot, worktree codeintel.Worktree) ([]repositories.IndexedFile, []rag.VectorRecord, error) {
	root, err := filepath.Abs(worktree.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve worktree path: %w", err)
	}
	var indexed []repositories.IndexedFile
	var texts []string
	var chunkRefs []*codeintel.Chunk
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if _, ignored := ignoredDirs[entry.Name()]; ignored {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := safeRelative(root, path)
		if err != nil {
			return err
		}
		language, supported := detectLanguage(rel)
		if !supported {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() <= 0 || info.Size() > maxIndexedFileBytes {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) {
			return nil
		}
		fileVersion := codeintel.FileVersion{
			ID:           uuid.New(),
			RepositoryID: repo.ID,
			SnapshotID:   snapshot.ID,
			Path:         rel,
			Language:     language,
			ContentHash:  hash(content),
			Content:      string(content),
			SizeBytes:    info.Size(),
			IsSupported:  true,
			IsTest:       isTestFile(rel),
		}
		chunks := chunkFile(repo.ID, snapshot, fileVersion, string(content))
		indexed = append(indexed, repositories.IndexedFile{Version: fileVersion, Chunks: chunks})
		for idx := range chunks {
			chunkRefs = append(chunkRefs, &indexed[len(indexed)-1].Chunks[idx])
			texts = append(texts, chunks[idx].Content)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	embeddings, err := i.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, nil, fmt.Errorf("embed repository chunks: %w", err)
	}
	records := make([]rag.VectorRecord, 0, len(embeddings))
	for idx, embedding := range embeddings {
		chunk := chunkRefs[idx]
		records = append(records, rag.VectorRecord{
			ID:     chunk.ID.String(),
			Values: embedding.Vector,
			Metadata: map[string]any{
				"source_type":   "repository_chunk",
				"repository_id": repo.ID.String(),
				"snapshot_id":   snapshot.ID.String(),
				"branch_name":   snapshot.BranchName,
				"commit_sha":    snapshot.CommitSHA,
				"path":          chunk.Path,
				"language":      chunk.Language,
				"chunk_type":    chunk.ChunkType,
				"start_line":    chunk.StartLine,
				"end_line":      chunk.EndLine,
			},
		})
	}
	return indexed, records, nil
}

func chunkFile(repositoryID uuid.UUID, snapshot codeintel.Snapshot, file codeintel.FileVersion, content string) []codeintel.Chunk {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}
	var chunks []codeintel.Chunk
	for start := 0; start < len(lines); {
		end := start + maxChunkLines
		if end > len(lines) {
			end = len(lines)
		}
		text := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		if text != "" {
			chunks = append(chunks, codeintel.Chunk{
				ID:            uuid.New(),
				RepositoryID:  repositoryID,
				SnapshotID:    snapshot.ID,
				FileVersionID: file.ID,
				ChunkIndex:    len(chunks),
				ChunkType:     codeintel.ChunkFile,
				Path:          file.Path,
				Language:      file.Language,
				StartLine:     start + 1,
				EndLine:       end,
				Content:       text,
				TokenCount:    len(strings.Fields(text)),
				Metadata: map[string]any{
					"repository_id": repositoryID.String(),
					"snapshot_id":   snapshot.ID.String(),
					"branch_name":   snapshot.BranchName,
					"commit_sha":    snapshot.CommitSHA,
					"path":          file.Path,
					"start_line":    start + 1,
					"end_line":      end,
					"language":      file.Language,
				},
			})
		}
		if end == len(lines) {
			break
		}
		start = end - chunkLineOverlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

func safeRelative(root, path string) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == "" || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("unsafe repository path %q", path)
	}
	return filepath.ToSlash(rel), nil
}

func detectLanguage(path string) (string, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return codeintel.LanguageGo, true
	case ".ts":
		return codeintel.LanguageTypeScript, true
	case ".tsx":
		return codeintel.LanguageTSX, true
	case ".md", ".mdx":
		return codeintel.LanguageMarkdown, true
	case ".txt", ".yaml", ".yml", ".json", ".toml", ".sql":
		return codeintel.LanguageText, true
	default:
		return "", false
	}
}

func isBinary(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	window := content
	if len(window) > 8192 {
		window = window[:8192]
	}
	return bytes.IndexByte(window, 0) >= 0
}

func isTestFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, "_test.go") ||
		strings.HasSuffix(lower, ".test.ts") ||
		strings.HasSuffix(lower, ".spec.ts") ||
		strings.HasSuffix(lower, ".test.tsx") ||
		strings.HasSuffix(lower, ".spec.tsx") ||
		strings.Contains(lower, "/test/") ||
		strings.Contains(lower, "/tests/")
}

func hash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
