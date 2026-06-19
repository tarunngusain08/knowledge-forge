# Knowledge Forge Portfolio Overview

Knowledge Forge is an evidence-grounded repository intelligence platform built
as a senior backend and AI engineering portfolio project.

## One-Sentence Story

Knowledge Forge indexes a codebase, retrieves relevant evidence, answers
architecture questions with citations, validates retrieval quality, and
generates grounded Deep-Dive Reports, implementation plans, and impact analyses.

## Why It Is Different From A Basic RAG Demo

Most RAG demos stop at:

```text
upload docs
-> ask question
-> get answer
```

Knowledge Forge adds production concerns:

- Go backend with clean provider boundaries
- async ingestion and worker processing
- hybrid document retrieval
- repository snapshot provenance
- cited file and line evidence
- Deep-Dive Reports
- refusal behavior for unsupported questions
- acceptance gates for known failure modes
- OpenTelemetry, structured logs, and cost tracking
- deployment docs for GCP services

## Interview Narrative

The project demonstrates three things:

1. Backend engineering: Go services, SQL migrations, pgx/sqlc, worker flows,
   structured APIs, and deployable containers.
2. AI engineering: embeddings, vector retrieval, reranking, grounding,
   citations, RAG evaluation, and refusal handling.
3. Engineering maturity: provider abstractions, validation gates, root-cause
   closure, evidence packages, and explicit roadmap boundaries.

## Demo Flow

```text
Open Knowledge Forge
-> import repository
-> generate Deep-Dive Report
-> inspect evidence
-> ask a repository question
-> view citations and trace
-> generate implementation plan or impact analysis
-> show acceptance proof
```

## Current Validation Status

Phase 17 is accepted:

```text
6/6 acceptance gates pass
0 evaluator issues
```

See:

- [Phase 17 Validation Proof](../proof/phase17-validation.md)
- [Readiness Scorecard](../readiness-scorecard.md)

## Best Discussion Topics

- Why RAG is used instead of fine-tuning for repository facts.
- Why evidence support is stricter than file existence.
- Why unsupported claims are refused.
- Why provider interfaces keep LangChainGo, Vertex AI, and Pinecone outside core
  business logic.
- Why Deep-Dive Reports use a shared evidence pass before targeted retrieval.
- How Phase 18 and 18.5 benchmark proof prevented premature static
  intelligence or graph retrieval work.

## Known Boundaries

Knowledge Forge is not currently:

- an autonomous code generator
- a PR review bot
- a multi-repository intelligence graph
- a full static-analysis engine
- a production SaaS multi-tenant platform

Those are intentionally outside the current validated scope.
