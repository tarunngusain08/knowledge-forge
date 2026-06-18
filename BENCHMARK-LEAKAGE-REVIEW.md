# Phase 18.5 Benchmark Leakage Review

Leakage Status: PASS

## Runtime Product Search

Searched paths:

- `cmd/`
- `internal/`
- `ui/`
- `deploy/`

Searched terms:

- `helm-arch`
- `otel-arch`
- `phase18_5`
- `knowledge_forge_candidate`
- `keyword_baseline`
- `retrieval_only_baseline`
- `BENCHMARK`
- `expected_answer_facts`
- `should_refuse`

Findings:

- No benchmark row IDs, result paths, Phase 18.5 output files, or benchmark report names were found in runtime product code.
- One generic `should_refuse` JSON field exists in `internal/evaluation/repository_metrics.go`; this is evaluation data modeling, not runtime benchmark awareness.

## Evaluator Search

Searched paths:

- `eval-runner/`
- `eval-runner/repo_benchmark_runner.py`
- `eval-runner/acceptance/`
- `eval-runner/benchmarks/`

Findings:

- `repo_benchmark_runner.py` loads candidate and baseline JSONL files supplied by CLI arguments.
- Scoring uses expected labels from result rows to compute metrics; this is the benchmark scoring contract.
- Reports are generated after scoring and do not feed back into scoring.
- Candidate outputs and baseline outputs are static proof artifacts. The evaluator does not mutate labels or regenerate baselines from reports.
- No runtime product code reads benchmark outputs or labels.

## Baseline Behavior

- Keyword baseline represents simple lexical/file-content matching through sparse retrieved files, symbols, and evidence groups.
- Retrieval-only baseline represents retrieved files/evidence without final grounded answer generation.
- Knowledge Forge candidate output includes generation, support-gate, rerank, dense, and lexical stage contributions where relevant.

Conclusion: no benchmark leakage was found that would invalidate Phase 18.5 scoring or roadmap decisions.
