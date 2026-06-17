# ADR-008: Use A Shared Evidence Pass Before Targeted Retrieval

## Status

Accepted

## Context

A deep-dive report has many sections. Running a full RAG pipeline independently
for every section would increase cost, latency, and debugging complexity.

## Decision

Report generation starts with one broad shared evidence pass. The system then
runs a small number of targeted retrievals only for high-value weak sections.
V1 caps targeted follow-up retrievals at four.

## Consequences

- Report generation is cheaper and easier to debug than ten independent RAG
  runs.
- The shared evidence set provides a consistent base for architecture and
  evidence-quality reporting.
- Some lower-priority sections may explicitly report missing context in v1.
- Benchmark results can later decide whether more targeted sections are worth
  the extra cost.
