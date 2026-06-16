-- +goose Up
CREATE TABLE repo_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL REFERENCES repo_retrieval_traces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    answer_correct BOOLEAN NOT NULL DEFAULT false,
    citation_correct BOOLEAN NOT NULL DEFAULT false,
    missing_file BOOLEAN NOT NULL DEFAULT false,
    missing_symbol BOOLEAN NOT NULL DEFAULT false,
    hallucinated_claim BOOLEAN NOT NULL DEFAULT false,
    should_have_refused BOOLEAN NOT NULL DEFAULT false,
    reviewer_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX repo_feedback_trace_idx ON repo_feedback (trace_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS repo_feedback;
