# Evaluation Methodology

RAG-bot evaluates retrieval and generation separately.

## Retrieval Metrics

- Hit Rate: at least one expected source appears in retrieved results
- Recall@K: fraction of expected sources retrieved
- MRR: reciprocal rank of the first expected source
- Retrieval latency
- Cost per answer

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

