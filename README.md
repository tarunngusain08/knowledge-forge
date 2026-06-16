# Knowledge Forge

Production-style evidence-grounded knowledge assistant built with Go,
PostgreSQL, Pinecone, Vertex AI, and LangChainGo behind internal provider
interfaces. The original company-document RAG flow remains available, and the
current milestone adds a focused repository-intelligence MVP.

## North-Star Workflow

```text
Index repository
-> Ask architecture/code question
-> Inspect cited evidence
-> Generate implementation plan
-> Generate impact analysis
```

The current repository-intelligence path can import/index one repository
snapshot, answer repository questions with cited file evidence, classify the
query, choose an adaptive retrieval budget, assemble context under a token
budget, and persist answer provenance for debugging and evaluation. Planning and
impact analysis are intentionally deferred.

## Target Retrieval Flow

```text
Question
-> Question Rewriter
-> Vertex Query Embedding
-> Pinecone Dense Retrieval
+
PostgreSQL FTS Retrieval
-> Reciprocal Rank Fusion
-> Vertex Ranking API
-> Gemini
-> Grounded Response + Citations
```

## Repository Intelligence MVP

Repository model:

```text
Repository
└── Branch
    └── Snapshot(commit SHA)
        ├── Files
        ├── Chunks
        ├── Symbols
        └── Graph
```

Repository Q&A retrieval contract:

```text
repository question
-> query classification
-> adaptive retrieval budget
-> Vertex/mock query embedding
-> Pinecone dense retrieval scoped to repository snapshot
-> gated reranking
-> context assembly under token budget
-> Gemini/mock grounded answer
-> citations with repository, branch, commit SHA, file path, and line range
```

The MVP includes repository registration, safe file walking, code chunking,
embedding/upsert, dense retrieval, retrieval traces, and worker/API job
processing. Phase 14 adds adaptive query policy, context compression, retrieved
chunk provenance, stage contribution tracking, and estimated generation cost in
repository retrieval traces. It does not enable graph retrieval, implementation
planning, impact analysis, PR review, diagrams, multi-repo intelligence, or
repository memory.

## Local Development

```bash
cp .env.example .env
make tidy
make migrate-up
make test
docker compose up --build
```

The API exposes `GET /healthz` on port `8080`.

The Streamlit demo UI runs on port `8501` when using Docker Compose.

## API Surface

- `POST /auth/login`
- `GET /me`
- `POST /documents`
- `GET /documents`
- `GET /documents/{id}`
- `DELETE /documents/{id}`
- `POST /chat/sessions`
- `GET /chat/sessions/{id}`
- `POST /chat/sessions/{id}/messages`
- `GET /debug/retrieval`
- `POST /eval/runs`
- `GET /eval/runs/{id}`
- `POST /internal/jobs/{job_id}/process`
- `POST /v1/repositories`
- `GET /v1/repositories/{repository_id}`
- `POST /v1/repositories/{repository_id}/ingestions`
- `GET /v1/ingestions/{job_id}`
- `POST /v1/ask`
- `GET /v1/retrieval-traces/{trace_id}`
- `POST /internal/repository-jobs/{job_id}/process`

## Validation

```bash
make test
make vet
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
```

Repository benchmark fixture:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl \
  --output /tmp/repo-benchmark.json
```

Compare a candidate run against a saved baseline:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input /tmp/knowledge-forge-candidate.jsonl \
  --baseline-input /tmp/naive-semantic-baseline.jsonl \
  --output /tmp/repo-benchmark-comparison.json
```

## Documentation

- [Docs Index](docs/README.md)
- [FR, NFR, and Scale Estimation](docs/01-fr-nfr-scale-estimation.md)
- [HLD and Component Design](docs/02-hld-component-design.md)
- [Use Cases and Sequence Diagrams](docs/03-usecases-sequence-diagrams.md)
- [LLD](docs/04-lld.md)
- [DB Design and Schema](docs/05-db-design-schema.md)
- [Architecture](docs/architecture.md)
- [Implementation Plan](docs/implementation-plan.md)
- [Deployment](deploy/README.md)
- [Cloud Run Deployment](deploy/cloud-run.md)
- [Storage Notes](docs/storage.md)
- [Evaluation](docs/evaluation.md)
- [Future Multi-Tenancy](docs/multitenancy.md)
