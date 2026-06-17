# ADR-006: Do Not Persist Deep-Dive Reports In V1

## Status

Accepted

## Context

Deep-dive reports contain generated summaries, citations, and Markdown export
content. Persisting them would require versioning by repository snapshot,
retrieval config, prompt version, and model version.

## Decision

V1 does not persist report documents. The API returns structured JSON and a
Markdown export in the response.

## Consequences

- The feature remains small and reviewable.
- There is no stale report state to manage when repositories are re-indexed.
- Markdown can still be copied or downloaded by the UI.
- If repeated report generation becomes expensive, persistence can be added with
  a clear key: repository snapshot, retrieval config, prompt version, and model
  version.
