# Phase 17 Validation Proof

This is the canonical Phase 17 validation narrative for Knowledge Forge.

Raw audit folders remain local Desktop evidence and are not copied into the
repository. This document consolidates the accepted result so reviewers can
understand what failed, what changed, and why Phase 17 is considered complete.

## Accepted Result

Phase 17 is accepted.

```text
Gate 1 Refusal Matrix: pass
Gate 2 Answer Relevance: pass
Gate 3 Architecture Evidence: pass
Gate 4 Metric Integrity: pass
Gate 5 Benchmark Label Completeness: pass
Gate 6 Adversarial Benchmark: pass
Evaluator issues: 0
Acceptance validator: pass
```

The final accepted evidence package is local:

```text
/Users/radhakrishna/Desktop/knowledge-forge-product-conformance-r2/20260618-154837/
```

## Original Problem

The repository intelligence and Deep-Dive Report features were implemented, but
the first accepted reality audit showed that implementation was not yet
trustworthy enough for Phase 18 benchmark publication.

The main failures were:

- answerability checks could over-answer unsupported questions
- stricter support checks could also false-refuse valid repository questions
- architecture claims needed source-code evidence, not documentation-only
  inference
- grounding metrics needed explicit claim-to-evidence mappings
- reports and validator summaries needed to agree with evaluator output

The result was a useful product surface with insufficient proof.

## Validation Gates

The hardened acceptance framework evaluates six gates:

| Gate | Purpose |
| --- | --- |
| Gate 1 Refusal Matrix | Unsupported questions must refuse; answerable repository questions must not false-refuse. |
| Gate 2 Answer Relevance | Answers must cite required files, symbols, evidence groups, and expected facts. |
| Gate 3 Architecture Evidence | Architecture layers must be backed by source-code evidence, not README-only or directory-only shortcuts. |
| Gate 4 Metric Integrity | Grounding metrics must require claim-to-citation-to-file-to-evidence mappings. |
| Gate 5 Benchmark Label Completeness | Benchmark fixtures must include complete labels needed for validation. |
| Gate 6 Adversarial Benchmark | Known red-team failure patterns must fail if the product repeats them. |

## Before And After

| Stage | Gate 1 | Gate 2 | Gate 3 | Gate 4 | Gate 5 | Gate 6 | Evaluator Issues |
| --- | --- | --- | --- | --- | --- | --- | ---: |
| Baseline before Branch A | fail | fail | pass | pass | pass | fail | 4 |
| R2 Cycle 1 | pass | fail | pass | pass | pass | fail | 2 |
| R2 Cycle 2 | pass | pass | pass | pass | pass | pass | 0 |

## Root Causes

| Issue | Root Cause | Fix Applied | Proof |
| --- | --- | --- | --- |
| RF-004 | Complete repository-registration evidence was present, but the support gate returned `missing_domain_terms` because `matched_terms` was empty. | Added a complete-required-evidence allow path before the zero-term fallback while preserving hard refusals. | Final output answers repository registration with `repository_supported_fact` and cites repository service, HTTP handler, and indexer files. |
| RF-005 | Complete report-generation evidence was present, but the support gate returned `missing_domain_terms` because `matched_terms` was empty. | Used the same complete-required-evidence allow path. | Final output answers report generation with `repository_supported_fact` and cites `internal/codeqa/reports.go`. |
| AR-004 | Report-generator and repository-QA evidence were retrieved, but context assembly could drop one required evidence group under the token budget. | Added required-evidence-aware context assembly that frontloads and compacts already retrieved required evidence hits before generation. | Final output cites both `internal/codeqa/service.go` and `internal/codeqa/reports.go`. |

## Fix Scope

The accepted product conformance fixes used existing data:

- existing retrieval hits
- existing support-gate evidence groups
- existing repository metadata
- existing follow-up retrieval path
- existing context budget machinery

They did not add:

- new retrieval engines
- new indexes
- new storage systems
- graph traversal
- static intelligence
- planners
- workflow engines
- agent frameworks

## Validator Independence

The accepted review confirmed product runtime code does not read or depend on:

- acceptance fixtures
- benchmark labels
- candidate JSON
- validation outputs
- row IDs such as `RF-004`, `RF-005`, or `AR-004`
- previous audit artifacts

The only repository-wide references outside the acceptance directory are the
`make validate-acceptance` command wiring in the Makefile.

See:

- [R2 Root Cause Closure](../../R2-ROOT-CAUSE-CLOSURE.md)
- [Validator Independence](../../VALIDATOR-INDEPENDENCE.md)

## What This Proves

The accepted Phase 17 result proves that the actual product output satisfies the
hardened acceptance framework for the scoped repository intelligence and
Deep-Dive Report workflow.

It proves:

- the product refuses unsupported acceptance-suite questions
- answerable repository questions are not false-refused
- answer relevance checks pass for the required fixture set
- architecture evidence requires source-code support
- claim grounding mappings exist when grounding coverage is reported
- benchmark labels and adversarial rows are validated consistently

## What This Does Not Prove

This proof does not claim:

- production traffic performance at enterprise scale
- exhaustive repository-language coverage
- multi-repository intelligence
- graph retrieval quality
- static code intelligence quality
- autonomous code generation
- PR review correctness
- Phase 18 benchmark superiority over baseline retrieval

Those are future or explicitly out-of-scope items.

## Current Decision

```text
Phase 17: accepted
Branch B final consolidation: authorized separately
Phase 18: validated next step, not started here
Phase 19: future candidate, not started here
```
