# RAG-bot Architecture

RAG-bot is a production-style company-document question answering system.

```mermaid
flowchart LR
  U[User Question] --> QR[Question Rewriter]
  QR --> QE[Vertex Query Embedding]
  QE --> P[Pinecone Dense Retrieval]
  QR --> FTS[PostgreSQL Full Text Search]
  P --> RRF[Reciprocal Rank Fusion]
  FTS --> RRF
  RRF --> RR[Vertex Ranking API]
  RR --> G[Gemini Grounded Generation]
  G --> C[Answer + Citations]
```

The Go service owns API, orchestration, auth, persistence, worker coordination,
retrieval observability, and cost accounting. Provider SDKs are hidden behind
internal interfaces so the core business logic does not depend on Vertex,
Pinecone, LangChainGo, or Ragas directly.

## Provider Boundaries

Core business logic depends on interfaces in `internal/rag`:

- `LLMProvider`
- `EmbeddingProvider`
- `VectorStoreProvider`
- `RerankerProvider`
- `LexicalSearchProvider`
- `ChunkingProvider`
- `Retriever`

Implementations live under `internal/providers`. LangChainGo is currently used
only by the chunking adapter.

## Document Storage

v1 stores uploaded source files in PostgreSQL `BYTEA`. This keeps local and
Cloud Run setup small and makes upload + metadata changes transactional. For
larger production deployments, raw file bytes should move to GCS while
PostgreSQL keeps document metadata, object URI, checksum, and indexing state.
