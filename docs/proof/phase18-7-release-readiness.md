# Phase 18.7 Release Readiness Review

## Decision

```text
READY
```

Knowledge Forge is ready for human review before Phase 19 planning. The project
has current evidence for correctness, benchmark value, cross-corpus
generalization within infrastructure/tooling repositories, and security
hardening.

This review is not a new audit framework and does not start Phase 19.

## Release Readiness Scoring

| Area | Status | Evidence |
| --- | --- | --- |
| Architecture | Pass | [README](../../README.md), [Architecture](../architecture.md), [HLD](../02-hld-component-design.md), [LLD](../04-lld.md) |
| Security | Pass | [Phase 18.6 Security Remediation](phase18-6-security-remediation.md), [Cloud Run Deployment](../../deploy/cloud-run.md) |
| Benchmarks | Pass | [Phase 18 Benchmark Proof](../evaluations/phase18-benchmark-proof.md), [Phase 18.5 Multi-Corpus Benchmark](../evaluations/phase18-5-multi-corpus-benchmark.md) |
| Docs | Pass | [Documentation Index](../README.md), [Roadmap](../roadmap.md), [Readiness Scorecard](../readiness-scorecard.md) |
| Onboarding | Pass | [README Run Locally](../../README.md#run-locally), [README Run Validation](../../README.md#run-validation), [.env.example](../../.env.example) |
| Deployment | Pass | [Cloud Run Deployment](../../deploy/cloud-run.md), [Future Multi-Tenancy Notes](../multitenancy.md) |

Status derivation:

```text
All six areas pass.
No release-readiness blocker was found.
Final status: READY.
```

## What Knowledge Forge Does

Knowledge Forge is an evidence-grounded repository intelligence system. It
indexes documents and source repositories, retrieves evidence, answers
architecture/code questions with citations, generates Deep-Dive Reports, and
produces read-only implementation plans and impact analyses.

The product rule remains:

```text
Claims require evidence.
Unsupported conclusions should be refused or placed in missing context.
```

## How The Architecture Works

The current architecture is a Go/Chi backend with PostgreSQL, PostgreSQL FTS,
Pinecone, Vertex AI embeddings/ranking/Gemini, provider abstractions, async
workers, and React/Vite UI with Streamlit fallback.

The repository model remains:

```text
Repository
└── Branch
    └── Snapshot(commit SHA)
        ├── Files
        ├── Chunks
        ├── Symbols
        └── Graph
```

The implementation is documented in the README and design docs. No architecture
claim in this review depends on raw Desktop audit packages.

## Quality And Security Evidence

### Correctness

Phase 17 is accepted:

```text
6/6 acceptance gates pass
0 evaluator issues
Acceptance validator pass
```

Canonical proof:

- [Phase 17 Validation Proof](phase17-validation.md)

### Benchmark Value

Phase 18 is complete and partially proven. Knowledge Forge materially improved
over keyword and retrieval-only baselines in architecture, dependency/impact,
and grounding categories.

Canonical proof:

- [Phase 18 Benchmark Proof](../evaluations/phase18-benchmark-proof.md)

### Generalization

Phase 18.5 is complete. Results generalized within the approved
infrastructure/platform/developer-tooling scope:

```text
Helm: 20/20 correct
OpenTelemetry Collector: 18/20 correct
Overall: 68/70 correct
Cross-corpus stability: Moderately Stable
```

Canonical proof:

- [Phase 18.5 Multi-Corpus Benchmark](../evaluations/phase18-5-multi-corpus-benchmark.md)

### Security

Phase 18.6 is complete. P0 tenant-isolation findings were remediated:

- cross-user dense retrieval leakage
- cross-user PostgreSQL FTS leakage
- retrieval trace IDOR

P1 deployment trust-boundary findings were remediated or regression-tested:

- symlink escape
- unsafe `local_path`
- unsafe `remote_url`
- unauthenticated internal job endpoints

Canonical proof:

- [Phase 18.6 Security Remediation](phase18-6-security-remediation.md)

## How To Run Locally

The README provides the shortest local path:

```bash
cp .env.example .env
make tidy
make migrate-up
make test
docker compose up --build
```

Default endpoints:

- API: `http://localhost:8080`
- Health: `GET /healthz`
- UI through Docker Compose: port `8501`

Environment examples are documented in [.env.example](../../.env.example).

## How To Validate

The README documents the full validation set:

```bash
make test
make vet
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
cd ui/web && npm test && npm run lint && npm run build
docker compose config
make validate-acceptance
```

Phase-specific proof docs record the exact commands used for acceptance,
benchmarks, and security remediation.

## How To Deploy Safely

Deployment docs cover Cloud Run, Cloud SQL, Secret Manager, Cloud Tasks, Vertex
AI, and Pinecone. Phase 18.6 added required security configuration:

- `INTERNAL_WORKER_TOKEN` is required for internal job endpoints.
- `ALLOW_LOCAL_REPOSITORY_PATHS=false` should remain the hosted default.
- `ALLOWED_GIT_REMOTE_HOSTS` controls approved Git hosts.
- Internal worker tokens should be Secret Manager-backed and restart-rotatable.

Reference:

- [Cloud Run Deployment](../../deploy/cloud-run.md)

## Next Justified Roadmap Step

Phase 19 remains blocked until this release-readiness result is reviewed by a
human.

Phase 19 may begin only if:

- release readiness status is not `BLOCKED`
- Phase 18.6 security remediation remains merged
- Phase 18.5 benchmark result remains accepted
- no unresolved release-readiness blockers exist
- human review explicitly accepts Phase 18.7

After acceptance, the evidence-backed roadmap options are:

- broader corpus expansion beyond infrastructure/tooling repositories
- Repository Structure Indexing investigation
- Static Code Intelligence investigation

Graph Retrieval remains rejected until graph-specific failures dominate measured
results.

## Review Notes

Docs-only gaps found during this review:

- `docs/readiness-scorecard.md` still described Phase 18 as the next step.
- `docs/roadmap.md` did not yet include Phase 18.6 security remediation.
- `docs/README.md` did not yet link Phase 18.6 and Phase 18.7 proof.

These were documentation consistency gaps only and were corrected in this
branch. No product behavior, security implementation, benchmark output, or
validator contradiction was found.
