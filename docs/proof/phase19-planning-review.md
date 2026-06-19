# Phase 19 Planning Review

## Decision

Selected Next Direction:

```text
Proceed with Larger Corpus Expansion
```

Decision Confidence:

```text
High
```

This decision is based on current evidence. It is not a permanent roadmap
commitment. Larger Corpus Expansion is the best-supported next experiment after
Phases 18 through 18.8, and the recommendation should be revisited after the
next corpus-expansion benchmark.

## Current Project State

| Milestone | Status | Evidence |
| --- | --- | --- |
| Phase 17 Product Conformance | Complete | [Phase 17 Validation](phase17-validation.md) |
| Phase 18 Benchmark Proof | Complete, Partially Proven | [Phase 18 Benchmark Proof](../evaluations/phase18-benchmark-proof.md) |
| Phase 18.5 Multi-Corpus Benchmark | Complete, Generalized, Moderately Stable | [Phase 18.5 Benchmark Report](../../eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md) |
| Phase 18.6 Tenant Isolation | Complete | [Phase 18.6 Security Remediation](phase18-6-security-remediation.md) |
| Phase 18.7 Release Readiness | READY | [Phase 18.7 Release Readiness](phase18-7-release-readiness.md) |
| Phase 18.8 Security Hardening | Complete | [Phase 18.8 Security Hardening](phase18-8-security-hardening.md) |

Phase 18.8 is merged into `main` as:

```text
2c39902 fix(security): harden phase 18.8 findings
```

## Evidence Summary

Phase 18 established measurable repository-intelligence advantage over keyword
and retrieval-only baselines in architecture, dependency/impact, and grounding
categories.

Phase 18.5 tested whether those gains generalized beyond the synthetic monolith.
The result was:

```text
Generalized
Moderately Stable
```

The Phase 18.5 benchmark showed:

- Helm: 20/20 correct and improved over both baselines.
- OpenTelemetry Collector: 18/20 correct and improved over both baselines.
- Cross-corpus stability: Moderately Stable.

The scope limitation remains important: this evidence applies to infrastructure,
platform, and developer-tooling repositories represented by the benchmark. It
must not be generalized to all repository types yet.

Phase 18.6 and Phase 18.8 closed security risk areas before additional roadmap
investment. Phase 18.7 marked release readiness as `READY`.

## Option Comparison

| Option | Recommendation | Reason |
| --- | --- | --- |
| Larger Corpus Expansion | Proceed | Best-supported, lowest-cost way to test external validity across more repository families. |
| Repository Structure Indexing | Investigate | May help if future failures cluster around architecture/dependency navigation, but existing failures are narrow and not yet proven as a bottleneck. |
| Static Code Intelligence | Investigate | Not justified until symbol/reference failures dominate or retrieval recall is high while reasoning accuracy remains low. |
| Graph Retrieval | Reject | Graph-specific failures do not dominate measured results. |
| Stop / Maintain Current Architecture | Reject for now | Leaves a valuable external-validity question unanswered. |

## Decision Rationale

Phase 18 established measurable repository-intelligence advantage.

Phase 18.5 demonstrated cross-corpus generalization.

Phase 18.6 and Phase 18.8 closed security risk areas.

Phase 18.7 established release readiness.

The largest remaining uncertainty is external validity across additional
repository families. Larger Corpus Expansion answers that uncertainty at lower
cost than introducing new retrieval architecture.

Current evidence does not demonstrate that Repository Structure Indexing or
Static Code Intelligence are primary bottlenecks. They remain investigation
candidates, not selected implementation work.

Graph Retrieval remains rejected until graph-specific failures dominate measured
results.

## Exit Criteria For Larger Corpus Expansion

The corpus-expansion effort should conclude when:

- At least three additional repository families are benchmarked.
- Cross-corpus stability is recomputed.
- A new recommendation is made between:
  - Maintain Current Architecture
  - Repository Structure Indexing
  - Static Code Intelligence
- No retrieval architecture changes occur during corpus expansion.
- A new roadmap review is held after corpus expansion.

## Selected Next Implementation Direction

The next implementation branch should focus only on Larger Corpus Expansion.

It should not add:

- repository structure indexing
- static code intelligence
- graph retrieval
- new retrieval architecture
- security or validator changes unrelated to corpus expansion

Human review is required before that implementation branch begins.
