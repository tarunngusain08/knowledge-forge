# Evaluation Methodology

Knowledge Forge evaluates retrieval and generation separately.

For acceptance gates and the current Phase 17 conformance proof, see:

- [Acceptance Methodology](evaluations/acceptance-methodology.md)
- [Phase 17 Validation Proof](proof/phase17-validation.md)

## Retrieval Metrics

- Hit Rate: at least one expected source appears in retrieved results
- Recall@K: fraction of expected sources retrieved
- MRR: reciprocal rank of the first expected source
- Retrieval latency
- Cost per answer

## Repository Benchmark Metrics

Repository intelligence is evaluated with source-aware labels, not only vector
IDs.

- File coverage: expected files retrieved or cited
- Symbol coverage: expected symbols retrieved or cited
- Line-range accuracy: citations overlap expected line ranges
- Citation coverage: answer includes usable evidence
- Evidence coverage: supported answer claims divided by total answer claims
- Refusal accuracy: unsupported questions are refused instead of hallucinated
- Average latency and cost per answer

Benchmark reports must separate:

- questions improved
- questions unchanged
- questions degraded

This keeps retrieval changes honest. Dense, lexical, symbol, graph, and reranker
stages must earn their place through measurable improvement.

Phase 13 includes a controlled synthetic enterprise monolith fixture under
`eval-runner/fixtures/synthetic-enterprise-monolith` and benchmark labels under
`eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl`.

Run the repository benchmark runner offline:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input eval-runner/benchmarks/synthetic_enterprise_monolith.jsonl \
  --output /tmp/repo-benchmark.json
```

Run it against a live API by adding `--api-base-url`, `--token`, and
`--repository-id`.

Compare a candidate result file against a saved baseline:

```bash
python3 eval-runner/repo_benchmark_runner.py \
  --input /tmp/knowledge-forge-candidate.jsonl \
  --baseline-input /tmp/naive-semantic-baseline.jsonl \
  --output /tmp/repo-benchmark-comparison.json
```

The comparison output includes improved, unchanged, and degraded questions,
metric deltas, latency/cost deltas, and retrieval-stage contribution deltas.

Decision thresholds:

- Keep symbol retrieval only if file/symbol coverage improves by at least 10%.
- Enable graph retrieval only if Recall@K or file coverage improves by at least
  10% without latency increasing by more than 25%.
- Keep reranking only if MRR improves by at least 5% or faithfulness improves
  meaningfully without unacceptable cost.

## Generation Metrics

The Python `eval-runner` accepts JSONL rows with:

```json
{"question":"...","answer":"...","contexts":["..."],"ground_truth":"..."}
```

It runs Ragas metrics when dependencies and model configuration are available:

- Faithfulness
- Answer relevancy
- Context precision
- Context recall

If Ragas is not installed, the runner emits a skipped result while preserving
the file contract.
