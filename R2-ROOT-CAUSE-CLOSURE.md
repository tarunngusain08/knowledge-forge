# R2 Root Cause Closure

Branch: `tgusain/m17-product-conformance-r2`

Base branch: `tgusain/m17-product-conformance`

Scope: Product conformance only. No validator fixtures, validator gates, README, portfolio docs, Phase 18 work, or Phase 19 work were changed.

## Result

Cycle 1 reduced the remaining issues to `AR-004` and downstream Gate 6.

Cycle 2 passed the fresh reality rerun:

```text
Gate 1 Refusal Matrix: pass
Gate 2 Answer Relevance: pass
Gate 3 Architecture Evidence: pass
Gate 4 Metric Integrity: pass
Gate 5 Benchmark Label Completeness: pass
Gate 6 Adversarial Benchmark: pass
Evaluator issues: 0
```

Fresh evidence package:

```text
/Users/radhakrishna/Desktop/knowledge-forge-product-conformance-r2/20260618-154837/
```

## Root Cause Closure

| Issue | Root Cause | Fix Applied | Proof |
| --- | --- | --- | --- |
| RF-004 | Complete repository-registration evidence was present, but `evaluateAnswerSupport` returned `missing_domain_terms` because `matched_terms` was empty. | Added a complete-required-evidence allow path before the zero-term fallback, while preserving hard refusal paths. | Fresh cycle-2 output answers `RF-004` with `repository_supported_fact`, cites `internal/repositories/service.go`, `internal/httpapi/repository_handlers.go`, and `internal/indexing/repository_indexer.go`, and has `missing_evidence=[]`. |
| RF-005 | Complete report-generation evidence was present, but `evaluateAnswerSupport` returned `missing_domain_terms` because `matched_terms` was empty. | Same complete-required-evidence allow path. | Fresh cycle-2 output answers `RF-005` with `repository_supported_fact`, cites `internal/codeqa/reports.go`, and has `missing_evidence=[]`. |
| AR-004 | Report-generator and repo-QA evidence were retrieved, but normal context assembly could drop one required evidence group under the token budget. | Added required-evidence-aware context assembly that frontloads and compacts already retrieved required evidence hits before generation. | Fresh cycle-2 output answers `AR-004`, cites both `internal/codeqa/service.go` and `internal/codeqa/reports.go`, and has `matched_evidence=[evidence_quality, repo_qa_service, report_generator]`. |

## Hypothesis

Was the single root-cause hypothesis confirmed?

```text
NO
```

The term-overlap hypothesis was confirmed for `RF-004` and `RF-005`.

The actual `AR-004` cause was separate: context assembly did not retain every required evidence group when oversized report and service chunks competed for the same token budget.

## Conformance Budget

The fixes used existing data only:

- Existing retrieval hits.
- Existing support-gate evidence groups.
- Existing repository metadata in hit metadata.
- Existing follow-up retrieval path for recoverable missing evidence groups.

No new retrieval engine, index, storage system, graph traversal, static intelligence, planner, workflow engine, or agent framework was added.

The chosen changes are the smallest viable fixes because they modify only:

- Support decision ordering for complete required evidence.
- Context assembly ordering/compaction for already retrieved required evidence.

## Validation

Required validation passed:

```text
GOCACHE=/private/tmp/knowledge-forge-go-cache make test
GOCACHE=/private/tmp/knowledge-forge-go-cache go vet ./...
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
cd ui/web && npm test
cd ui/web && npm run lint
cd ui/web && npm run build
docker compose config
make validate-acceptance
```

Fresh cycle-2 reality rerun passed:

```text
acceptance validation passed
6/6 gates pass
0 evaluator issues
```
