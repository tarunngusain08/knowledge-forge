-- +goose Up
ALTER TABLE repo_retrieval_traces
    ADD COLUMN query_category TEXT NOT NULL DEFAULT '',
    ADD COLUMN retrieval_path JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN retrieval_config JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN retrieved_chunk_ids TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN stage_contributions JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN context_token_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN prompt_version TEXT NOT NULL DEFAULT '',
    ADD COLUMN generation_model TEXT NOT NULL DEFAULT '',
    ADD COLUMN estimated_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE INDEX repo_retrieval_traces_category_idx
    ON repo_retrieval_traces (repository_id, query_category, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS repo_retrieval_traces_category_idx;

ALTER TABLE repo_retrieval_traces
    DROP COLUMN IF EXISTS estimated_cost_usd,
    DROP COLUMN IF EXISTS generation_model,
    DROP COLUMN IF EXISTS prompt_version,
    DROP COLUMN IF EXISTS context_token_count,
    DROP COLUMN IF EXISTS stage_contributions,
    DROP COLUMN IF EXISTS retrieved_chunk_ids,
    DROP COLUMN IF EXISTS retrieval_config,
    DROP COLUMN IF EXISTS retrieval_path,
    DROP COLUMN IF EXISTS query_category;
