-- +goose Up
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    remote_url TEXT NOT NULL DEFAULT '',
    local_path TEXT NOT NULL DEFAULT '',
    default_branch TEXT NOT NULL DEFAULT 'main',
    status TEXT NOT NULL CHECK (status IN ('active', 'archived', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (remote_url <> '' OR local_path <> ''),
    UNIQUE (owner_user_id, name)
);

CREATE TABLE repo_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    head_sha TEXT NOT NULL DEFAULT '',
    last_indexed_snapshot_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repository_id, name)
);

CREATE TABLE repo_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    branch_name TEXT NOT NULL,
    commit_sha TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('indexing', 'indexed', 'failed')),
    file_count INTEGER NOT NULL DEFAULT 0,
    symbol_count INTEGER NOT NULL DEFAULT 0,
    chunk_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    indexed_at TIMESTAMPTZ
);

ALTER TABLE repo_branches
    ADD CONSTRAINT repo_branches_last_indexed_snapshot_fkey
    FOREIGN KEY (last_indexed_snapshot_id) REFERENCES repo_snapshots(id) ON DELETE SET NULL;

CREATE TABLE repository_ingestion_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    branch_name TEXT NOT NULL DEFAULT '',
    commit_sha TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled')),
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    error_message TEXT,
    locked_at TIMESTAMPTZ,
    locked_by TEXT,
    run_after TIMESTAMPTZ NOT NULL DEFAULT now(),
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    snapshot_id UUID REFERENCES repo_snapshots(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX repository_ingestion_jobs_status_run_after_idx
    ON repository_ingestion_jobs (status, run_after);

CREATE TABLE repo_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    language TEXT NOT NULL,
    is_supported BOOLEAN NOT NULL DEFAULT false,
    is_test BOOLEAN NOT NULL DEFAULT false,
    latest_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repository_id, path)
);

CREATE INDEX repo_files_repository_language_idx ON repo_files (repository_id, language);

CREATE TABLE repo_file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    snapshot_id UUID NOT NULL REFERENCES repo_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repo_files(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    language TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    content TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    is_supported BOOLEAN NOT NULL DEFAULT false,
    is_test BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (snapshot_id, path)
);

CREATE INDEX repo_file_versions_file_idx ON repo_file_versions (file_id);
CREATE INDEX repo_file_versions_snapshot_path_idx ON repo_file_versions (snapshot_id, path);

CREATE TABLE repo_symbols (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    snapshot_id UUID NOT NULL REFERENCES repo_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repo_files(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    qualified_name TEXT NOT NULL,
    kind TEXT NOT NULL,
    language TEXT NOT NULL,
    signature TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    start_line INTEGER NOT NULL DEFAULT 1,
    end_line INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX repo_symbols_repository_snapshot_idx
    ON repo_symbols (repository_id, snapshot_id);

CREATE TABLE repo_file_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    snapshot_id UUID NOT NULL REFERENCES repo_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repo_files(id) ON DELETE CASCADE,
    file_version_id UUID NOT NULL REFERENCES repo_file_versions(id) ON DELETE CASCADE,
    symbol_id UUID REFERENCES repo_symbols(id) ON DELETE SET NULL,
    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),
    chunk_type TEXT NOT NULL,
    path TEXT NOT NULL,
    language TEXT NOT NULL,
    start_line INTEGER NOT NULL DEFAULT 1,
    end_line INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (snapshot_id, file_id, chunk_index)
);

CREATE INDEX repo_file_chunks_repository_snapshot_idx
    ON repo_file_chunks (repository_id, snapshot_id);

CREATE TABLE git_commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    sha TEXT NOT NULL,
    parent_shas TEXT[] NOT NULL DEFAULT '{}',
    author_name TEXT NOT NULL DEFAULT '',
    author_email_hash TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    committed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repository_id, sha)
);

CREATE INDEX git_commits_repository_committed_at_idx
    ON git_commits (repository_id, committed_at DESC);

CREATE TABLE git_commit_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    commit_sha TEXT NOT NULL,
    path TEXT NOT NULL,
    change_type TEXT NOT NULL DEFAULT '',
    additions INTEGER NOT NULL DEFAULT 0,
    deletions INTEGER NOT NULL DEFAULT 0,
    UNIQUE (repository_id, commit_sha, path)
);

CREATE TABLE repo_retrieval_traces (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    snapshot_id UUID REFERENCES repo_snapshots(id) ON DELETE SET NULL,
    branch_name TEXT NOT NULL,
    original_query TEXT NOT NULL,
    rewritten_query TEXT NOT NULL,
    top_k INTEGER NOT NULL,
    reranker_enabled BOOLEAN NOT NULL DEFAULT false,
    dense_hits JSONB NOT NULL DEFAULT '[]',
    lexical_hits JSONB NOT NULL DEFAULT '[]',
    symbol_hits JSONB NOT NULL DEFAULT '[]',
    graph_hits JSONB NOT NULL DEFAULT '[]',
    fused_hits JSONB NOT NULL DEFAULT '[]',
    reranked_hits JSONB NOT NULL DEFAULT '[]',
    prompt_preview TEXT NOT NULL DEFAULT '',
    latency_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS repo_retrieval_traces;
DROP TABLE IF EXISTS git_commit_files;
DROP TABLE IF EXISTS git_commits;
DROP TABLE IF EXISTS repo_file_chunks;
DROP TABLE IF EXISTS repo_symbols;
DROP TABLE IF EXISTS repo_file_versions;
DROP TABLE IF EXISTS repo_files;
DROP TABLE IF EXISTS repository_ingestion_jobs;
ALTER TABLE IF EXISTS repo_branches
    DROP CONSTRAINT IF EXISTS repo_branches_last_indexed_snapshot_fkey;
DROP TABLE IF EXISTS repo_snapshots;
DROP TABLE IF EXISTS repo_branches;
DROP TABLE IF EXISTS repositories;
