from __future__ import annotations

import argparse
import json
import urllib.request
from pathlib import Path
from typing import Any


def load_jsonl(path: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if line:
                rows.append(json.loads(line))
    return rows


def write_json(path: Path, value: dict[str, Any]) -> None:
    path.write_text(json.dumps(value, indent=2), encoding="utf-8")


def call_ask(api_base_url: str, token: str, row: dict[str, Any], repository_id: str, top_k: int, reranker: bool) -> dict[str, Any]:
    body = {
        "repository_id": row.get("repository_id") or repository_id,
        "branch_name": row.get("branch_name", "main"),
        "question": row["question"],
        "top_k": row.get("top_k", top_k),
        "reranker_enabled": row.get("reranker_enabled", reranker),
    }
    data = json.dumps(body).encode("utf-8")
    request = urllib.request.Request(
        api_base_url.rstrip("/") + "/v1/ask",
        data=data,
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=60) as response:  # nosec: benchmark CLI target is explicit user input
        return json.loads(response.read().decode("utf-8"))


def result_from_response(row: dict[str, Any], response: dict[str, Any]) -> dict[str, Any]:
    citations = response.get("citations", [])
    retrieval = response.get("retrieval", {})
    retrieved_files = ordered_unique(
        [citation.get("path", "") for citation in citations]
        + [hit.get("chunk", {}).get("metadata", {}).get("path", "") for hit in retrieval.get("reranked_hits", [])]
    )
    retrieved_symbols = ordered_unique(
        [hit.get("chunk", {}).get("metadata", {}).get("symbol_name", "") for hit in retrieval.get("reranked_hits", [])]
    )
    citation_ranges = [
        {
            "path": citation.get("path", ""),
            "start_line": int(citation.get("start_line") or 0),
            "end_line": int(citation.get("end_line") or 0),
        }
        for citation in citations
        if citation.get("path")
    ]
    answer = response.get("answer", "")
    return {
        "question": row["question"],
        "expected_files": row.get("expected_files", []),
        "retrieved_files": retrieved_files,
        "expected_symbols": row.get("expected_symbols", []),
        "retrieved_symbols": retrieved_symbols,
        "expected_line_ranges": row.get("expected_line_ranges", []),
        "citation_line_ranges": citation_ranges,
        "should_refuse": row.get("should_refuse", False),
        "refused": answer.lower().startswith("i could not find"),
        "supported_claim_count": row.get("supported_claim_count", 0),
        "total_claim_count": row.get("total_claim_count", 0),
        "latency_ms": latency_ms(retrieval),
        "cost_usd": float(row.get("cost_usd", 0)),
        "answer": answer,
        "trace_id": response.get("trace_id", ""),
    }


def metrics(results: list[dict[str, Any]]) -> dict[str, Any]:
    if not results:
        return {"question_count": 0}
    question_count = len(results)
    file_hits = 0
    refusal_total = 0
    refusal_correct = 0
    answerable_total = 0
    answerable_correct = 0
    totals = {
        "file_coverage": 0.0,
        "symbol_coverage": 0.0,
        "line_range_accuracy": 0.0,
        "citation_coverage": 0.0,
        "evidence_coverage": 0.0,
        "latency_ms": 0.0,
        "cost_usd": 0.0,
    }
    for result in results:
        file_coverage = coverage(result.get("expected_files", []), result.get("retrieved_files", []))
        totals["file_coverage"] += file_coverage
        if result.get("expected_files") and file_coverage > 0:
            file_hits += 1
        totals["symbol_coverage"] += coverage(result.get("expected_symbols", []), result.get("retrieved_symbols", []))
        totals["line_range_accuracy"] += line_coverage(result.get("expected_line_ranges", []), result.get("citation_line_ranges", []))
        totals["citation_coverage"] += 1.0 if result.get("citation_line_ranges") or result.get("retrieved_files") else 0.0
        totals["evidence_coverage"] += ratio(result.get("supported_claim_count", 0), result.get("total_claim_count", 0), default=1.0)
        totals["latency_ms"] += float(result.get("latency_ms", 0))
        totals["cost_usd"] += float(result.get("cost_usd", 0))
        if result.get("should_refuse"):
            refusal_total += 1
            if result.get("refused"):
                refusal_correct += 1
        else:
            answerable_total += 1
            if not result.get("refused") and answerable_question_correct(result):
                answerable_correct += 1
    return {
        "question_count": question_count,
        "file_hit_rate": file_hits / question_count,
        "file_coverage": totals["file_coverage"] / question_count,
        "symbol_coverage": totals["symbol_coverage"] / question_count,
        "line_range_accuracy": totals["line_range_accuracy"] / question_count,
        "citation_coverage": totals["citation_coverage"] / question_count,
        "evidence_coverage": totals["evidence_coverage"] / question_count,
        "refusal_accuracy": ratio(refusal_correct, refusal_total),
        "answerable_question_accuracy": ratio(answerable_correct, answerable_total),
        "avg_latency_ms": totals["latency_ms"] / question_count,
        "avg_cost_usd": totals["cost_usd"] / question_count,
    }


def comparison_summary(baseline: list[dict[str, Any]], candidate: list[dict[str, Any]]) -> dict[str, Any]:
    before_by_question = {result["question"]: result for result in baseline}
    questions = {"improved": [], "unchanged": [], "degraded": []}
    for result in candidate:
        previous = before_by_question.get(result["question"])
        if not previous:
            continue
        delta = question_score(result) - question_score(previous)
        status = "unchanged"
        if delta > 0.001:
            status = "improved"
        elif delta < -0.001:
            status = "degraded"
        questions[status].append({"question": result["question"], "status": status, "delta": delta})

    before_metrics = metrics(baseline)
    after_metrics = metrics(candidate)
    metric_delta = {
        key: after_metrics.get(key, 0.0) - before_metrics.get(key, 0.0)
        for key in [
            "file_hit_rate",
            "file_coverage",
            "symbol_coverage",
            "line_range_accuracy",
            "citation_coverage",
            "evidence_coverage",
            "refusal_accuracy",
            "answerable_question_accuracy",
            "avg_latency_ms",
            "avg_cost_usd",
        ]
    }
    return {
        "questions_improved": questions["improved"],
        "questions_unchanged": questions["unchanged"],
        "questions_degraded": questions["degraded"],
        "metric_delta": metric_delta,
        "stage_contribution_delta": stage_contribution_delta(baseline, candidate),
    }


def question_score(result: dict[str, Any]) -> float:
    if result.get("should_refuse"):
        return 1.0 if result.get("refused") else 0.0
    return (
        coverage(result.get("expected_files", []), result.get("retrieved_files", []))
        + coverage(result.get("expected_symbols", []), result.get("retrieved_symbols", []))
        + line_coverage(result.get("expected_line_ranges", []), result.get("citation_line_ranges", []))
        + ratio(result.get("supported_claim_count", 0), result.get("total_claim_count", 0), default=1.0)
    ) / 4.0


def answerable_question_correct(result: dict[str, Any]) -> bool:
    expected_files = result.get("expected_files", [])
    expected_symbols = result.get("expected_symbols", [])
    if expected_files and coverage(expected_files, result.get("retrieved_files", [])) == 0:
        return False
    if expected_symbols and coverage(expected_symbols, result.get("retrieved_symbols", [])) == 0:
        return False
    return True


def latency_ms(retrieval: dict[str, Any]) -> float:
    if "latency_ms" in retrieval and retrieval.get("latency_ms") is not None:
        return float(retrieval.get("latency_ms") or 0)
    if "latency" in retrieval and retrieval.get("latency") is not None:
        return float(retrieval.get("latency") or 0) / 1_000_000.0
    return 0.0


def coverage(expected: list[str], actual: list[str]) -> float:
    if not expected:
        return 1.0
    actual_set = set(actual)
    return sum(1 for item in expected if item in actual_set) / len(expected)


def line_coverage(expected: list[dict[str, Any]], actual: list[dict[str, Any]]) -> float:
    if not expected:
        return 1.0
    found = 0
    for want in expected:
        if any(overlaps(want, got) for got in actual):
            found += 1
    return found / len(expected)


def overlaps(a: dict[str, Any], b: dict[str, Any]) -> bool:
    return (
        a.get("path") == b.get("path")
        and int(a.get("start_line", 0)) <= int(b.get("end_line", 0))
        and int(b.get("start_line", 0)) <= int(a.get("end_line", 0))
    )


def ratio(numerator: float, denominator: float, default: float = 0.0) -> float:
    if denominator == 0:
        return default
    return numerator / denominator


def ordered_unique(values: list[str]) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for value in values:
        if value and value not in seen:
            seen.add(value)
            out.append(value)
    return out


def stage_contribution_delta(baseline: list[dict[str, Any]], candidate: list[dict[str, Any]]) -> dict[str, int]:
    delta: dict[str, int] = {}
    for result in candidate:
        for stage, count in result.get("stage_contributions", {}).items():
            delta[stage] = delta.get(stage, 0) + int(count)
    for result in baseline:
        for stage, count in result.get("stage_contributions", {}).items():
            delta[stage] = delta.get(stage, 0) - int(count)
    return delta


def main() -> None:
    parser = argparse.ArgumentParser(description="Run Knowledge Forge repository benchmark metrics.")
    parser.add_argument("--input", required=True, help="Benchmark JSONL or saved result JSONL")
    parser.add_argument("--output", required=True, help="Output JSON metrics file")
    parser.add_argument("--api-base-url", default="", help="Optional API URL; when set, calls /v1/ask")
    parser.add_argument("--token", default="", help="Bearer token for API mode")
    parser.add_argument("--repository-id", default="", help="Default repository id for API mode")
    parser.add_argument("--baseline-input", default="", help="Optional saved baseline JSONL for comparison output")
    parser.add_argument("--top-k", type=int, default=8)
    parser.add_argument("--reranker", action="store_true")
    args = parser.parse_args()

    rows = load_jsonl(Path(args.input))
    if args.api_base_url:
        if not args.token or not args.repository_id:
            raise SystemExit("--token and --repository-id are required in API mode")
        results = [
            result_from_response(row, call_ask(args.api_base_url, args.token, row, args.repository_id, args.top_k, args.reranker))
            for row in rows
        ]
    else:
        results = rows
    output = {"metrics": metrics(results), "results": results}
    if args.baseline_input:
        output["comparison"] = comparison_summary(load_jsonl(Path(args.baseline_input)), results)
    write_json(Path(args.output), output)


if __name__ == "__main__":
    main()
