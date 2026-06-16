# Database Design and Schema

## Database Role

PostgreSQL is the durable source of truth for Knowledge Forge. It stores users,
documents, chunks, indexing jobs, chat history, citations, retrieval traces, cost
events, evaluation runs, repositories, repository snapshots, repository chunks,
Git commit metadata, and repository retrieval traces.

Pinecone is not the source of truth. Pinecone stores vector records for semantic
search, while PostgreSQL stores the authoritative metadata and text chunks.

## Entity Relationship Diagram

```mermaid
erDiagram
  users ||--o{ documents : owns
  documents ||--o{ chunks : split_into
  documents ||--o{ indexing_jobs : indexed_by

  users ||--o{ chat_sessions : owns
  chat_sessions ||--o{ chat_messages : contains
  chat_messages ||--o{ citations : supports
  chunks ||--o{ citations : cited_by
  documents ||--o{ citations : cited_by

  users ||--o{ retrieval_traces : creates
  chat_sessions ||--o{ retrieval_traces : traces

  users ||--o{ eval_runs : starts
  users ||--o{ token_cost_events : incurs
  documents ||--o{ token_cost_events : costs
  chat_sessions ||--o{ token_cost_events : costs

  users ||--o{ repositories : owns
  repositories ||--o{ repo_branches : has
  repositories ||--o{ repo_snapshots : indexed_as
  repositories ||--o{ repository_ingestion_jobs : indexed_by
  repositories ||--o{ git_commits : records
  repositories ||--o{ repo_retrieval_traces : traces
  repo_snapshots ||--o{ repo_file_versions : contains
  repo_snapshots ||--o{ repo_file_chunks : chunked_into
  repo_snapshots ||--o{ repo_symbols : defines
  repo_files ||--o{ repo_file_versions : versioned_as
  repo_files ||--o{ repo_file_chunks : chunked_as

  users {
    uuid id PK
    text email UK
    text password_hash
    text role
    timestamptz created_at
    timestamptz updated_at
  }

  documents {
    uuid id PK
    uuid owner_user_id FK
    text filename
    text content_type
    bigint size_bytes
    text sha256
    bytea raw_bytes
    text status
    text error_message
  }

  chunks {
    uuid id PK
    uuid document_id FK
    int chunk_index
    text content
    int page_number
    int token_count
    jsonb metadata
    tsvector search_vector
  }

  indexing_jobs {
    uuid id PK
    uuid document_id FK
    text status
    int attempts
    int max_attempts
    text error_message
    timestamptz locked_at
    text locked_by
    timestamptz run_after
  }

  chat_sessions {
    uuid id PK
    uuid user_id FK
    text title
  }

  chat_messages {
    uuid id PK
    uuid session_id FK
    text role
    text content
    text rewritten_query
  }

  citations {
    uuid id PK
    uuid message_id FK
    uuid chunk_id FK
    uuid document_id FK
    text document_name
    int page_number
    text excerpt
    float dense_score
    int lexical_rank
    int fused_rank
    float rerank_score
  }

  retrieval_traces {
    uuid id PK
    uuid user_id FK
    uuid session_id FK
    text original_query
    text rewritten_query
    int top_k
    boolean reranker_enabled
    jsonb dense_hits
    jsonb lexical_hits
    jsonb fused_hits
    jsonb reranked_hits
  }

  token_cost_events {
    uuid id PK
    uuid user_id FK
    uuid document_id FK
    uuid chat_session_id FK
    uuid eval_run_id
    text provider
    text model
    text operation
    int input_tokens
    int output_tokens
    numeric estimated_cost_usd
  }

  eval_runs {
    uuid id PK
    uuid user_id FK
    text name
    jsonb config
    jsonb metrics
    text status
  }

  repositories {
    uuid id PK
    uuid owner_user_id FK
    text name
    text remote_url
    text local_path
    text default_branch
    text status
  }

  repo_snapshots {
    uuid id PK
    uuid repository_id FK
    text branch_name
    text commit_sha
    text status
    int file_count
    int symbol_count
    int chunk_count
  }

  repo_file_chunks {
    uuid id PK
    uuid repository_id FK
    uuid snapshot_id FK
    uuid file_id FK
    uuid file_version_id FK
    text path
    text language
    int start_line
    int end_line
    text content
    jsonb metadata
  }

  repo_retrieval_traces {
    uuid id PK
    uuid repository_id FK
    uuid snapshot_id FK
    text original_query
    text rewritten_query
    jsonb dense_hits
    jsonb reranked_hits
  }
```

## Table-by-Table Design

### users

Stores authenticated users.

Important columns:

- `email`: unique login identity.
- `password_hash`: hashed password, never raw password.
- `role`: `admin` or `user`.

Why it exists:

- All protected resources are scoped to users.

### documents

Stores uploaded file metadata and raw bytes.

Important columns:

- `owner_user_id`: document owner.
- `filename`: original filename.
- `content_type`: MIME/content type.
- `size_bytes`: upload size.
- `sha256`: duplicate detection.
- `raw_bytes`: source file bytes.
- `status`: upload/index lifecycle.

Constraints:

- `UNIQUE (owner_user_id, sha256)` prevents duplicate uploads per user.
- `size_bytes > 0`.
- `status` is constrained to valid states.

Why BYTEA in v1:

- Simple local setup.
- Transactional file + metadata insert.
- Fewer cloud services.

Production evolution:

- Move raw bytes to GCS.
- Keep metadata, checksum, object URI, and status in PostgreSQL.

### chunks

Stores searchable document chunks.

Important columns:

- `document_id`: parent document.
- `chunk_index`: stable order within document.
- `content`: chunk text.
- `page_number`: source page when available.
- `token_count`: approximate size.
- `metadata`: provider/file metadata.
- `search_vector`: generated PostgreSQL FTS vector.

Indexes:

- `chunks_search_vector_idx`: GIN index for FTS.
- `chunks_document_id_idx`: fast document-to-chunks lookup.

Why it exists:

- Chunks are the core evidence unit used for retrieval and citations.

### repositories

Stores repository registrations owned by users.

Important columns:

- `owner_user_id`: repository owner.
- `remote_url`: Git remote used for clone-based ingestion.
- `local_path`: local path used for local demos and smoke tests.
- `default_branch`: branch used when ingestion does not specify one.
- `status`: active/archive/delete lifecycle.

Constraints:

- Either `remote_url` or `local_path` must be present.
- `(owner_user_id, name)` is unique.

### repo_snapshots

Stores immutable indexing runs for one repository branch and commit SHA.

Why it exists:

- Every repository answer and benchmark result must be reproducible against the
  exact source state that was indexed.

### repo_file_versions and repo_file_chunks

`repo_file_versions` stores file content captured in a snapshot.
`repo_file_chunks` stores retrievable evidence units with file path and line
range metadata.

Why they exist:

- Pinecone stores vectors, but PostgreSQL remains the source of truth for code
  text, line ranges, citations, and vector rebuilds.

### indexing_jobs

Durable job queue for document indexing.

Important columns:

- `document_id`: document to process.
- `status`: queued/running/succeeded/failed/cancelled.
- `attempts`: retry count.
- `max_attempts`: retry limit.
- `locked_at`, `locked_by`: worker ownership.
- `run_after`: scheduled retry time.

Index:

- `(status, run_after)` supports polling due jobs efficiently.

Why it exists:

- Upload should not block on extraction, embeddings, or Pinecone upsert.

### repository_ingestion_jobs

Durable job queue for repository indexing.

Why it exists:

- Repository indexing can be run asynchronously by the worker or synchronously
  in local smoke tests with `process_now=true`.

### repo_retrieval_traces

Stores repository Q&A retrieval evidence.

Why it exists:

- Debugging repository answers requires seeing the dense candidates, reranked
  context, prompt preview, and latency for the exact snapshot.

### chat_sessions

Groups chat messages.

Why it exists:

- Enables conversational memory and follow-up question rewriting.

### chat_messages

Stores user and assistant messages.

Important columns:

- `role`: user or assistant.
- `content`: message text.
- `rewritten_query`: standalone query used for retrieval.

Why it exists:

- Preserves conversation history and lets the system rewrite questions like
  “What about contractors?” into a standalone query.

### citations

Stores evidence attached to assistant messages.

Why it exists:

- Makes answers auditable.
- Preserves source details even after answer generation.

### retrieval_traces

Stores retrieval debugging payloads.

Important JSON fields:

- `dense_hits`
- `lexical_hits`
- `fused_hits`
- `reranked_hits`

Why it exists:

- Helps debug bad answers by showing what the retriever found at each stage.

### token_cost_events

Tracks estimated model/provider cost.

Important columns:

- `provider`
- `model`
- `operation`
- `input_tokens`
- `output_tokens`
- `estimated_cost_usd`

Why it exists:

- RAG systems need cost observability because embeddings, reranking, and
  generation all create usage-based cost.

### eval_runs

Stores evaluation runs.

Important columns:

- `config`: evaluation configuration.
- `metrics`: computed metrics.
- `status`: queued/running/succeeded/failed.

Why it exists:

- Keeps evaluation repeatable and inspectable.

## Physical Schema

Current schema source:

- `migrations/00001_initial_schema.sql`

Generated query code:

- `internal/db/*.sql.go`

SQL query definitions:

- `queries/*.sql`

## Indexing Strategy

| Index | Purpose |
|---|---|
| `chunks_search_vector_idx` | Full-text search over chunk content |
| `chunks_document_id_idx` | Fetch chunks for a document |
| `indexing_jobs_status_run_after_idx` | Worker polling |
| `chat_messages_session_id_created_at_idx` | Load chat history in order |

## Data Flow: Upload to Search

```mermaid
flowchart LR
  D[documents.raw_bytes] --> J[indexing_jobs]
  J --> W[Worker]
  W --> C[chunks.content]
  C --> FTS[chunks.search_vector]
  C --> E[Vertex Embeddings]
  E --> P[Pinecone Vector Record]
```

## Data Flow: Answer Persistence

```mermaid
flowchart LR
  Q[User Message] --> CM[chat_messages]
  CM --> RT[retrieval_traces]
  RT --> A[Assistant Message]
  A --> CIT[citations]
  A --> COST[token_cost_events]
```

## Schema Tradeoffs

### BYTEA vs GCS

`BYTEA` benefits:

- simple v1,
- transactional,
- easy local development.

`BYTEA` drawbacks:

- database grows quickly,
- backups become heavier,
- less suitable for very large files.

GCS benefits:

- better large-file storage,
- lifecycle policies,
- cheaper blob storage,
- avoids DB bloat.

### JSONB for Trace Payloads

Benefits:

- flexible shape for dense/lexical/fused/reranked hits,
- easy to evolve during early product development.

Drawbacks:

- weaker schema guarantees,
- harder analytics at large scale.

### Generated FTS Vector

Benefits:

- always consistent with chunk content,
- efficient GIN index search.

Drawbacks:

- English configuration may need tuning for multilingual documents.
