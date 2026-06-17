# ADR-005: Generate Deep-Dive Reports On Demand

## Status

Accepted

## Context

Knowledge Forge now produces repository deep-dive reports that summarize a
repository snapshot with cited evidence, evidence quality, and missing context.
The report is intended to be an interview/demo artifact and a due-diligence
view over an already indexed repository snapshot.

## Decision

Deep-dive reports are generated on demand from the current repository snapshot,
retrieval pipeline, and model configuration.

## Consequences

- Users always see a report generated from current indexed evidence.
- The implementation avoids a new report lifecycle, invalidation model, and
  database tables in v1.
- The durable audit trail remains the underlying retrieval traces and snapshot
  provenance.
- Report generation cost and latency happen at request time.
