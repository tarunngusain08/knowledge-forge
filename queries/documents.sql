-- name: CreateDocument :one
INSERT INTO documents (
    owner_user_id,
    filename,
    content_type,
    size_bytes,
    sha256,
    raw_bytes,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'uploaded'
)
RETURNING id, owner_user_id, filename, content_type, size_bytes, sha256, status, error_message, created_at, updated_at;

-- name: GetDocumentByHash :one
SELECT id, owner_user_id, filename, content_type, size_bytes, sha256, status, error_message, created_at, updated_at
FROM documents
WHERE owner_user_id = $1 AND sha256 = $2 AND status <> 'deleted';

-- name: ListDocumentsByOwner :many
SELECT id, owner_user_id, filename, content_type, size_bytes, sha256, status, error_message, created_at, updated_at
FROM documents
WHERE owner_user_id = $1 AND status <> 'deleted'
ORDER BY created_at DESC;

-- name: GetDocumentByIDAndOwner :one
SELECT id, owner_user_id, filename, content_type, size_bytes, sha256, status, error_message, created_at, updated_at
FROM documents
WHERE id = $1 AND owner_user_id = $2 AND status <> 'deleted';

-- name: GetDocumentBytes :one
SELECT id, owner_user_id, filename, content_type, size_bytes, sha256, raw_bytes, status, error_message, created_at, updated_at
FROM documents
WHERE id = $1 AND status <> 'deleted';

-- name: MarkDocumentStatus :one
UPDATE documents
SET status = $2, error_message = $3, updated_at = now()
WHERE id = $1 AND status <> 'deleted'
RETURNING id, owner_user_id, filename, content_type, size_bytes, sha256, status, error_message, created_at, updated_at;

-- name: GetChunkByVectorID :one
SELECT c.id, c.document_id, c.chunk_index, c.content, c.page_number, c.token_count, c.metadata, c.created_at,
       d.filename
FROM chunks c
JOIN documents d ON d.id = c.document_id
WHERE c.document_id = $1
  AND c.chunk_index = $2
  AND d.owner_user_id = $3
  AND d.status = 'indexed';

-- name: SearchChunksFTS :many
SELECT c.id, c.document_id, c.chunk_index, c.content, c.page_number, c.token_count, c.metadata, c.created_at,
       d.filename,
       ts_rank_cd(c.search_vector, websearch_to_tsquery('english', $1))::float8 AS lexical_score
FROM chunks c
JOIN documents d ON d.id = c.document_id
WHERE d.status = 'indexed'
  AND d.owner_user_id = $2
  AND c.search_vector @@ websearch_to_tsquery('english', $1)
ORDER BY lexical_score DESC, c.created_at DESC
LIMIT $3;

-- name: MarkDocumentDeleted :exec
UPDATE documents
SET status = 'deleted', updated_at = now()
WHERE id = $1 AND owner_user_id = $2;

-- name: CreateIndexingJob :one
INSERT INTO indexing_jobs (document_id, status)
VALUES ($1, 'queued')
RETURNING id, document_id, status, attempts, max_attempts, error_message, locked_at, locked_by, run_after, created_at, updated_at;

-- name: CreateChunk :one
INSERT INTO chunks (document_id, chunk_index, content, page_number, token_count, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, document_id, chunk_index, content, page_number, token_count, metadata, created_at;

-- name: DeleteChunksByDocument :exec
DELETE FROM chunks
WHERE document_id = $1;

-- name: CountChunksByDocument :one
SELECT count(*)
FROM chunks
WHERE document_id = $1;
