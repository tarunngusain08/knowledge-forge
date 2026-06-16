package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

type CreateRepositoryInput struct {
	OwnerUserID   uuid.UUID
	Name          string
	RemoteURL     string
	LocalPath     string
	DefaultBranch string
}

type IndexedFile struct {
	Version codeintel.FileVersion
	Chunks  []codeintel.Chunk
	Symbols []codeintel.Symbol
}

type RetrievalTraceInput struct {
	UserID          uuid.UUID
	RepositoryID    uuid.UUID
	SnapshotID      uuid.UUID
	BranchName      string
	OriginalQuery   string
	RewrittenQuery  string
	TopK            int
	RerankerEnabled bool
	DenseHits       any
	LexicalHits     any
	SymbolHits      any
	GraphHits       any
	FusedHits       any
	RerankedHits    any
	PromptPreview   string
	LatencyMS       int64
}

type RetrievalTrace struct {
	ID             uuid.UUID       `json:"id"`
	UserID         *uuid.UUID      `json:"user_id,omitempty"`
	RepositoryID   uuid.UUID       `json:"repository_id"`
	SnapshotID     *uuid.UUID      `json:"snapshot_id,omitempty"`
	BranchName     string          `json:"branch_name"`
	OriginalQuery  string          `json:"original_query"`
	RewrittenQuery string          `json:"rewritten_query"`
	TopK           int             `json:"top_k"`
	DenseHits      json.RawMessage `json:"dense_hits"`
	LexicalHits    json.RawMessage `json:"lexical_hits"`
	SymbolHits     json.RawMessage `json:"symbol_hits"`
	GraphHits      json.RawMessage `json:"graph_hits"`
	FusedHits      json.RawMessage `json:"fused_hits"`
	RerankedHits   json.RawMessage `json:"reranked_hits"`
	PromptPreview  string          `json:"prompt_preview"`
	LatencyMS      int64           `json:"latency_ms"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (s *Store) CreateRepository(ctx context.Context, input CreateRepositoryInput) (codeintel.Repository, error) {
	if strings.TrimSpace(input.DefaultBranch) == "" {
		input.DefaultBranch = "main"
	}
	row := s.pool.QueryRow(ctx, `
INSERT INTO repositories (owner_user_id, name, remote_url, local_path, default_branch, status)
VALUES ($1, $2, $3, $4, $5, 'active')
RETURNING id, owner_user_id, name, remote_url, local_path, default_branch, status, created_at, updated_at
`, input.OwnerUserID, strings.TrimSpace(input.Name), strings.TrimSpace(input.RemoteURL), strings.TrimSpace(input.LocalPath), strings.TrimSpace(input.DefaultBranch))
	return scanRepository(row)
}

func (s *Store) GetRepositoryForUser(ctx context.Context, ownerUserID, id uuid.UUID) (codeintel.Repository, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, owner_user_id, name, remote_url, local_path, default_branch, status, created_at, updated_at
FROM repositories
WHERE id = $1 AND owner_user_id = $2 AND status <> 'deleted'
`, id, ownerUserID)
	return scanRepository(row)
}

func (s *Store) GetRepository(ctx context.Context, id uuid.UUID) (codeintel.Repository, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, owner_user_id, name, remote_url, local_path, default_branch, status, created_at, updated_at
FROM repositories
WHERE id = $1 AND status <> 'deleted'
`, id)
	return scanRepository(row)
}

func (s *Store) CreateIngestionJob(ctx context.Context, repositoryID uuid.UUID, branchName, commitSHA string, requestedBy uuid.UUID) (codeintel.IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
INSERT INTO repository_ingestion_jobs (repository_id, branch_name, commit_sha, status, requested_by)
VALUES ($1, $2, $3, 'queued', $4)
RETURNING id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
`, repositoryID, branchName, commitSHA, nullableUUID(requestedBy))
	return scanIngestionJob(row)
}

func (s *Store) GetIngestionJob(ctx context.Context, id uuid.UUID) (codeintel.IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
FROM repository_ingestion_jobs
WHERE id = $1
`, id)
	return scanIngestionJob(row)
}

func (s *Store) LeaseIngestionJobs(ctx context.Context, limit int32, workerName string) ([]codeintel.IngestionJob, error) {
	if limit <= 0 {
		limit = 1
	}
	rows, err := s.pool.Query(ctx, `
UPDATE repository_ingestion_jobs
SET status = 'running',
    attempts = attempts + 1,
    locked_at = now(),
    locked_by = $2,
    updated_at = now()
WHERE id IN (
    SELECT id
    FROM repository_ingestion_jobs
    WHERE status = 'queued' AND run_after <= now() AND attempts < max_attempts
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
RETURNING id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
`, limit, workerName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIngestionJobs(rows)
}

func (s *Store) MarkIngestionJobRunning(ctx context.Context, id uuid.UUID, workerName string) (codeintel.IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
UPDATE repository_ingestion_jobs
SET status = 'running',
    attempts = CASE WHEN status = 'queued' THEN attempts + 1 ELSE attempts END,
    locked_at = now(),
    locked_by = $2,
    updated_at = now()
WHERE id = $1 AND status IN ('queued', 'running')
RETURNING id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
`, id, workerName)
	return scanIngestionJob(row)
}

func (s *Store) MarkIngestionJobSucceeded(ctx context.Context, id uuid.UUID, snapshotID uuid.UUID) (codeintel.IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
UPDATE repository_ingestion_jobs
SET status = 'succeeded', snapshot_id = $2, error_message = NULL, updated_at = now()
WHERE id = $1
RETURNING id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
`, id, snapshotID)
	return scanIngestionJob(row)
}

func (s *Store) MarkIngestionJobFailed(ctx context.Context, id uuid.UUID, cause error) (codeintel.IngestionJob, error) {
	row := s.pool.QueryRow(ctx, `
UPDATE repository_ingestion_jobs
SET status = 'failed', error_message = $2, updated_at = now()
WHERE id = $1
RETURNING id, repository_id, branch_name, commit_sha, status, attempts, max_attempts, error_message, locked_by, requested_by, snapshot_id, created_at, updated_at
`, id, cause.Error())
	return scanIngestionJob(row)
}

func (s *Store) CreateSnapshot(ctx context.Context, repositoryID uuid.UUID, branchName, commitSHA string) (codeintel.Snapshot, error) {
	row := s.pool.QueryRow(ctx, `
INSERT INTO repo_snapshots (repository_id, branch_name, commit_sha, status)
VALUES ($1, $2, $3, 'indexing')
RETURNING id, repository_id, branch_name, commit_sha, status, file_count, symbol_count, chunk_count, error_message, created_at, indexed_at
`, repositoryID, branchName, commitSHA)
	return scanSnapshot(row)
}

func (s *Store) MarkSnapshotFailed(ctx context.Context, snapshotID uuid.UUID, cause error) error {
	_, err := s.pool.Exec(ctx, `
UPDATE repo_snapshots
SET status = 'failed', error_message = $2
WHERE id = $1
`, snapshotID, cause.Error())
	return err
}

func (s *Store) LatestSnapshot(ctx context.Context, repositoryID uuid.UUID, branchName string) (codeintel.Snapshot, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, repository_id, branch_name, commit_sha, status, file_count, symbol_count, chunk_count, error_message, created_at, indexed_at
FROM repo_snapshots
WHERE repository_id = $1 AND branch_name = $2 AND status = 'indexed'
ORDER BY indexed_at DESC NULLS LAST, created_at DESC
LIMIT 1
`, repositoryID, branchName)
	return scanSnapshot(row)
}

func (s *Store) ReplaceSnapshotIndex(ctx context.Context, snapshot codeintel.Snapshot, files []IndexedFile) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fileCount := 0
	symbolCount := 0
	chunkCount := 0
	for _, file := range files {
		fileID, err := upsertRepoFile(ctx, tx, snapshot.RepositoryID, file.Version)
		if err != nil {
			return fmt.Errorf("upsert file %s: %w", file.Version.Path, err)
		}
		file.Version.FileID = fileID
		if err := insertFileVersion(ctx, tx, file.Version); err != nil {
			return fmt.Errorf("insert file version %s: %w", file.Version.Path, err)
		}
		for _, symbol := range file.Symbols {
			symbol.RepositoryID = snapshot.RepositoryID
			symbol.SnapshotID = snapshot.ID
			symbol.FileID = fileID
			if err := insertSymbol(ctx, tx, symbol); err != nil {
				return fmt.Errorf("insert symbol %s: %w", symbol.Name, err)
			}
			symbolCount++
		}
		for _, chunk := range file.Chunks {
			chunk.RepositoryID = snapshot.RepositoryID
			chunk.SnapshotID = snapshot.ID
			chunk.FileID = fileID
			chunk.FileVersionID = file.Version.ID
			if chunk.Metadata == nil {
				chunk.Metadata = map[string]any{}
			}
			chunk.Metadata["file_id"] = fileID.String()
			if err := insertChunk(ctx, tx, chunk); err != nil {
				return fmt.Errorf("insert chunk %s:%d: %w", chunk.Path, chunk.ChunkIndex, err)
			}
			chunkCount++
		}
		fileCount++
	}
	if _, err = tx.Exec(ctx, `
UPDATE repo_snapshots
SET status = 'indexed', file_count = $2, symbol_count = $3, chunk_count = $4, indexed_at = now(), error_message = NULL
WHERE id = $1
`, snapshot.ID, fileCount, symbolCount, chunkCount); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `
INSERT INTO repo_branches (repository_id, name, head_sha, last_indexed_snapshot_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (repository_id, name)
DO UPDATE SET head_sha = EXCLUDED.head_sha,
              last_indexed_snapshot_id = EXCLUDED.last_indexed_snapshot_id,
              updated_at = now()
`, snapshot.RepositoryID, snapshot.BranchName, snapshot.CommitSHA, snapshot.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) GetChunk(ctx context.Context, repositoryID uuid.UUID, chunkID uuid.UUID) (codeintel.Chunk, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, repository_id, snapshot_id, file_id, file_version_id, symbol_id, chunk_index,
       chunk_type, path, language, start_line, end_line, content, token_count, metadata
FROM repo_file_chunks
WHERE repository_id = $1 AND id = $2
`, repositoryID, chunkID)
	return scanChunk(row)
}

func (s *Store) CreateRetrievalTrace(ctx context.Context, input RetrievalTraceInput) (uuid.UUID, error) {
	id := uuid.New()
	_, err := s.pool.Exec(ctx, `
INSERT INTO repo_retrieval_traces (
    id, user_id, repository_id, snapshot_id, branch_name, original_query, rewritten_query, top_k,
    reranker_enabled, dense_hits, lexical_hits, symbol_hits, graph_hits, fused_hits, reranked_hits,
    prompt_preview, latency_ms
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15,
    $16, $17
)
`, id, nullableUUID(input.UserID), input.RepositoryID, nullableUUID(input.SnapshotID), input.BranchName,
		input.OriginalQuery, input.RewrittenQuery, input.TopK, input.RerankerEnabled,
		mustJSON(input.DenseHits), mustJSON(input.LexicalHits), mustJSON(input.SymbolHits), mustJSON(input.GraphHits),
		mustJSON(input.FusedHits), mustJSON(input.RerankedHits), input.PromptPreview, input.LatencyMS)
	return id, err
}

func (s *Store) GetRetrievalTrace(ctx context.Context, id uuid.UUID) (RetrievalTrace, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, user_id, repository_id, snapshot_id, branch_name, original_query, rewritten_query, top_k,
       dense_hits, lexical_hits, symbol_hits, graph_hits, fused_hits, reranked_hits,
       prompt_preview, latency_ms, created_at
FROM repo_retrieval_traces
WHERE id = $1
`, id)
	var trace RetrievalTrace
	var userID pgtype.UUID
	var snapshotID pgtype.UUID
	if err := row.Scan(&trace.ID, &userID, &trace.RepositoryID, &snapshotID, &trace.BranchName, &trace.OriginalQuery, &trace.RewrittenQuery, &trace.TopK, &trace.DenseHits, &trace.LexicalHits, &trace.SymbolHits, &trace.GraphHits, &trace.FusedHits, &trace.RerankedHits, &trace.PromptPreview, &trace.LatencyMS, &trace.CreatedAt); err != nil {
		return RetrievalTrace{}, err
	}
	trace.UserID = uuidPtr(userID)
	trace.SnapshotID = uuidPtr(snapshotID)
	return trace, nil
}

func (s *Store) SaveGitCommits(ctx context.Context, repositoryID uuid.UUID, commits []codeintel.GitCommit) error {
	for _, commit := range commits {
		_, err := s.pool.Exec(ctx, `
INSERT INTO git_commits (repository_id, sha, parent_shas, author_name, author_email_hash, message, committed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (repository_id, sha) DO UPDATE
SET parent_shas = EXCLUDED.parent_shas,
    author_name = EXCLUDED.author_name,
    author_email_hash = EXCLUDED.author_email_hash,
    message = EXCLUDED.message,
    committed_at = EXCLUDED.committed_at
`, repositoryID, commit.SHA, commit.ParentSHAs, commit.AuthorName, commit.AuthorEmailHash, commit.Message, nullableTime(commit.CommittedAt))
		if err != nil {
			return err
		}
		for _, file := range commit.Files {
			_, err = s.pool.Exec(ctx, `
INSERT INTO git_commit_files (repository_id, commit_sha, path, change_type, additions, deletions)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (repository_id, commit_sha, path) DO UPDATE
SET change_type = EXCLUDED.change_type,
    additions = EXCLUDED.additions,
    deletions = EXCLUDED.deletions
`, repositoryID, commit.SHA, file.Path, file.ChangeType, file.Additions, file.Deletions)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertRepoFile(ctx context.Context, tx pgx.Tx, repositoryID uuid.UUID, file codeintel.FileVersion) (uuid.UUID, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO repo_files (repository_id, path, language, is_supported, is_test, latest_hash, status)
VALUES ($1, $2, $3, $4, $5, $6, 'active')
ON CONFLICT (repository_id, path)
DO UPDATE SET language = EXCLUDED.language,
              is_supported = EXCLUDED.is_supported,
              is_test = EXCLUDED.is_test,
              latest_hash = EXCLUDED.latest_hash,
              status = 'active',
              updated_at = now()
RETURNING id
`, repositoryID, file.Path, file.Language, file.IsSupported, file.IsTest, file.ContentHash)
	var id uuid.UUID
	return id, row.Scan(&id)
}

func insertFileVersion(ctx context.Context, tx pgx.Tx, file codeintel.FileVersion) error {
	_, err := tx.Exec(ctx, `
INSERT INTO repo_file_versions (id, repository_id, snapshot_id, file_id, path, language, content_hash, content, size_bytes, is_supported, is_test)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, file.ID, file.RepositoryID, file.SnapshotID, file.FileID, file.Path, file.Language, file.ContentHash, file.Content, file.SizeBytes, file.IsSupported, file.IsTest)
	return err
}

func insertSymbol(ctx context.Context, tx pgx.Tx, symbol codeintel.Symbol) error {
	if symbol.ID == uuid.Nil {
		symbol.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, `
INSERT INTO repo_symbols (id, repository_id, snapshot_id, file_id, name, qualified_name, kind, language, signature, content, start_line, end_line, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`, symbol.ID, symbol.RepositoryID, symbol.SnapshotID, symbol.FileID, symbol.Name, symbol.QualifiedName, symbol.Kind, symbol.Language, symbol.Signature, symbol.Content, symbol.StartLine, symbol.EndLine, mustJSON(symbol.Metadata))
	return err
}

func insertChunk(ctx context.Context, tx pgx.Tx, chunk codeintel.Chunk) error {
	_, err := tx.Exec(ctx, `
INSERT INTO repo_file_chunks (id, repository_id, snapshot_id, file_id, file_version_id, symbol_id, chunk_index, chunk_type, path, language, start_line, end_line, content, token_count, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`, chunk.ID, chunk.RepositoryID, chunk.SnapshotID, chunk.FileID, chunk.FileVersionID, nullableUUIDPtr(chunk.SymbolID), chunk.ChunkIndex, chunk.ChunkType, chunk.Path, chunk.Language, chunk.StartLine, chunk.EndLine, chunk.Content, chunk.TokenCount, mustJSON(chunk.Metadata))
	return err
}

func scanRepository(row pgx.Row) (codeintel.Repository, error) {
	var repo codeintel.Repository
	err := row.Scan(&repo.ID, &repo.OwnerUserID, &repo.Name, &repo.RemoteURL, &repo.LocalPath, &repo.DefaultBranch, &repo.Status, &repo.CreatedAt, &repo.UpdatedAt)
	return repo, err
}

func scanSnapshot(row pgx.Row) (codeintel.Snapshot, error) {
	var snapshot codeintel.Snapshot
	var errorMessage pgtype.Text
	var indexedAt pgtype.Timestamptz
	err := row.Scan(&snapshot.ID, &snapshot.RepositoryID, &snapshot.BranchName, &snapshot.CommitSHA, &snapshot.Status, &snapshot.FileCount, &snapshot.SymbolCount, &snapshot.ChunkCount, &errorMessage, &snapshot.CreatedAt, &indexedAt)
	snapshot.ErrorMessage = textPtr(errorMessage)
	if indexedAt.Valid {
		snapshot.IndexedAt = &indexedAt.Time
	}
	return snapshot, err
}

func scanIngestionJob(row pgx.Row) (codeintel.IngestionJob, error) {
	var job codeintel.IngestionJob
	var errorMessage, lockedBy pgtype.Text
	var requestedBy, snapshotID pgtype.UUID
	err := row.Scan(&job.ID, &job.RepositoryID, &job.BranchName, &job.CommitSHA, &job.Status, &job.Attempts, &job.MaxAttempts, &errorMessage, &lockedBy, &requestedBy, &snapshotID, &job.CreatedAt, &job.UpdatedAt)
	job.ErrorMessage = textPtr(errorMessage)
	job.LockedBy = textPtr(lockedBy)
	job.RequestedBy = uuidPtr(requestedBy)
	job.SnapshotID = uuidPtr(snapshotID)
	return job, err
}

func scanIngestionJobs(rows pgx.Rows) ([]codeintel.IngestionJob, error) {
	var jobs []codeintel.IngestionJob
	for rows.Next() {
		job, err := scanIngestionJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func scanChunk(row pgx.Row) (codeintel.Chunk, error) {
	var chunk codeintel.Chunk
	var symbolID pgtype.UUID
	var metadata []byte
	err := row.Scan(&chunk.ID, &chunk.RepositoryID, &chunk.SnapshotID, &chunk.FileID, &chunk.FileVersionID, &symbolID, &chunk.ChunkIndex, &chunk.ChunkType, &chunk.Path, &chunk.Language, &chunk.StartLine, &chunk.EndLine, &chunk.Content, &chunk.TokenCount, &metadata)
	chunk.SymbolID = uuidPtr(symbolID)
	chunk.Metadata = decodeMetadata(metadata)
	return chunk, err
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func decodeMetadata(raw []byte) map[string]any {
	metadata := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &metadata)
	}
	return metadata
}

func textPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func uuidPtr(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}

func nullableUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func nullableUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func nullableTime(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}
