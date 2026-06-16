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

- `tgusain/m13-repository-evaluation-benchmarks`: benchmark corpus, naive
  semantic baseline, failure benchmarks, and complexity elimination gates.
- `tgusain/m14-adaptive-retrieval-cost-control`: query classification,
  retrieval budgets, context compression, answer provenance, and cost controls.
- `tgusain/m15-product-experience`: React/Vite product UI and focused demo mode.
- `tgusain/m16-planning-impact-analysis`: benchmarked implementation planning
  and impact analysis with evidence-derived confidence.

Every milestone must leave the repository buildable and include relevant tests
and documentation updates.
