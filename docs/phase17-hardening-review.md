# Phase 17 Hardening Review

Status: branch-local hardening package for `tgusain/m17-report-quality-hardening`.

Phase 18 and Phase 19 remain blocked until this package is reviewed and no P0 findings remain.

## Before

| Area | Audit Finding |
| --- | --- |
| UI crash | Deep-Dive Report could crash when report arrays were `null`. |
| Refusal accuracy | Unsupported-question refusal accuracy was `0.00`. |
| Answerable-question accuracy | Not explicitly protected, so a naive refusal gate could over-refuse known-answer questions. |
| Evidence coverage | Baseline report evidence coverage was `0.56`. |
| Architecture quality | Architecture Overview could be driven by retrieved prose or docs instead of repository/code structure. |
| Latency metric validity | Python benchmark runner read Go duration nanoseconds as milliseconds. |

## After

| Area | Hardening Result | Evidence |
| --- | --- | --- |
| UI crash | Report formatting normalizes null or missing arrays to empty arrays. | `ui/web/src/reportFormat.ts`, `ui/web/tests/reportFormat.test.ts` |
| Refusal accuracy | Unsupported questions are gated before generation and return `I could not find this in the indexed context.` | `internal/codeqa/support.go`, `internal/codeqa/service_test.go` |
| Answerable-question accuracy | Known-answer questions such as authentication lookup continue to generate answers; Python metrics now expose `answerable_question_accuracy`. | `internal/codeqa/service_test.go`, `eval-runner/repo_benchmark_runner.py` |
| Evidence coverage | Report quality test requires evidence coverage `>= 0.70` and at least `25%` relative improvement from the `0.56` audit baseline. | `internal/codeqa/reports_test.go` |
| Architecture quality | Architecture Overview derives primary findings from cited code/repository paths and must identify API, retrieval/RAG, and UI layers. | `internal/codeqa/reports.go`, `internal/codeqa/reports_test.go` |
| Latency metric validity | `latency_ms` is the canonical API and benchmark field; legacy numeric `latency` is treated as nanoseconds and converted. | `internal/rag/types.go`, `internal/retrieval/code_service.go`, `eval-runner/test_repo_benchmark_runner.py` |

## Support-Gate Trace Metadata

Every repository Q&A trace now includes support-gate metadata in retrieval config:

```json
{
  "answerable": false,
  "reason": "missing_domain_terms",
  "matched_terms": [],
  "missing_terms": ["payroll"]
}
```

## Phase 18 Go/No-Go Gate

Phase 18 may start only after:

- Deep-Dive Report UI renders with null-array payloads.
- Markdown export controls remain available after report generation.
- `refusal_accuracy >= 0.90`.
- Answerable-question accuracy regression is `<= 5%`.
- Evidence coverage is `>= 0.70` and at least `25%` improved from `0.56`.
- Architecture Overview is primarily based on repository/code structure.
- Benchmark latency is reported in valid milliseconds.
- No P0 finding remains.

## P0 Findings

Current branch target: none after validation.

P0 examples:

- UI crash.
- Broken Markdown export.
- Invalid latency metrics.
- Refusal accuracy below target.
- Answerable-question regression above allowed threshold.

