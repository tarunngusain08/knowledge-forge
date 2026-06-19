# Knowledge Forge Documentation

This directory is curated for contributors, reviewers, and interviewers. It
keeps product explanation, architecture, validation proof, and roadmap status in
separate places so there is not more than one source of truth for the same
claim.

## Start Here

- [Project README](../README.md): product overview, local setup, validation, and
  demo flow.
- [Phase 17 Validation Proof](proof/phase17-validation.md): accepted validation
  result, root causes, fixes, and final gate status.
- [Phase 18.6 Security Remediation](proof/phase18-6-security-remediation.md):
  tenant isolation, deployment trust-boundary findings, and security guardrails.
- [Roadmap](roadmap.md): what is complete, what is validated next, and what is
  not started.
- [Readiness Scorecard](readiness-scorecard.md): feature completeness,
  validation coverage, limitations, and next step.

## Core Design Docs

- [Functional Requirements, Non-Functional Requirements, and Scale Estimation](01-fr-nfr-scale-estimation.md)
- [High-Level Design and Component Design](02-hld-component-design.md)
- [Use Cases and Sequence Diagrams](03-usecases-sequence-diagrams.md)
- [Low-Level Design](04-lld.md)
- [Database Design and Schema](05-db-design-schema.md)
- [UI and Backend Quality Guide](06-ui-backend-quality.md)

## Architecture And Implementation

- [Architecture](architecture.md): current system architecture and provider
  boundaries.
- [Implementation Plan](implementation-plan.md): milestone history and phase
  sequence.
- [Storage Notes](storage.md): PostgreSQL BYTEA now, GCS production path later.
- [Future Multi-Tenancy](multitenancy.md): future tenant isolation notes.
- [Architecture Decision Records](adr): durable architectural decisions.

## Evaluation And Proof

- [Evaluation](evaluation.md): retrieval and generation metrics.
- [Acceptance Methodology](evaluations/acceptance-methodology.md): acceptance
  gates and how CI validates them.
- [Phase 18 Benchmark Design](evaluations/phase18-benchmark-design.md):
  benchmark schema, baselines, metrics, and freeze rules.
- [Phase 18 Benchmark Proof](evaluations/phase18-benchmark-proof.md):
  measured comparison against keyword and retrieval-only baselines.
- [Phase 18.5 Multi-Corpus Benchmark](evaluations/phase18-5-multi-corpus-benchmark.md):
  Helm and OpenTelemetry Collector generalization proof.
- [Phase 17 Validation Proof](proof/phase17-validation.md): only detailed
  Phase 17 proof narrative.
- [Phase 18.6 Security Remediation](proof/phase18-6-security-remediation.md):
  current security remediation proof narrative.
- [Phase 18.7 Release Readiness](proof/phase18-7-release-readiness.md):
  release-style review for architecture, security, benchmarks, docs,
  onboarding, and deployment.
- [Phase 17 Retrospective](postmortems/phase17-retrospective.md): reusable
  operating lessons for Phase 18.
- [Deep-Dive Report Case Study](case-studies/deep-dive-report.md): product
  workflow example.

## Portfolio And Demo

- [Portfolio Overview](portfolio/README.md): concise story for interviews and
  project review.
- [Repository Health](repository-health.md): documentation discoverability and
  cleanup recommendations.

## Tooling

- [Git Deliverable Commits](git-deliverable-commits.md): local CLI for clean
  deliverable-based commits.

## Historical Note

`phase17-hardening-review.md` is retained as historical context for the earlier
hardening branch. The canonical current validation summary is
[proof/phase17-validation.md](proof/phase17-validation.md).
