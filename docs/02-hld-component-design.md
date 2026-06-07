# High-Level Design and Component Design

## System Context

Knowledge Forge is a RAG system for company documents. The system turns uploaded
files into searchable chunks, retrieves relevant chunks for a user question, and
uses Gemini to produce grounded answers with citations.

```mermaid
flowchart LR
  User[User] --> UI[Streamlit Demo UI]
  UI --> API[Go Chi API]
  API --> PG[(PostgreSQL)]
  API --> Pinecone[(Pinecone Vector DB)]
  API --> Vertex[Vertex AI]
  Worker[Go Indexing Worker] --> PG
  Worker --> Pinecone
  Worker --> Vertex
```

## Deployment HLD

```mermaid
flowchart TB
  subgraph GCP[Google Cloud]
    API[Cloud Run API]
    Worker[Cloud Run Worker]
    UI[Cloud Run Streamlit UI]
    Migrate[Cloud Run Migration Job]
    SQL[(Cloud SQL PostgreSQL)]
    Secrets[Secret Manager]
    Trace[Cloud Trace / OpenTelemetry]
    Tasks[Cloud Tasks]
  end

  Pinecone[(Pinecone)]
  Vertex[Vertex AI: Gemini, Embeddings, Ranking]

  UI --> API
  API --> SQL
  API --> Vertex
  API --> Pinecone
  API --> Secrets
  API --> Tasks
  Tasks --> Worker
  Worker --> SQL
  Worker --> Vertex
  Worker --> Pinecone
  Migrate --> SQL
  API -. traces .-> Trace
  Worker -. traces .-> Trace
```

## Main Runtime Components

### Streamlit UI

Purpose:

- Provides a demo frontend for uploads, chat, debug, and evaluation.

Why it exists:

- Makes the project easy to demo without building a full React application.

If removed:

- The system still works through APIs, but demos become less visual.

### Go Chi API

Purpose:

- Owns HTTP routing, auth, document APIs, chat APIs, debug endpoints, and eval
  endpoints.

Why it exists:

- Keeps business workflows behind a production-style backend instead of putting
  logic in the UI.

If removed:

- No stable API surface; UI and workers would have no coordination point.

### Indexing Worker

Purpose:

- Processes durable indexing jobs.
- Extracts text, chunks documents, generates embeddings, and upserts vectors.

Why it exists:

- Indexing is slow and should not block uploads.

If removed:

- Uploads would either never become searchable or would need to perform expensive
  indexing inline.

### PostgreSQL

Purpose:

- Stores durable relational data:
  - users,
  - documents,
  - chunks,
  - jobs,
  - chat sessions,
  - citations,
  - retrieval traces,
  - eval runs,
  - token cost events.

Why it exists:

- RAG needs durable metadata, transactional state, and keyword search.

If removed:

- The system loses source-of-truth storage, job durability, FTS, chat history,
  and auditability.

### Pinecone

Purpose:

- Stores chunk embeddings and supports semantic nearest-neighbor search.

Why it exists:

- User questions often use different words than the documents. Pinecone finds
  chunks by meaning, not only exact words.

If removed:

- Search becomes keyword-only and misses semantic matches.

### Vertex AI

Purpose:

- Provides:
  - document/query embeddings,
  - Gemini answer generation,
  - ranking/reranking.

Why it exists:

- Centralizes managed model access on Google Cloud.

If removed:

- The system needs replacement providers for embeddings, generation, and
  reranking.

### Provider Layer

Purpose:

- Hides external SDKs behind internal interfaces:
  - `LLMProvider`,
  - `EmbeddingProvider`,
  - `VectorStoreProvider`,
  - `RerankerProvider`,
  - `LexicalSearchProvider`,
  - `ChunkingProvider`,
  - `Retriever`.

Why it exists:

- Keeps core business logic independent of Gemini, Pinecone, LangChainGo, and
  other SDKs.

If removed:

- Business logic becomes tightly coupled to vendors and hard to test.

## Component Diagram

```mermaid
flowchart TB
  subgraph API[cmd/api]
    Router[HTTP Router]
    AuthHandlers[Auth Handlers]
    DocumentHandlers[Document Handlers]
    ChatHandlers[Chat Handlers]
    EvalHandlers[Eval Handlers]
    JobHandlers[Internal Job Handlers]
  end

  subgraph Domain[Domain Services]
    Auth[Auth Service]
    Documents[Document Service]
    Chat[Chat Service]
    Retrieval[Retrieval Service]
    Evaluation[Evaluation Service]
    Costs[Cost Service]
  end

  subgraph Providers[Provider Implementations]
    LangChain[LangChainGo Chunker/Extractor]
    VertexEmb[Vertex Embeddings]
    VertexGemini[Gemini LLM]
    VertexRank[Vertex Ranking]
    Pinecone[Pinecone Vector Store]
    FTS[PostgreSQL FTS Provider]
  end

  subgraph Storage[Storage]
    DB[(PostgreSQL)]
    Vector[(Pinecone Index)]
  end

  Router --> AuthHandlers
  Router --> DocumentHandlers
  Router --> ChatHandlers
  Router --> EvalHandlers
  Router --> JobHandlers

  AuthHandlers --> Auth
  DocumentHandlers --> Documents
  ChatHandlers --> Chat
  EvalHandlers --> Evaluation
  JobHandlers --> Documents

  Chat --> Retrieval
  Documents --> LangChain
  Documents --> VertexEmb
  Documents --> Pinecone
  Retrieval --> VertexEmb
  Retrieval --> Pinecone
  Retrieval --> FTS
  Retrieval --> VertexRank
  Chat --> VertexGemini

  Auth --> DB
  Documents --> DB
  Chat --> DB
  Evaluation --> DB
  Costs --> DB
  Pinecone --> Vector
  FTS --> DB
```

## Request-Time HLD

```mermaid
flowchart LR
  Q[Question] --> API[API]
  API --> CH[Chat Service]
  CH --> RW[Question Rewrite]
  RW --> EMB[Query Embedding]
  EMB --> Dense[Pinecone Dense Search]
  RW --> Lex[Postgres FTS Search]
  Dense --> RRF[RRF Fusion]
  Lex --> RRF
  RRF --> Rank[Vertex Reranker]
  Rank --> Prompt[Prompt Assembly]
  Prompt --> Gemini[Gemini]
  Gemini --> Answer[Answer]
  Rank --> Citations[Citations]
  Answer --> Store[Store Message + Trace]
```

## Indexing-Time HLD

```mermaid
flowchart LR
  Upload[Document Upload] --> Validate[Validation]
  Validate --> Save[Store Raw Bytes]
  Save --> Job[Create indexing_jobs Row]
  Job --> Worker[Worker Polls Job]
  Worker --> Extract[Extract Text]
  Extract --> Chunk[Split into Chunks]
  Chunk --> StoreChunks[Store Chunks in PostgreSQL]
  Chunk --> Embed[Vertex Document Embeddings]
  Embed --> Upsert[Pinecone Upsert]
  Upsert --> Indexed[Mark Document Indexed]
```

## Key Design Decisions

| Decision | Reason | Tradeoff |
|---|---|---|
| Go backend | Strong concurrency, typed services, production backend signal | Less AI ecosystem depth than Python |
| Async indexing | Keeps upload fast and durable | Requires jobs and worker operational logic |
| PostgreSQL BYTEA for v1 files | Simple, transactional, fewer services | DB bloat for large files |
| Pinecone vector DB | Managed semantic retrieval | External dependency and cost |
| PostgreSQL FTS | Exact identifier/keyword search | Needs index tuning |
| RRF fusion | Combines dense and lexical rankings safely | Rank-based, not score-calibrated |
| Vertex reranking | Higher precision context | Extra latency and cost |
| Gemini grounded generation | Strong managed LLM | Must guard against hallucination |
| Provider interfaces | Testability and vendor isolation | More upfront structure |
| OpenTelemetry | Production debugging | Requires trace hygiene |

