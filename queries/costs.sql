-- name: CreateTokenCostEvent :one
INSERT INTO token_cost_events (
    user_id,
    document_id,
    chat_session_id,
    eval_run_id,
    provider,
    model,
    operation,
    input_tokens,
    output_tokens,
    estimated_cost_usd
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, user_id, document_id, chat_session_id, eval_run_id, provider, model, operation, input_tokens, output_tokens, estimated_cost_usd, created_at;

