# Use Cases and Sequence Diagrams

## Actors

```mermaid
flowchart LR
  User[Authenticated User]
  Admin[Admin User]
  API[Knowledge Forge API]
  Worker[Indexing Worker]
  PG[(PostgreSQL)]
  Pinecone[(Pinecone)]
  Vertex[Vertex AI]
  UI[React UI]

  User --> UI
  Admin --> UI
  UI --> API
  API --> PG
  Worker --> PG
  API --> Pinecone
  Worker --> Pinecone
  API --> Vertex
  Worker --> Vertex
```

## Use Case 1: Login

Goal:

- User receives a JWT for protected API calls.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant Auth as Auth Service
  participant DB as PostgreSQL

  User->>UI: Enter email/password
  UI->>API: POST /auth/login
  API->>Auth: Validate credentials
  Auth->>DB: Find user by email
  DB-->>Auth: User + password_hash
  Auth->>Auth: Compare password hash
  Auth-->>API: JWT
  API-->>UI: token + user profile
  UI-->>User: Logged in
```

## Use Case 2: Upload Document

Goal:

- Store the file and enqueue indexing without blocking on embeddings.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant Doc as Document Service
  participant DB as PostgreSQL

  User->>UI: Upload PDF/Markdown/Text
  UI->>API: POST /documents
  API->>Doc: Validate upload
  Doc->>Doc: MIME, extension, size, SHA-256
  Doc->>DB: Insert document raw bytes + metadata
  Doc->>DB: Insert indexing_jobs row
  DB-->>Doc: document_id + job_id
  Doc-->>API: Upload accepted
  API-->>UI: 201 Created
  UI-->>User: Document uploaded, indexing queued
```

## Use Case 3: Index Document

Goal:

- Convert raw document bytes into searchable chunks and vectors.

```mermaid
sequenceDiagram
  participant Worker as Indexing Worker
  participant DB as PostgreSQL
  participant Extractor as Document Extractor
  participant Chunker as Chunking Provider
  participant Vertex as Vertex Embeddings
  participant Pinecone as Pinecone

  Worker->>DB: Claim queued indexing job
  DB-->>Worker: job + document
  Worker->>DB: Mark job running, document indexing
  Worker->>Extractor: Extract text from raw bytes
  Extractor-->>Worker: Plain text
  Worker->>Chunker: Split text into chunks
  Chunker-->>Worker: Chunk list
  Worker->>DB: Insert chunks
  Worker->>Vertex: Embed chunks as RETRIEVAL_DOCUMENT
  Vertex-->>Worker: Embedding vectors
  Worker->>Pinecone: Upsert vector records
  Pinecone-->>Worker: Upsert success
  Worker->>DB: Mark job succeeded, document indexed
```

## Use Case 4: Ask a Question

Goal:

- Answer from trusted retrieved context with citations.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant Chat as Chat Service
  participant Retrieval as Retrieval Service
  participant Vertex as Vertex AI
  participant Pinecone as Pinecone
  participant DB as PostgreSQL
  participant Gemini as Gemini

  User->>UI: Ask "What is our PTO policy?"
  UI->>API: POST /chat/sessions/{id}/messages
  API->>Chat: Add user message
  Chat->>DB: Load chat history
  Chat->>Gemini: Rewrite question if needed
  Gemini-->>Chat: Standalone query
  Chat->>Retrieval: Retrieve(query, topK)
  Retrieval->>Vertex: Embed query as RETRIEVAL_QUERY
  Vertex-->>Retrieval: Query vector
  Retrieval->>Pinecone: Dense search top 20
  Pinecone-->>Retrieval: Dense hits
  Retrieval->>DB: PostgreSQL FTS top 20
  DB-->>Retrieval: Lexical hits
  Retrieval->>Retrieval: RRF fusion
  Retrieval->>Vertex: Rerank fused hits
  Vertex-->>Retrieval: Reranked top chunks
  Retrieval-->>Chat: Context + trace data
  Chat->>Gemini: Prompt with question + context
  Gemini-->>Chat: Grounded answer
  Chat->>DB: Store assistant message, citations, trace
  Chat-->>API: Answer + citations
  API-->>UI: Response
  UI-->>User: Answer with citations
```

## Use Case 5: Debug Retrieval

Goal:

- Understand why a question produced a certain answer.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant DB as PostgreSQL

  User->>UI: Open debug view
  UI->>API: GET /debug/retrieval
  API->>DB: Read retrieval_traces
  DB-->>API: dense, lexical, fused, reranked hits
  API-->>UI: Trace payload
  UI-->>User: Show retrieval pipeline details
```

## Use Case 6: Run Evaluation

Goal:

- Measure retrieval and generation quality.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant Eval as Evaluation Service
  participant DB as PostgreSQL
  participant Ragas as Python Ragas Runner

  User->>UI: Start evaluation
  UI->>API: POST /eval/runs
  API->>Eval: Create eval run
  Eval->>DB: Insert eval_runs row
  Eval->>Eval: Compute retrieval metrics
  Eval->>Ragas: Optional JSONL generation eval
  Ragas-->>Eval: Faithfulness/relevancy metrics
  Eval->>DB: Store metrics and status
  API-->>UI: eval_run_id
  UI->>API: GET /eval/runs/{id}
  API-->>UI: Metrics
```

## Use Case 7: Delete Document

Goal:

- Remove a document and prevent stale citations/retrieval.

```mermaid
sequenceDiagram
  actor User
  participant API as Go API
  participant Doc as Document Service
  participant DB as PostgreSQL
  participant Pinecone as Pinecone

  User->>API: DELETE /documents/{id}
  API->>Doc: Delete document
  Doc->>Pinecone: Delete vectors for document
  Pinecone-->>Doc: Deleted
  Doc->>DB: Mark document deleted / cascade chunks
  DB-->>Doc: Success
  Doc-->>API: Deleted
  API-->>User: 204 No Content
```

## Use Case 8: Generate Repository Plan or Impact Analysis

Goal:

- Turn cited repository evidence into a read-only implementation plan or impact
  analysis without inventing unsupported changes.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant CodeQA as Code Q&A Service
  participant Retrieval as Repository Retriever
  participant Gemini as Gemini
  participant DB as PostgreSQL

  User->>UI: Click Generate Plan or Analyze Impact
  UI->>API: POST /v1/plans or POST /v1/impact
  API->>CodeQA: Validate user/repository request
  CodeQA->>Retrieval: Retrieve cited repository evidence
  Retrieval->>DB: Hydrate chunks and provenance
  DB-->>Retrieval: Files, chunks, line ranges, commit SHA
  Retrieval-->>CodeQA: Evidence hits
  CodeQA->>Gemini: Grounded workflow prompt with untrusted context
  Gemini-->>CodeQA: Draft plan or impact analysis
  CodeQA->>CodeQA: Add evidence-derived confidence and missing context
  CodeQA-->>API: Structured workflow response
  API-->>UI: Evidence, plan/impact, confidence
  UI-->>User: Show cited workflow output
```

## Use Case 9: Generate Repository Deep-Dive Report

Goal:

- Produce a downloadable, cited repository due-diligence report without running
  a separate autonomous agent workflow.

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go API
  participant CodeQA as Code Q&A Service
  participant Retrieval as Repository Retriever
  participant Gemini as Gemini
  participant DB as PostgreSQL

  User->>UI: Click Generate Deep-Dive Report
  UI->>API: POST /v1/reports/deep-dive
  API->>CodeQA: Validate user/repository request
  CodeQA->>Retrieval: Run shared broad evidence pass
  Retrieval->>DB: Hydrate chunks and snapshot provenance
  DB-->>Retrieval: Files, chunks, line ranges, commit SHA
  Retrieval-->>CodeQA: Shared evidence hits
  CodeQA->>Retrieval: Run up to four targeted section retrievals
  Retrieval-->>CodeQA: Section evidence hits
  CodeQA->>Gemini: Grounded report prompts with untrusted context
  Gemini-->>CodeQA: Cited section answers
  CodeQA->>CodeQA: Add missing context and evidence quality
  CodeQA-->>API: Report JSON and Markdown export
  API-->>UI: Deep-dive report, citations, trace IDs
  UI-->>User: Show report and download Markdown
```

## Use Case Summary

| Use Case | Primary Components | Main Risk |
|---|---|---|
| Login | API, Auth, PostgreSQL | Bad secret/password handling |
| Upload | API, Document Service, PostgreSQL | Unsafe file ingestion |
| Index | Worker, Vertex, Pinecone, PostgreSQL | Provider failures/rate limits |
| Ask | Chat, Retrieval, Vertex, Pinecone, Gemini | Bad retrieval or hallucination |
| Debug | API, retrieval_traces | Missing observability |
| Eval | Evaluation Service, Ragas | Misleading quality metrics |
| Delete | Document Service, Pinecone, PostgreSQL | Stale vectors |
| Plan/Impact | React UI, Code Q&A, Repository Retriever, Gemini | Unsupported recommendations |
| Deep-Dive Report | React UI, Code Q&A, Repository Retriever, Gemini | Expensive or unsupported report claims |
