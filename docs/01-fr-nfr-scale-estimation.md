# Functional Requirements, Non-Functional Requirements, and Scale Estimation

## Scope

Knowledge Forge is a production-style evidence-grounded knowledge assistant.
Users can upload trusted company documents or index repository snapshots. The
system retrieves cited evidence and uses Gemini to answer questions, generate
repository deep-dive reports, and produce read-only planning/impact workflows.

## Functional Requirements

### Authentication and Users

- Allow an administrator or user to log in with email and password.
- Issue a JWT for authenticated API access.
- Support `GET /me` so clients can identify the current user.
- Restrict document, chat, debug, and evaluation actions to authenticated users.

### Document Management

- Allow users to upload PDF, Markdown, text, and supported office-style document
  inputs.
- Enforce upload constraints:
  - maximum file size: 20 MB by default,
  - MIME type validation,
  - extension allowlist,
  - SHA-256 duplicate detection per user,
  - virus scan hook interface.
- Store uploaded document bytes in PostgreSQL `BYTEA` for v1.
- List, fetch, delete, and reindex documents.
- Keep document status transitions visible:
  - `uploaded`,
  - `indexing`,
  - `indexed`,
  - `failed`,
  - `deleted`.

### Async Indexing

- Create a durable indexing job on upload.
- Process indexing jobs outside the upload request path.
- Extract document text.
- Split extracted text into chunks.
- Store chunks in PostgreSQL.
- Generate document embeddings with Vertex AI.
- Upsert chunk vectors to Pinecone.
- Retry transient failures with attempts and backoff.
- Capture job failure reason when indexing fails.

### Chat and RAG Q&A

- Create chat sessions.
- Store user and assistant messages.
- Rewrite follow-up questions using chat history.
- Retrieve relevant chunks from Pinecone and PostgreSQL FTS.
- Fuse dense and lexical results using Reciprocal Rank Fusion.
- Rerank fused candidates using Vertex Ranking API.
- Generate grounded answers using Gemini.
- Return citations with:
  - document name,
  - page number when available,
  - chunk ID,
  - excerpt,
  - dense score,
  - lexical rank,
  - fused rank,
  - rerank score.

### Retrieval Debugging

- Expose retrieval debug data.
- Show:
  - original query,
  - rewritten query,
  - dense hits,
  - lexical hits,
  - fused hits,
  - reranked hits,
  - prompt preview,
  - latency,
  - estimated cost.

### Evaluation

- Support evaluation runs.
- Compute Go-side retrieval metrics:
  - Hit Rate,
  - Recall@K,
  - MRR,
  - citation coverage,
  - retrieval latency,
  - cost per answer.
- Support Python Ragas runner for generation metrics:
  - faithfulness,
  - answer relevancy,
  - context precision,
  - context recall.

### Repository Deep-Dive Reports

- Generate an on-demand repository deep-dive report for one repository snapshot.
- Start report generation with one shared broad evidence pass.
- Run at most four targeted follow-up retrievals for weak/high-value sections.
- Return report JSON and downloadable Markdown.
- Include:
  - architecture overview,
  - entry points,
  - main packages,
  - authentication flow,
  - data layer,
  - external services,
  - testing strategy,
  - risk areas,
  - suggested improvements,
  - missing context,
  - evidence quality.
- Every report claim must be backed by citations or placed under missing
  context.

### Cost and Token Tracking

- Track provider, model, operation, input tokens, output tokens, and estimated
  cost.
- Associate cost events with users, documents, chat sessions, and eval runs when
  available.

## Non-Functional Requirements

### Availability

- User-facing API should remain available even while indexing large documents.
- Indexing failures should not crash the API.
- Worker failures should be retryable.

### Reliability

- Upload and job creation should be durable.
- Indexing jobs should be idempotent enough to safely retry.
- Failed jobs should preserve error messages for debugging.
- Deleting a document should remove related chunks and prevent stale retrieval.

### Performance

- Upload request should return quickly after validation and job creation.
- Retrieval should complete within a practical interactive latency budget.
- Dense and lexical retrieval should be independently optimizable.
- Reranking should be configurable because it improves precision but adds
  latency.
- Repository reports should avoid one full RAG pipeline per section by using a
  shared evidence pass before targeted retrieval.

### Scalability

- API and worker should scale independently.
- Worker concurrency should be bounded to protect Vertex AI, Pinecone, and
  PostgreSQL.
- PostgreSQL indexes should support full-text search and job polling.
- Pinecone should handle vector search at higher document/chunk volumes.

### Security

- Use JWT authentication for protected endpoints.
- Do not store secrets in code.
- Use Secret Manager in production.
- Validate files before accepting uploads.
- Treat document content as untrusted context during prompting.
- Avoid capturing full documents or full prompts in traces by default.

### Observability

- Emit structured logs.
- Add OpenTelemetry spans for:
  - HTTP requests,
  - database access,
  - upload,
  - extraction,
  - chunking,
  - embedding,
  - Pinecone calls,
  - FTS,
  - RRF,
  - reranking,
  - Gemini,
  - evaluation.
- Track latency, provider calls, token counts, and cost estimates.

### Maintainability

- Keep business logic behind internal interfaces.
- Keep provider SDK usage isolated in provider packages.
- Use sqlc for type-safe database access.
- Use Goose for schema migrations.
- Keep LangChainGo as an implementation detail.

## Scale Estimation

These are reasonable v1 sizing assumptions for interview discussion, not hard
capacity guarantees.

### Input Assumptions

| Item | Estimate |
|---|---:|
| Users | 100 to 1,000 internal users |
| Documents | 10,000 to 100,000 documents |
| Average document size | 500 KB to 2 MB |
| Max upload size | 20 MB |
| Chunks per document | 50 to 500 |
| Total chunks at 10k docs | 500k to 5M |
| Chat questions per day | 10k to 100k |
| Dense candidates per query | 20 |
| Final context chunks | 5 to 8 |

### Storage Estimate

Example with 10,000 documents:

```text
10,000 documents
* 1 MB average raw bytes
= ~10 GB raw document bytes in PostgreSQL BYTEA
```

Example chunks:

```text
10,000 documents
* 100 chunks per document
= 1,000,000 chunks
```

PostgreSQL stores:

- raw document bytes,
- document metadata,
- chunk text,
- FTS indexes,
- chat history,
- citations,
- traces,
- jobs,
- cost events.

Pinecone stores:

- one vector per chunk,
- metadata for filtering and citation hydration.

### Query-Time Work Estimate

For one user question:

```text
1 query rewrite call when chat history exists
1 query embedding call
1 Pinecone search
1 PostgreSQL FTS search
1 RRF fusion in Go
1 reranking call for ~20 candidates
1 Gemini generation call using top 5-8 chunks
1 trace write
1 answer + citation write
```

### Bottlenecks

| Bottleneck | Why It Matters | Mitigation |
|---|---|---|
| Embedding API rate limits | Indexing can generate many calls | Batch embeddings, rate limit workers |
| Pinecone upsert/search latency | Vector operations are remote | Batch upserts, tune metadata filters |
| PostgreSQL FTS indexes | Large chunk table can grow quickly | GIN indexes, partitioning later |
| Reranker latency | Reranker reads candidate text | Make reranker optional/configurable |
| Gemini token cost | Long prompts cost more | Limit topK, compress context |
| BYTEA storage growth | DB bloat for large files | Move raw bytes to GCS later |

### Scaling Strategy

Short term:

- Scale Cloud Run API horizontally.
- Scale worker replicas separately.
- Use bounded worker pools.
- Add provider rate limits.
- Keep indexes tuned.

Medium term:

- Move raw files from PostgreSQL `BYTEA` to GCS.
- Add Cloud Tasks for durable distributed job dispatch.
- Add per-tenant quotas.
- Add caching for repeated queries or embeddings.
- Add separate read replicas for heavy debug/eval workloads.

Long term:

- Partition large tables by tenant or time.
- Add vector namespaces per tenant.
- Add offline evaluation pipelines.
- Add SLO dashboards for retrieval latency, answer latency, failure rate, and
  cost per answer.
