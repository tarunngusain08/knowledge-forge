# UI and Backend Quality Guide

This document records the quality bar for the Knowledge Forge product UI and Go backend. It complements the HLD, LLD, sequence, use-case, and database design docs by turning best practices into concrete implementation rules.

## Quality Goals

- Keep the UI operational and evidence-focused, not decorative.
- Keep backend handlers thin: validate input, call services, return typed JSON.
- Keep business logic behind internal services and provider interfaces.
- Return safe, predictable API responses that the UI can parse cleanly.
- Preserve observability without logging full prompts, document text, or secrets.
- Make every user-visible claim traceable to retrieved evidence and repository provenance.

## UI Best Practices

- Use a task-first layout: repository import, indexing, Q&A, evidence, plan, impact, and developer tools.
- Keep state explicit with separate values for token, repository, ingestion job, answer, plan, impact, trace, busy state, and error state.
- Disable actions while their request is in flight.
- Parse API error envelopes so users see `invalid JSON body` instead of raw JSON strings.
- Send `Content-Type: application/json` only when a body exists.
- Treat empty `204` responses as valid success responses.
- Keep benchmark/debug views under Developer Tools instead of making them the primary workflow.
- Keep evidence and citations close to the answer so the product remains grounded.

## Backend Best Practices

- Use Chi middleware for request IDs, real IPs, panic recovery, timeouts, logging, and tracing.
- Decode JSON through a shared helper with:
  - 1 MB body limit.
  - unknown field rejection.
  - multiple JSON value rejection.
  - optional-body support where the endpoint allows an empty body.
- Return JSON responses with `application/json; charset=utf-8`.
- Set `X-Content-Type-Options: nosniff`.
- Keep internal business logic out of HTTP handlers.
- Keep LangChainGo, Vertex, Pinecone, and external SDKs behind internal provider interfaces.
- Keep repository answers tied to repository ID, branch, snapshot, commit SHA, retrieved chunks, retrieval config, and model versions.
- Avoid tracing full prompts or full document/code contents by default.

## HLD: UI and Backend Boundary

```mermaid
flowchart LR
    User["Developer / Interview Demo User"]
    UI["React/Vite UI"]
    API["Go Chi API"]
    Auth["JWT Auth"]
    RepoSvc["Repository Service"]
    CodeQA["Code QA Service"]
    Retrieval["Retriever"]
    Store["PostgreSQL"]
    Vector["Pinecone"]
    Models["Vertex AI / Gemini"]
    Trace["OpenTelemetry + JSON Logs"]

    User --> UI
    UI --> API
    API --> Auth
    API --> RepoSvc
    API --> CodeQA
    CodeQA --> Retrieval
    Retrieval --> Store
    Retrieval --> Vector
    CodeQA --> Models
    API --> Trace
    CodeQA --> Trace
```

## LLD: Backend Request Handling

```mermaid
flowchart TB
    Request["HTTP Request"]
    Middleware["RequestID + RealIP + Recoverer + Timeout + OTel"]
    AuthGuard["Auth Middleware"]
    Handler["HTTP Handler"]
    Decode["Strict JSON Decode"]
    Validate["Endpoint Validation"]
    Service["Domain Service"]
    Providers["Provider Interfaces"]
    Response["writeJSON / writeError"]

    Request --> Middleware
    Middleware --> AuthGuard
    AuthGuard --> Handler
    Handler --> Decode
    Decode --> Validate
    Validate --> Service
    Service --> Providers
    Providers --> Service
    Service --> Response
```

## Frontend Flow

```mermaid
flowchart TB
    Login["Login"]
    Repo["Save Repository"]
    Index["Queue / Process Indexing"]
    Question["Ask Question"]
    Evidence["Inspect Evidence"]
    Plan["Generate Plan"]
    Impact["Analyze Impact"]
    Trace["Load Trace"]
    Feedback["Submit Feedback"]

    Login --> Repo
    Repo --> Index
    Index --> Question
    Question --> Evidence
    Evidence --> Plan
    Evidence --> Impact
    Evidence --> Trace
    Evidence --> Feedback
```

## Sequence: Repository Q&A

```mermaid
sequenceDiagram
    actor User
    participant UI as React UI
    participant API as Go API
    participant Auth as Auth Middleware
    participant CodeQA as CodeQA Service
    participant Retrieval as Retriever
    participant Store as PostgreSQL
    participant Vector as Pinecone
    participant LLM as Gemini

    User->>UI: Ask "Where is auth implemented?"
    UI->>API: POST /v1/ask
    API->>Auth: Validate JWT
    Auth-->>API: User context
    API->>CodeQA: Ask request
    CodeQA->>Retrieval: Retrieve evidence
    Retrieval->>Store: Lexical/snapshot metadata lookup
    Retrieval->>Vector: Dense search scoped by snapshot
    Retrieval-->>CodeQA: Ranked evidence
    CodeQA->>LLM: Grounded prompt with citations
    LLM-->>CodeQA: Answer
    CodeQA-->>API: Answer + citations + trace ID
    API-->>UI: JSON response
    UI-->>User: Answer with evidence
```

## Use-Case Diagram

```mermaid
flowchart LR
    User["Developer"]
    UC1["Import repository"]
    UC2["Index snapshot"]
    UC3["Ask code question"]
    UC4["Inspect citations"]
    UC5["Generate implementation plan"]
    UC6["Analyze impact"]
    UC7["Inspect trace"]
    UC8["Submit quality feedback"]

    User --> UC1
    User --> UC2
    User --> UC3
    UC3 --> UC4
    UC4 --> UC5
    UC4 --> UC6
    User --> UC7
    User --> UC8
```

## API Response Contract

Successful responses:

```json
{
  "answer": "The authentication flow is implemented in ...",
  "citations": [],
  "trace_id": "..."
}
```

Error responses:

```json
{
  "error": "invalid JSON body"
}
```

Frontend handling rules:

- Prefer structured `error` or `message` fields.
- Fall back to response text for non-JSON failures.
- Fall back to HTTP status if the response has no body.
- Preserve HTTP status and request path on `ApiError`.

## Quality Gates

Before merging UI/backend changes, run:

```bash
go test ./...
go vet ./...
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py eval-runner/repo_benchmark_runner.py
cd ui/web && npm test && npm run lint && npm run build
docker compose config
```

When Docker is running, also run:

```bash
docker compose up --build -d
docker compose ps
docker compose down
```

## Review Checklist

- Does the UI show evidence beside answers?
- Are buttons disabled while the matching request is in flight?
- Are API errors human-readable?
- Are JSON request bodies bounded and strict?
- Are handlers free of provider SDK logic?
- Are repository answers tied to commit SHA and trace ID?
- Are logs/traces useful without leaking sensitive content?
- Are diagrams and docs updated when architecture changes?
