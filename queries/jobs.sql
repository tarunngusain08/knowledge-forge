-- name: LeaseIndexingJobs :many
UPDATE indexing_jobs
SET status = 'running',
    attempts = attempts + 1,
    locked_at = now(),
    locked_by = $2,
    updated_at = now()
WHERE id IN (
    SELECT id
    FROM indexing_jobs
    WHERE status = 'queued'
      AND run_after <= now()
    ORDER BY created_at
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING id, document_id, status, attempts, max_attempts, error_message, locked_at, locked_by, run_after, created_at, updated_at;

-- name: MarkIndexingJobRunning :one
UPDATE indexing_jobs
SET status = 'running',
    attempts = attempts + 1,
    locked_at = now(),
    locked_by = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, document_id, status, attempts, max_attempts, error_message, locked_at, locked_by, run_after, created_at, updated_at;

-- name: MarkIndexingJobSucceeded :one
UPDATE indexing_jobs
SET status = 'succeeded',
    error_message = NULL,
    updated_at = now()
WHERE id = $1
RETURNING id, document_id, status, attempts, max_attempts, error_message, locked_at, locked_by, run_after, created_at, updated_at;

-- name: MarkIndexingJobFailed :one
UPDATE indexing_jobs
SET status = 'failed',
    error_message = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, document_id, status, attempts, max_attempts, error_message, locked_at, locked_by, run_after, created_at, updated_at;

