-- name: CreateChatSession :one
INSERT INTO chat_sessions (user_id, title)
VALUES ($1, $2)
RETURNING id, user_id, title, created_at, updated_at;

-- name: GetChatSession :one
SELECT id, user_id, title, created_at, updated_at
FROM chat_sessions
WHERE id = $1 AND user_id = $2;

-- name: ListChatMessages :many
SELECT id, session_id, role, content, rewritten_query, created_at
FROM chat_messages
WHERE session_id = $1
ORDER BY created_at ASC;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (session_id, role, content, rewritten_query)
VALUES ($1, $2, $3, $4)
RETURNING id, session_id, role, content, rewritten_query, created_at;

-- name: CreateCitation :one
INSERT INTO citations (
    message_id,
    chunk_id,
    document_id,
    document_name,
    page_number,
    excerpt,
    dense_score,
    lexical_rank,
    fused_rank,
    rerank_score,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING id, message_id, chunk_id, document_id, document_name, page_number, excerpt, dense_score, lexical_rank, fused_rank, rerank_score, metadata, created_at;

-- name: CreateRetrievalTrace :one
INSERT INTO retrieval_traces (
    user_id,
    session_id,
    original_query,
    rewritten_query,
    top_k,
    reranker_enabled,
    dense_hits,
    lexical_hits,
    fused_hits,
    reranked_hits,
    prompt_preview,
    latency_ms,
    estimated_cost_usd
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id, user_id, session_id, original_query, rewritten_query, top_k, reranker_enabled, dense_hits, lexical_hits, fused_hits, reranked_hits, prompt_preview, latency_ms, estimated_cost_usd, created_at;

