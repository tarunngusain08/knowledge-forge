# Implementation Plan

Knowledge Forge has completed the document-RAG foundation through `m11` and is
now extending into repository intelligence with `tgusain/` feature branches.

## Completed Foundation

1. `tgusain/m00-project-foundation`
2. `tgusain/m01-database-auth`
3. `tgusain/m02-provider-abstractions`
4. `tgusain/m03-document-ingestion`
5. `tgusain/m04-indexing-worker`
6. `tgusain/m05-hybrid-retrieval`
7. `tgusain/m06-rerank-generate-citations`
8. `tgusain/m07-chat-memory-debug`
9. `tgusain/m08-observability-costs`
10. `tgusain/m09-evaluation`
11. `tgusain/m10-streamlit-ui`
12. `tgusain/m11-cloud-deploy-docs`

## North-Star Workflow

```text
Index repository
-> Ask architecture/code question
-> Inspect cited evidence
-> Generate implementation plan
-> Generate impact analysis
```

Every new feature must strengthen this workflow or remain disabled until
benchmarks prove value.

## Phase 12: Repository Intelligence MVP

Branch: `tgusain/m12-repository-intelligence-mvp`

Implemented scope:

- Register one repository from a local path or Git remote.
- Create repository ingestion jobs.
- Process repository jobs through the API or worker.
- Resolve one branch to an immutable commit SHA snapshot.
- Safely walk supported text/code files.
- Chunk files with line ranges.
- Store repository files, chunks, symbols table placeholders, snapshots, Git
  commits, and retrieval traces in PostgreSQL.
- Embed chunks and upsert Pinecone/mock vectors with repository/snapshot
  metadata.
- Answer repository questions through dense retrieval scoped to the snapshot.
- Return citations with repository ID, branch, commit SHA, file path, line range,
  and excerpt.

Explicit non-goals for Phase 12:

- impact analysis
- implementation planning
- PR review
- architecture diagrams
- benchmark dashboard
- multi-repo intelligence
- autonomous agents
- graph retrieval enabled by default
- repository memory system

Readiness checklist before Phase 13:

```text
Repository can be imported
Repository can be indexed
Question returns cited answer
Commit SHA is attached
Dense retrieval works
Basic safety controls work
Tests pass
Benchmark runner can call repo Q&A
```

## Next Phases

- `tgusain/m13-repository-evaluation-benchmarks`: repository benchmark metrics,
  synthetic enterprise monolith corpus, failure benchmarks, live/offline
  benchmark runner, and complexity elimination gates.
- `tgusain/m14-adaptive-retrieval-cost-control`: query classification,
  retrieval budgets, context compression, answer provenance, and cost controls.
- `tgusain/m15-product-experience`: React/Vite product UI and focused demo mode.
- `tgusain/m16-planning-impact-analysis`: benchmarked implementation planning
  and impact analysis with evidence-derived confidence.

Every milestone must leave the repository buildable and include relevant tests
and documentation updates.

## Phase 13: Repository Evaluation and Complexity Review

Branch: `tgusain/m13-repository-evaluation-benchmarks`

Implemented scope:

- Repository-specific Go metrics for file coverage, symbol coverage,
  line-range accuracy, citation coverage, evidence coverage, refusal accuracy,
  latency, and cost.
- Question-level comparison output: improved, unchanged, degraded.
- Python benchmark runner that can score saved result JSONL or call `/v1/ask`
  against a live API.
- Synthetic Enterprise Monolith fixture with auth, billing, notifications,
  API wiring, interfaces, and tests.
- Failure labels for unsupported questions, deleted symbols, wrong-domain
  questions, and prompt-injection attempts.
- Evaluation documentation with component decision thresholds.

Complexity gates:

- Keep symbol retrieval only if expected file/symbol coverage improves by at
  least 10%.
- Enable graph retrieval only if Recall@K or file coverage improves by at least
  10% without increasing latency by more than 25%.
- Keep reranking only if MRR improves by at least 5% or faithfulness improves
  meaningfully without unacceptable cost.

## Phase 14: Adaptive Retrieval, Context Compression, and Cost Control

Branch: `tgusain/m14-adaptive-retrieval-cost-control`

Implemented scope:

- Query classification for exact lookups, architecture questions,
  implementation questions, impact questions, and unsupported/unknown prompts.
- Adaptive retrieval policy with:
  - final `top_k`
  - candidate depth
  - context token budget
  - retrieval path
  - reranker gating and skip reason
- Repository dense retrieval now honors candidate depth separately from final
  answer context size.
- Context assembly collapses adjacent chunks from the same file and enforces a
  token budget before Gemini receives context.
- Repository Q&A defaults to adaptive reranker behavior when the API caller
  omits `reranker_enabled`; explicit `false` still disables reranking.
- Repository retrieval traces now persist query category, retrieval path,
  retrieval config, retrieved chunk IDs, stage contributions, context token
  count, prompt version, generation model, and estimated answer cost.
- Repository Q&A responses include a `provenance` object so demos can explain
  why a retrieval path was chosen and what evidence entered the prompt.

Tradeoffs:

- Phase 14 does not introduce graph retrieval, symbol retrieval, or answer
  caching. Those remain gated by Phase 13 benchmark evidence.
- Cost control is implemented at the retrieval/generation boundary through
  candidate depth, reranker gating, context token budget, and persisted cost
  analytics. Hard per-user and per-repository budget enforcement can be added
  after usage data exists.
