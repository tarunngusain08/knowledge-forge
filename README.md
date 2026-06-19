# Knowledge Forge

Knowledge Forge is an evidence-grounded repository intelligence system.

It indexes documents and code, retrieves source evidence, answers questions with
citations, generates repository deep-dive reports, and produces read-only
implementation plans and impact analyses. The project is built to demonstrate
production-grade Go backend engineering and practical GenAI/RAG engineering, not
just a toy chatbot.

## Why It Exists

Most AI code assistants can produce confident answers even when they did not
actually find the evidence. Knowledge Forge is built around a stricter rule:

```text
Claims require evidence.

If evidence is insufficient:
- say what is missing
- refuse unsupported conclusions
- preserve citations and traces for review
```

That makes the system useful for architecture questions, codebase onboarding,
interview demos, and repository due diligence where trust matters more than
fluent prose.

## North-Star Workflow

```text
Index repository
-> Ask architecture or code question
-> Inspect cited evidence
-> Generate Deep-Dive Report
-> Generate implementation plan
-> Generate impact analysis
```

The current validated product supports repository import, repository Q&A,
Deep-Dive Reports, evidence inspection, implementation planning, impact
analysis, structured validation, and traceable acceptance proof.

## What It Can Do

- Index company documents and repository source files.
- Store chunks and metadata in PostgreSQL.
- Use Pinecone for vector retrieval and PostgreSQL FTS for lexical retrieval.
- Use Vertex AI embeddings, Vertex Ranking API, and Gemini through internal
  provider interfaces.
- Answer repository questions with file paths, line ranges, commit SHA, and
  citations.
- Generate on-demand repository Deep-Dive Reports with evidence quality and
  missing-context sections.
- Produce read-only implementation plans and impact analyses grounded in
  retrieved evidence.
- Expose retrieval traces, cost/token accounting, OpenTelemetry instrumentation,
  and acceptance validation.

## Deep-Dive Reports

A Deep-Dive Report is a cited repository due-diligence artifact. It starts with
one shared evidence pass, performs targeted follow-up retrieval only for weak
sections, and returns structured JSON plus Markdown export.

The report focuses on:

- architecture overview
- entry points
- main packages
- authentication flow
- data layer
- external services
- testing strategy
- risk areas
- suggested improvements
- evidence quality
- missing context

Reports are generated on demand in v1. They are not persisted as first-class
database objects; durable review data comes from repository snapshots, citations,
retrieval traces, provenance, and validation outputs.

## Architecture

```text
User
-> React/Vite UI or API
-> Go Chi backend
-> repository/document services
-> retrieval pipeline
-> Pinecone + PostgreSQL
-> Vertex AI / Gemini
-> grounded response + citations + traces
```

Core stack:

- Go, Chi, pgx, sqlc, Goose
- PostgreSQL and PostgreSQL full text search
- Pinecone vector search
- Vertex AI embeddings, ranking, and Gemini
- LangChainGo behind internal interfaces only
- Python Ragas runner for generation evaluation
- React/Vite primary UI and Streamlit fallback
- Cloud Run, Cloud SQL, Secret Manager, Cloud Tasks deployment docs

The detailed architecture lives in [docs/architecture.md](docs/architecture.md),
[docs/02-hld-component-design.md](docs/02-hld-component-design.md), and
[docs/04-lld.md](docs/04-lld.md).

## Repository Intelligence Model

```text
Repository
└── Branch
    └── Snapshot(commit SHA)
        ├── Files
        ├── Chunks
        ├── Symbols
        └── Graph
```

Every repository answer is tied to a repository snapshot. Citations include the
repository, branch, commit SHA, file path, line range, excerpt, and retrieval
metadata where available.

## Retrieval Flow

Document RAG uses the approved hybrid retrieval flow:

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

Repository Q&A uses the repository retrieval contract implemented through the
same evidence-grounded service boundary:

```text
repository question
-> query classification
-> adaptive retrieval budget
-> dense retrieval scoped to repository snapshot
-> optional reranking
-> evidence support gate
-> context assembly under token budget
-> grounded answer, report, plan, or impact output
```

## Phase 17 Validation Result

Phase 17 is accepted.

```text
Gate 1 Refusal Matrix: pass
Gate 2 Answer Relevance: pass
Gate 3 Architecture Evidence: pass
Gate 4 Metric Integrity: pass
Gate 5 Benchmark Label Completeness: pass
Gate 6 Adversarial Benchmark: pass
Evaluator issues: 0
Acceptance validator: pass
```

The proof summary is intentionally consolidated in one place:

- [Phase 17 Validation Proof](docs/proof/phase17-validation.md)

The raw Desktop evidence packages are not copied into this repository. The proof
doc links to the local evidence package path for traceability.

## Run Locally

```bash
cp .env.example .env
make tidy
make migrate-up
make test
docker compose up --build
```

Default local services:

- API: `http://localhost:8080`
- API health check: `GET /healthz`
- React/Vite UI through Docker Compose: port `8501`
- Streamlit fallback: `ui/streamlit`

Real Vertex AI and Pinecone integration tests are environment-gated so the local
test suite can run without cloud credentials.

## Run Validation

Core validation:

```bash
make test
make vet
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
cd ui/web && npm test && npm run lint && npm run build
docker compose config
make validate-acceptance
```

Repository benchmark runner:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl \
  --output /tmp/repo-benchmark.json
```

Acceptance methodology is documented in
[docs/evaluations/acceptance-methodology.md](docs/evaluations/acceptance-methodology.md).

## Main API Surface

- `POST /auth/login`
- `GET /me`
- `POST /documents`
- `GET /documents`
- `GET /debug/retrieval`
- `POST /eval/runs`
- `POST /v1/repositories`
- `POST /v1/repositories/{repository_id}/ingestions`
- `POST /v1/ask`
- `POST /v1/reports/deep-dive`
- `POST /v1/plans`
- `POST /v1/impact`
- `GET /v1/retrieval-traces/{trace_id}`
- `POST /v1/feedback`

## Documentation Map

Start here:

- [Documentation Index](docs/README.md)
- [Phase 17 Validation Proof](docs/proof/phase17-validation.md)
- [Phase 18 Benchmark Proof](docs/evaluations/phase18-benchmark-proof.md)
- [Phase 18.5 Multi-Corpus Benchmark](docs/evaluations/phase18-5-multi-corpus-benchmark.md)
- [Phase 18.6 Security Remediation](docs/proof/phase18-6-security-remediation.md)
- [Phase 18.7 Release Readiness](docs/proof/phase18-7-release-readiness.md)
- [Roadmap](docs/roadmap.md)
- [Readiness Scorecard](docs/readiness-scorecard.md)

Design and implementation:

- [Architecture](docs/architecture.md)
- [HLD and Component Design](docs/02-hld-component-design.md)
- [LLD](docs/04-lld.md)
- [Database Design and Schema](docs/05-db-design-schema.md)
- [Evaluation](docs/evaluation.md)
- [Acceptance Methodology](docs/evaluations/acceptance-methodology.md)
- [Deep-Dive Report Case Study](docs/case-studies/deep-dive-report.md)
- [Portfolio Overview](docs/portfolio/README.md)

Operational notes:

- [Deployment](deploy/README.md)
- [Cloud Run Deployment](deploy/cloud-run.md)
- [Storage Notes](docs/storage.md)
- [Future Multi-Tenancy](docs/multitenancy.md)
- [Git Deliverable Commits](docs/git-deliverable-commits.md)
- [Architecture Decision Records](docs/adr)

## What To Work On Next

The validated next step is Phase 18: Benchmark Proof Pack.

Phase 18 should publish concise benchmark evidence showing how Knowledge Forge
performs against naive semantic retrieval and adversarial repository questions.
Static code intelligence remains a future candidate for Phase 19 and should only
start after Phase 18 identifies a measured weakness.
