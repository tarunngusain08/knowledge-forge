# Roadmap

This roadmap separates implemented and validated work from future candidates.
No item should be marked complete unless it is implemented, validated, and
merged.

## Completed

### Phase 17: Repository Intelligence And Deep-Dive Report Conformance

Status:

```text
Complete
Accepted
Merged
```

Validated result:

```text
6/6 acceptance gates pass
0 evaluator issues
```

Scope:

- repository Q&A conformance
- Deep-Dive Report conformance
- architecture evidence validation
- claim grounding validation
- validator independence proof

Proof:

- [Phase 17 Validation Proof](proof/phase17-validation.md)

### Phase 18: Benchmark Proof Pack

Status:

```text
Complete
Partially Proven
Merged
```

Goal:

Publish concise benchmark evidence that proves where Knowledge Forge improves
over keyword search and retrieval-only baselines and where it does not.

Validated outputs:

- frozen synthetic benchmark corpus
- keyword baseline
- retrieval-only baseline
- Knowledge Forge candidate output
- benchmark JSON
- human-readable benchmark report
- improved/unchanged/degraded question analysis
- latency and cost comparison
- Phase 19 justification decision table

Proof:

- [Phase 18 Benchmark Design](evaluations/phase18-benchmark-design.md)
- [Phase 18 Benchmark Proof](evaluations/phase18-benchmark-proof.md)

Result:

```text
Partially Proven
```

Knowledge Forge materially outperformed keyword and retrieval-only baselines in
architecture, dependency/impact, and grounding categories. Refusal remained
unchanged against the stronger retrieval-only baseline.

### Phase 18.5: Multi-Corpus Benchmark Expansion

Status:

```text
Complete
Generalized within infrastructure/tooling scope
Merged
```

Goal:

Test whether Phase 18 repository-intelligence gains generalize beyond the
synthetic monolith.

Validated outputs:

- Helm curated corpus
- OpenTelemetry Collector curated corpus
- corpus coverage report
- cross-corpus stability analysis
- failure cluster analysis
- benchmark leakage review
- evidence-backed Phase 19 decision table

Proof:

- [Phase 18.5 Multi-Corpus Benchmark](evaluations/phase18-5-multi-corpus-benchmark.md)
- [Phase 18.5 Benchmark Report](../eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md)

Result:

```text
Generalized
Moderately Stable
```

The result applies only to infrastructure, platform, and developer-tooling
repositories represented by the benchmark. It must not be generalized to all
repository types.

### Phase 18.6: Security Remediation And Tenant Isolation

Status:

```text
Complete
Merged
```

Goal:

Validate and remediate concrete tenant-isolation and deployment trust-boundary
findings before Phase 19.

Validated outputs:

- cross-user dense retrieval isolation
- cross-user PostgreSQL FTS isolation
- retrieval trace owner authorization
- deleted-document retrieval revocation test
- repository input guards for local paths and unsafe remotes
- internal worker token enforcement
- security findings disposition table
- benchmark regression guardrail

Proof:

- [Phase 18.6 Security Remediation](proof/phase18-6-security-remediation.md)

Result:

```text
Security remediation complete
P0 tenant-isolation findings fixed
No security blocker report generated
```

### Phase 18.7: Release Readiness Review

Status:

```text
Complete
READY
Merged
```

Goal:

Confirm that a new engineer can understand, run, evaluate, deploy, and trust
Knowledge Forge without reading the full project history.

Proof:

- [Phase 18.7 Release Readiness](proof/phase18-7-release-readiness.md)

Result:

```text
READY
No release-readiness blocker found
```

### Phase 18.8: Security Hardening

Status:

```text
Complete
Merged
```

Goal:

Validate and close remaining concrete medium-severity security issues affecting
tenant boundaries, refusal leakage, document lifecycle integrity, and upload
availability.

Proof:

- [Phase 18.8 Security Hardening](proof/phase18-8-security-hardening.md)

Result:

```text
Security hardening complete
Five findings reproduced, fixed, and regression-tested
Merged as 2c39902
```

### Phase 19 Planning Review And Independent Challenge

Status:

```text
Complete
Docs-only
Merged
```

Goal:

Select exactly one next roadmap direction using Phase 18 through Phase 18.8
evidence, then attempt to overturn that selection through an independent
challenge review.

Proof:

- [Phase 19 Planning Review](proof/phase19-planning-review.md)
- [Independent Roadmap Challenge](proof/independent-roadmap-challenge.md)

Result:

```text
Selected direction: Larger Corpus Expansion
Board decision: APPROVE WITH RESERVATIONS
Challenge result: SURVIVED
Decision confidence: Medium-High
```

The independent challenge review lowered confidence from unconditional `High`
to `Medium-High` because external validity remains the highest-risk assumption.
It did not find evidence strong enough to overturn the selected direction.

## Validated Next

### Phase 19: Larger Corpus Expansion

Status:

```text
Selected Next Direction
Not started
Human review required before implementation
```

Decision:

```text
Proceed with Larger Corpus Expansion
Decision Confidence: Medium-High
```

Reason:

Phase 18 established measurable repository-intelligence advantage. Phase 18.5
showed the advantage was generalized and moderately stable across the synthetic
monolith, Helm, and OpenTelemetry Collector. Phase 19 Planning Review selected
Larger Corpus Expansion, and the Independent Roadmap Challenge approved that
decision with reservations.

The strongest remaining uncertainty is external validity across additional
repository families. Larger Corpus Expansion is the lowest-cost way to test that
uncertainty before introducing new retrieval architecture.

Required boundaries:

- no retrieval architecture changes
- no Repository Structure Indexing implementation
- no Static Code Intelligence implementation
- no Graph Retrieval implementation
- no benchmark label changes after freeze
- no Phase 20 work

Exit criteria:

- at least three additional repository families benchmarked
- cross-corpus stability recomputed
- failure clusters recomputed
- new recommendation made between Maintain Current Architecture, Repository
  Structure Indexing, and Static Code Intelligence
- new roadmap review held after corpus expansion

## Investigation Candidates

These items are not selected implementation work. They remain candidates only if
future benchmark evidence justifies them.

### Repository Structure Indexing

Status:

```text
Candidate
Not started
```

Allowed only if future benchmark failures cluster around architecture,
repository topology, package boundaries, or dependency navigation.

Candidate scope:

- directory/package structure extraction
- entry point and module relationship summaries
- lightweight repository topology evidence

### Static Code Intelligence v1

Status:

```text
Candidate
Not started
```

Allowed only if benchmark evidence shows symbol/reference failures dominate, or
retrieval recall is high while reasoning accuracy remains low.

Candidate scope:

- interfaces
- structs
- constructors
- imports
- calls
- simple implements mapping

Not allowed by default:

- full call graph
- runtime graph
- dependency-injection inference
- reflection tracking
- multi-repo graph traversal

### Graph Retrieval

Status:

```text
Rejected for now
Not started
```

Graph Retrieval remains rejected until graph-specific failures dominate measured
benchmark results. Phase 18.5 found only one multi-hop dependency reasoning row,
which is not enough to justify graph retrieval.

## Not Yet Started

Anything beyond Phase 19 is not started and should not be described as
implemented or validated.

Examples:

- multi-repository intelligence
- GitHub PR ingestion
- GitHub issue creation
- autonomous code changes
- code generation
- PR review workflow
- architecture diagram generation
- repository memory system
- runtime tracing beyond current observability
- reflection-aware call graph
- full SaaS multi-tenancy

## Roadmap Rule

Every future phase must strengthen the North-Star workflow:

```text
Index repository
-> Ask architecture/code question
-> Inspect cited evidence
-> Generate Deep-Dive Report
-> Generate implementation plan
-> Generate impact analysis
```

Features that do not improve this workflow should remain out of scope.
