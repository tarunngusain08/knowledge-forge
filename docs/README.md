# Knowledge Forge Documentation

This directory is the curated source of truth for Knowledge Forge. It is written
for new contributors, reviewers, interviewers, and future maintainers who need
to understand what the project does, how it works, how it is validated, and what
should happen next without reading raw audit history.

The documentation is intentionally split by purpose:

- `README.md` at the repository root explains the product and how to run it.
- `docs/architecture.md` and the HLD/LLD docs explain how the system is built.
- `docs/proof/` contains concise milestone proof documents.
- `docs/evaluations/` explains benchmark and acceptance methodology.
- `docs/roadmap.md` separates completed work from future candidates.

## Start Here

Read these in order if you are new to the repository:

1. [Project README](../README.md)
   Product overview, architecture summary, local setup, validation commands,
   evidence links, and current roadmap.
2. [Architecture](architecture.md)
   End-to-end system architecture, repository intelligence model, provider
   boundaries, retrieval flow, and deployment shape.
3. [Readiness Scorecard](readiness-scorecard.md)
   Current maturity snapshot across product, validation, benchmark, security,
   and roadmap areas.
4. [Roadmap](roadmap.md)
   What is complete, what is selected next, what is only a candidate, and what
   is explicitly not started.
5. [Documentation Gap Analysis](../DOCS-GAP-ANALYSIS.md)
   What was improved during the documentation refresh and what screenshot or
   demo gaps remain.

## Product And Architecture

- [Functional Requirements, Non-Functional Requirements, and Scale Estimation](01-fr-nfr-scale-estimation.md)
  Defines product capabilities, non-functional expectations, scale assumptions,
  and infrastructure tradeoffs.
- [High-Level Design and Component Design](02-hld-component-design.md)
  Explains the major services, UI/API boundary, provider boundaries, and
  operational components.
- [Use Cases and Sequence Diagrams](03-usecases-sequence-diagrams.md)
  Shows user journeys for auth, upload, retrieval, evaluation, repository Q&A,
  planning, impact analysis, and Deep-Dive Reports.
- [Low-Level Design](04-lld.md)
  Describes interfaces, internal service responsibilities, and implementation
  contracts.
- [Database Design and Schema](05-db-design-schema.md)
  Documents relational entities for users, documents, chunks, repositories,
  snapshots, traces, feedback, and evaluation state.
- [UI and Backend Quality Guide](06-ui-backend-quality.md)
  Captures UI/backend quality expectations and validation commands.
- [Architecture](architecture.md)
  The shortest implementation-level architecture overview.

## Repository Intelligence Concepts

Knowledge Forge has two related product paths:

- document RAG for uploaded company or project documents
- repository intelligence for source-code repositories

Repository intelligence is the current project center of gravity. It indexes
repository snapshots, retrieves source evidence, answers questions with
citations, generates Deep-Dive Reports, and produces read-only implementation
plans and impact analyses.

Key concepts:

- Claims require evidence.
- Repository answers are tied to commit SHA snapshots.
- Evidence includes file path, line range, excerpt, retrieval metadata, and
  citations.
- Unsupported claims should be refused or represented as missing context.
- Planning and impact workflows are read-only; they do not mutate code.

## Evaluation And Proof

The project has accumulated proof in layers. Prefer these documents over raw
Desktop evidence packages or intermediate audit notes.

| Area | Canonical Document | What It Proves |
| --- | --- | --- |
| Acceptance methodology | [Acceptance Methodology](evaluations/acceptance-methodology.md) | How refusal, relevance, architecture evidence, metric integrity, labels, and adversarial gates are validated. |
| Phase 17 conformance | [Phase 17 Validation Proof](proof/phase17-validation.md) | Product conformance reached 6/6 gates, 0 evaluator issues, and accepted evidence. |
| Phase 18 benchmark proof | [Phase 18 Benchmark Proof](evaluations/phase18-benchmark-proof.md) | Knowledge Forge outperformed keyword and retrieval-only baselines in architecture, dependency/impact, and grounding categories on the synthetic monolith. |
| Phase 18.5 generalization | [Phase 18.5 Multi-Corpus Benchmark](evaluations/phase18-5-multi-corpus-benchmark.md) | Results generalized within infrastructure/platform/developer-tooling scope across Helm and OpenTelemetry Collector. |
| Phase 18.6 security | [Phase 18.6 Security Remediation](proof/phase18-6-security-remediation.md) | P0 tenant-isolation and deployment trust-boundary findings were remediated and regression-tested. |
| Phase 18.7 readiness | [Phase 18.7 Release Readiness](proof/phase18-7-release-readiness.md) | Architecture, security, benchmarks, docs, onboarding, and deployment were reviewed as READY. |
| Phase 18.8 security hardening | [Phase 18.8 Security Hardening](proof/phase18-8-security-hardening.md) | Medium-severity security findings were reproduced, fixed, and regression-tested. |
| Phase 19 decision | [Phase 19 Planning Review](proof/phase19-planning-review.md) | Larger Corpus Expansion was selected as the next direction based on current evidence. |
| Independent challenge | [Independent Roadmap Challenge](proof/independent-roadmap-challenge.md) | The Phase 19 decision survived adversarial review, but with reservations around external validity. |

## Visual Evidence

The repository includes small, evidence-backed visual summaries generated from
committed docs and benchmark artifacts:

| Visual | Path | Source Evidence |
| --- | --- | --- |
| Architecture overview | [architecture-overview.png](images/architecture/architecture-overview.png) | `README.md`, `docs/architecture.md`, Phase 18.6/18.8 security proof. |
| Benchmark summary | [benchmark-summary.png](images/benchmarks/benchmark-summary.png) | Phase 18 and Phase 18.5 committed benchmark JSON. |
| Benchmark comparison | [benchmark-comparison.png](images/benchmarks/benchmark-comparison.png) | Phase 18 committed benchmark JSON. |
| Security posture summary | [security-posture-summary.png](images/security/security-posture-summary.png) | Phase 18.6 and Phase 18.8 proof docs. |

UI screenshots are intentionally not committed until a real runnable demo
session is available. See [UI Screenshot Inventory](images/ui/README.md).

## Benchmarks And Validation

Benchmark docs:

- [Evaluation](evaluation.md)
  Defines repository evaluation concepts and metric vocabulary.
- [Phase 18 Benchmark Design](evaluations/phase18-benchmark-design.md)
  Defines benchmark schema, baselines, primary/secondary metrics, freeze rules,
  material improvement, and failure criteria.
- [Phase 18 Benchmark Proof](evaluations/phase18-benchmark-proof.md)
  Presents the synthetic-monolith proof pack.
- [Phase 18.5 Multi-Corpus Benchmark](evaluations/phase18-5-multi-corpus-benchmark.md)
  Presents multi-corpus methodology, scope limitation, coverage, stability,
  failure clusters, and roadmap implications.

Important benchmark limitation:

```text
Current benchmark evidence is strongest for infrastructure, platform, and
developer-tooling repositories. It does not yet prove performance across every
repository type.
```

That limitation is why Phase 19 selects Larger Corpus Expansion instead of
new retrieval architecture.

## Security And Deployment

Security proof:

- [Phase 18.6 Security Remediation](proof/phase18-6-security-remediation.md)
- [Phase 18.8 Security Hardening](proof/phase18-8-security-hardening.md)

Deployment docs:

- [Deployment Overview](../deploy/README.md)
- [Cloud Run Deployment](../deploy/cloud-run.md)
- [Storage Notes](storage.md)
- [Future Multi-Tenancy Notes](multitenancy.md)

Hosted deployments should keep:

- `ALLOW_LOCAL_REPOSITORY_PATHS=false`
- `INTERNAL_WORKER_TOKEN` in Secret Manager
- `ALLOWED_GIT_REMOTE_HOSTS` restricted to approved Git hosts
- Cloud Run ingress and IAM restricted for internal worker calls

## ADRs

Architecture Decision Records capture durable choices that should not be
re-litigated without new evidence.

- [ADR 005: On-Demand Deep-Dive Reports](adr/005-on-demand-deep-dive-reports.md)
- [ADR 006: No Report Persistence v1](adr/006-no-report-persistence-v1.md)
- [ADR 007: Reuse Repository QA For Reports](adr/007-reuse-repository-qa-for-reports.md)
- [ADR 008: Shared Evidence Before Targeted Retrieval](adr/008-shared-evidence-before-targeted-retrieval.md)

## Case Studies, Portfolio, And Retrospectives

- [Deep-Dive Report Case Study](case-studies/deep-dive-report.md)
  Product workflow example for repository due diligence.
- [Portfolio Overview](portfolio/README.md)
  Interview-friendly summary of product capability and engineering rigor.
- [Phase 17 Retrospective](postmortems/phase17-retrospective.md)
  Concise lessons about validation, stop gates, and avoiding audit loops.
- [Repository Health](repository-health.md)
  Discoverability and documentation cleanup recommendations.

## Tooling

- [Git Deliverable Commits](git-deliverable-commits.md)
  Local CLI notes for clean deliverable-based commits.

## Historical Documents

Some documents are retained for history but are not the current source of truth:

- `phase17-hardening-review.md`
  Historical hardening context. Use [Phase 17 Validation Proof](proof/phase17-validation.md)
  for current Phase 17 status.
- `implementation-plan.md`
  Milestone history and sequencing notes. Use [Roadmap](roadmap.md) for current
  validated state.

## Current Next Step

The selected next direction is:

```text
Phase 19: Larger Corpus Expansion
```

It is not started. Human review is required before implementation.

The independent challenge review lowered confidence from unconditional `High`
to `Medium-High` and approved the decision with reservations. The strongest
remaining risk is external validity: the current benchmark evidence may not
generalize beyond infrastructure/platform/developer-tooling repositories.
