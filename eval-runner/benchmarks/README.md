# Repository Benchmarks

This directory stores repository-intelligence benchmark labels for Phase 13.

The first controlled corpus is `synthetic_enterprise_monolith.jsonl`. It points
to `eval-runner/fixtures/synthetic-enterprise-monolith`, a deliberately small
repo with auth, billing, notifications, interface boundaries, API wiring, and
tests.

Each JSONL row can include:

- `question`
- `repository_fixture`
- `branch_name`
- `expected_files`
- `expected_symbols`
- `expected_line_ranges`
- `expected_answer_facts`
- `should_refuse`

Run offline scoring against saved result rows:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl \
  --output /tmp/repo-benchmark.json
```

Run against a live API:

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

Decision gates:

- Keep symbol retrieval only if expected file/symbol coverage improves by at least 10%.
- Enable graph retrieval only if Recall@K or file coverage improves by at least 10% without increasing latency by more than 25%.
- Keep reranking only if MRR improves by at least 5% or faithfulness improves meaningfully without unacceptable cost.
- Disable or delete retrieval components that do not earn their place.
