-- name: CreateEvalRun :one
INSERT INTO eval_runs (user_id, name, config, metrics, status)
VALUES ($1, $2, $3, '{}', 'running')
RETURNING id, user_id, name, config, metrics, status, error_message, created_at, updated_at;

-- name: UpdateEvalRun :one
UPDATE eval_runs
SET metrics = $2,
    status = $3,
    error_message = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, name, config, metrics, status, error_message, created_at, updated_at;

-- name: GetEvalRun :one
SELECT id, user_id, name, config, metrics, status, error_message, created_at, updated_at
FROM eval_runs
WHERE id = $1 AND user_id = $2;

