# Repository Benchmarks

This directory stores repository-intelligence benchmark labels and proof
artifacts.

The controlled Phase 18 corpus is `synthetic_enterprise_monolith.jsonl`. It
points to `eval-runner/fixtures/synthetic-enterprise-monolith`, a deliberately
small repo with auth, billing, notifications, orders, audit logging, API wiring,
and tests.

Phase 18 freezes this corpus at 30 curated rows before result generation.

Phase 18.5 uses `phase18_5_multi_corpus.jsonl`, which keeps the original
synthetic rows unchanged and adds curated Helm and OpenTelemetry Collector
subsets for multi-corpus generalization testing.

## Row Schema

Each JSONL row includes:

- `id`
- `category`
- `question`
- `repository_fixture`
- `branch_name`
- `expected_files`
- `expected_symbols`
- `expected_answer_facts`
- `required_evidence_groups`
- `should_refuse`

`expected_line_ranges` are optional and are not required for Phase 18.

## Offline Scoring

Run offline scoring against saved result rows:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/results/phase18/knowledge_forge_candidate.jsonl \
  --baseline keyword=eval-runner/benchmarks/results/phase18/keyword_baseline.jsonl \
  --baseline retrieval_only=eval-runner/benchmarks/results/phase18/retrieval_only_baseline.jsonl \
  --output /tmp/phase18-benchmark.json \
  --report-output /tmp/phase18-benchmark.md
```

## Live API Mode

Run against a live API when a repository has been indexed:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl \
  --output /tmp/repo-benchmark.json \
  --api-base-url http://localhost:8080 \
  --token "$KNOWLEDGE_FORGE_TOKEN" \
  --repository-id "$REPOSITORY_ID" \
  --top-k 8 \
  --reranker
```

## Phase 18 Artifacts

Committed Phase 18 artifacts live under:

```text
eval-runner/benchmarks/results/phase18/
```

Required files:

- `keyword_baseline.jsonl`
- `retrieval_only_baseline.jsonl`
- `knowledge_forge_candidate.jsonl`
- `phase18-benchmark.json`
- `phase18-benchmark.md`

## Phase 18.5 Artifacts

Committed Phase 18.5 artifacts live under:

```text
eval-runner/benchmarks/results/phase18_5/
```

Required files:

- `keyword_baseline.jsonl`
- `retrieval_only_baseline.jsonl`
- `knowledge_forge_candidate.jsonl`
- `phase18_5-benchmark.json`
- `phase18_5-benchmark.md`

Run offline scoring:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/results/phase18_5/knowledge_forge_candidate.jsonl \
  --baseline keyword=eval-runner/benchmarks/results/phase18_5/keyword_baseline.jsonl \
  --baseline retrieval_only=eval-runner/benchmarks/results/phase18_5/retrieval_only_baseline.jsonl \
  --output /tmp/phase18_5-benchmark.json \
  --report-output /tmp/phase18_5-benchmark.md
```

## Decision Gates

- Keep symbol retrieval only if expected file/symbol coverage improves by at
  least 10%.
- Enable graph retrieval only if Recall@K or file coverage improves by at least
  10% without increasing latency by more than 25%.
- Keep reranking only if MRR improves by at least 5% or faithfulness improves
  meaningfully without unacceptable cost.
- Disable or delete retrieval components that do not earn their place.
