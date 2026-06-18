# Phase 18.5 Multi-Corpus Benchmark

## Purpose

Phase 18.5 tests whether Knowledge Forge's repository-intelligence gains from
Phase 18 generalize beyond the synthetic enterprise monolith.

This phase is evidence generation only. It does not add product capabilities,
retrieval architecture, graph retrieval, static code intelligence, agents, or
new APIs.

## Scope Limitation

The corpus expansion covers infrastructure, platform, and developer-tooling
repositories:

- Synthetic enterprise monolith
- Helm
- OpenTelemetry Collector

Results must not be generalized to all repository types. They do not prove
performance for messy application repositories, frontend-heavy repositories,
data science repositories, or repositories with poor source organization.

## Corpora

| Corpus | Source | Snapshot | Fixture | Size |
| --- | --- | --- | --- | --- |
| Synthetic enterprise monolith | local fixture | not applicable | `eval-runner/fixtures/synthetic-enterprise-monolith` | 10 Go files, 267 LOC |
| Helm | `https://github.com/helm/helm` | `v3.15.4`, `fa9efb07d9d8debbb4306d72af76a383895aa8c4` | `eval-runner/fixtures/helm` | 44 Go files, 12,066 LOC |
| OpenTelemetry Collector | `https://github.com/open-telemetry/opentelemetry-collector` | `v0.104.0`, `a082f2e439e8f77a9a9503d54d8afea576f2d08c` | `eval-runner/fixtures/otel-collector` | 49 Go files, 7,360 LOC |

Helm and OpenTelemetry Collector are curated subsets, not full vendored
repositories. Each subset remains under the Phase 18.5 cap of 75 source files or
50k LOC.

## Benchmark Corpus

Phase 18.5 uses `eval-runner/benchmarks/phase18_5_multi_corpus.jsonl`.

The existing Phase 18 synthetic rows are copied into the combined benchmark and
tagged with `corpus=synthetic_enterprise_monolith`. The original synthetic
benchmark file remains unchanged.

New rows:

- 20 Helm rows
- 20 OpenTelemetry Collector rows

Distribution per new corpus:

- 7 architecture / implementation
- 6 dependency / impact / testing
- 4 unsupported / refusal / prompt-injection
- 3 Deep-Dive / grounding / architecture evidence

`expected_line_ranges` remain optional.

## Benchmark Freeze

Before any result generation, the corpus must pass
`BENCHMARK-INTEGRITY-REVIEW.md`.

After freeze:

- labels cannot change
- expected facts cannot change
- evidence groups cannot change
- refusal expectations cannot change

unless a documented labeling defect is found. If more than 20% of rows require
rewrite after freeze, Phase 18.5 must stop and produce
`PHASE18_5_BLOCKER_REVIEW.md`.

## Baselines

Result artifacts are generated in this order:

1. `keyword_baseline.jsonl`
2. `retrieval_only_baseline.jsonl`
3. `knowledge_forge_candidate.jsonl`

Baselines must be preserved before Knowledge Forge candidate output is created.
They must not be regenerated after candidate creation.

## Primary Metrics

- Retrieval recall / file coverage
- Evidence recall
- Answerable-question accuracy
- Refusal precision
- Refusal recall
- Grounding coverage

## Secondary Metrics

- MRR
- Citation accuracy
- Latency
- Cost

## Corpus Coverage

`CORPUS-COVERAGE-REPORT.md` reports, for each corpus:

- files indexed
- files benchmarked
- symbols benchmarked
- evidence groups benchmarked
- benchmarked file coverage percentage

Coverage is required so a positive result cannot hide a tiny happy-path slice of
Helm or OpenTelemetry Collector.

## Cross-Corpus Stability

`CROSS-CORPUS-STABILITY.md` reports the best corpus, worst corpus, and metric
range for primary metrics.

Classification:

- `Stable`: range <= 0.10 and no corpus degraded
- `Moderately Stable`: range <= 0.20 and no major category failure
- `Unstable`: range > 0.20 or any major category degraded

## Failure Clusters

`FAILURE-CLUSTER-ANALYSIS.md` groups failures by:

- missing symbol retrieval
- missing architecture evidence
- multi-hop dependency reasoning
- impact analysis
- refusal classification
- grounding gaps
- citation gaps

These clusters are used to make Phase 19 recommendations evidence-backed.

## Outcome Classification

Per corpus/category:

- `IMPROVED`: Knowledge Forge beats both baselines by at least 10% relative
  improvement or at least 3 additional correct rows.
- `UNCHANGED`: no material improvement.
- `DEGRADED`: Knowledge Forge underperforms either baseline.

Overall:

- `Generalized`: both new corpora improve in at least two major understanding
  categories with no major regression.
- `Mixed`: improvement is limited to one corpus or narrow categories.
- `Not Generalized`: no material advantage over baselines.

## Phase 19 Decision Thresholds

| Candidate | Proceed Only If |
| --- | --- |
| Repository Structure Indexing | Architecture/dependency failures are more than 25% of total failures |
| Static Code Intelligence | Retrieval recall is high but answer/reasoning accuracy is low |
| Graph Retrieval | Multi-hop dependency failures dominate the failure clusters |
| Larger Corpus | Results are stable or moderately stable across corpora |

If a threshold is not met, Phase 18.5 must not recommend `Proceed`.

## Result

Phase 18.5 result:

```text
Generalized
```

Within the infrastructure/platform/developer-tooling scope, Knowledge Forge
improved over both keyword and retrieval-only baselines across the two added
non-synthetic corpora:

- Helm: 20/20 correct.
- OpenTelemetry Collector: 18/20 correct.
- Overall: 68/70 correct.
- Cross-corpus stability: `Moderately Stable`.

The result does not prove generalization to all repository types.

## Proof Artifacts

- `BENCHMARK-INTEGRITY-REVIEW.md`
- `BENCHMARK-LEAKAGE-REVIEW.md`
- `CORPUS-COVERAGE-REPORT.md`
- `CROSS-CORPUS-STABILITY.md`
- `FAILURE-CLUSTER-ANALYSIS.md`
- `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`
