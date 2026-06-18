# Acceptance Validation Framework

This directory implements the Phase 17 acceptance-redesign package as executable validation.

It is intentionally validation-only:

- It does not change retrieval.
- It does not change report generation.
- It does not change architecture extraction.
- It does not change the support gate.

## Run

```bash
make validate-acceptance
```

The command validates a passing candidate artifact, writes markdown reports under `eval-runner/acceptance/reports`, and runs validator self-tests.

## Candidate Contract

Candidate outputs live under `eval-runner/acceptance/candidates`.

A candidate must provide:

- refusal decisions and support-gate traces
- answer relevance evidence groups and fact support
- architecture layer evidence
- metric integrity metadata
- benchmark label completeness metadata

The validator exits nonzero when the candidate repeats the red-team failures captured in `red-team-repeat-candidate.json`.

## Gates

1. Refusal Decision Matrix
2. Answer Relevance
3. Architecture Evidence
4. Metric Integrity
5. Benchmark Label Completeness
6. Adversarial Benchmark
