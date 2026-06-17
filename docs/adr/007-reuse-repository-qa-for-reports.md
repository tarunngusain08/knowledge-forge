# ADR-007: Reuse Repository Q&A For Report Sections

## Status

Accepted

## Context

Knowledge Forge already has a repository Q&A path with query rewriting,
retrieval policy, dense retrieval, reranking, context assembly, grounded Gemini
generation, citations, retrieval traces, cost estimates, and provenance.

## Decision

Deep-dive report sections reuse the repository Q&A path instead of introducing
a separate agent or report-specific retrieval system.

## Consequences

- Reports inherit existing citations, prompt-injection isolation, traces, and
  provenance.
- Business logic continues to depend on internal provider interfaces.
- Report behavior remains measurable by the same retrieval/debug machinery.
- Future report quality improvements should improve the shared repository Q&A
  path rather than creating parallel logic.
