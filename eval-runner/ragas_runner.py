from __future__ import annotations

import argparse
import json
from pathlib import Path


def load_jsonl(path: Path) -> list[dict]:
    rows: list[dict] = []
    with path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if line:
                rows.append(json.loads(line))
    return rows


def main() -> None:
    parser = argparse.ArgumentParser(description="Run Ragas metrics for Knowledge Forge JSONL outputs.")
    parser.add_argument("--input", required=True, help="JSONL file with question, answer, contexts, ground_truth")
    parser.add_argument("--output", required=True, help="Output JSON metrics file")
    args = parser.parse_args()

    rows = load_jsonl(Path(args.input))
    try:
        from datasets import Dataset
        from ragas import evaluate
        from ragas.metrics import answer_relevancy, context_precision, context_recall, faithfulness

        dataset = Dataset.from_list(rows)
        result = evaluate(dataset, metrics=[faithfulness, answer_relevancy, context_precision, context_recall])
        metrics = result.to_pandas().mean(numeric_only=True).to_dict()
    except Exception as exc:  # Keeps the contract testable without cloud keys or Ragas installed.
        metrics = {"status": "skipped", "reason": str(exc), "rows": len(rows)}

    Path(args.output).write_text(json.dumps(metrics, indent=2), encoding="utf-8")


if __name__ == "__main__":
    main()

