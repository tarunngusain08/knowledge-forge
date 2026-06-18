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


def write_markdown(path: Path, output: dict[str, Any]) -> None:
    title = output.get("report_title", "Phase 18 Benchmark Report")
    lines = [f"# {title}", ""]
    lines.extend(["## Candidate Metrics", ""])
    lines.extend(metric_table(output.get("metrics", {})))
    if output.get("baseline_comparisons"):
        lines.extend(["", "## Baseline Comparisons", ""])
        for name, comparison in output["baseline_comparisons"].items():
            lines.extend([f"### {name}", ""])
            lines.extend(delta_table(comparison.get("metric_delta", {})))
            lines.extend([""])
    if output.get("category_metrics"):
        lines.extend(["## Per-Category Metrics", ""])
        for category, category_metric in output["category_metrics"].items():
            lines.extend([f"### {category}", ""])
            lines.extend(metric_table(category_metric))
            lines.extend([""])
    if output.get("corpus_metrics"):
        lines.extend(["## Per-Corpus Metrics", ""])
        for corpus, corpus_metric in output["corpus_metrics"].items():
            lines.extend([f"### {corpus}", ""])
            lines.extend(metric_table(corpus_metric))
            lines.extend([""])
    if output.get("corpus_category_metrics"):
        lines.extend(["## Per-Corpus Category Metrics", ""])
        for key, corpus_category_metric in output["corpus_category_metrics"].items():
            lines.extend([f"### {key}", ""])
            lines.extend(metric_table(corpus_category_metric))
            lines.extend([""])
    if output.get("category_outcomes"):
        lines.extend(["## Category Outcomes", ""])
        lines.append("| Category | Outcome | Best baseline | Correct delta | Primary metric delta |")
        lines.append("| --- | --- | --- | ---: | ---: |")
        for outcome in output["category_outcomes"]:
            lines.append(
                f"| {outcome['category']} | {outcome['outcome']} | {outcome['best_baseline']} | "
                f"{outcome['correct_delta']} | {outcome['primary_metric_delta']:.3f} |"
            )
    if output.get("corpus_outcomes"):
        lines.extend(["", "## Corpus Outcomes", ""])
        lines.append("| Corpus | Outcome | Best baseline | Correct delta | Primary metric delta |")
        lines.append("| --- | --- | --- | ---: | ---: |")
        for outcome in output["corpus_outcomes"]:
            lines.append(
                f"| {outcome['corpus']} | {outcome['outcome']} | {outcome['best_baseline']} | "
                f"{outcome['correct_delta']} | {outcome['primary_metric_delta']:.3f} |"
            )
    if output.get("corpus_category_outcomes"):
        lines.extend(["", "## Corpus Category Outcomes", ""])
        lines.append("| Corpus / Category | Outcome | Best baseline | Correct delta | Primary metric delta |")
        lines.append("| --- | --- | --- | ---: | ---: |")
        for outcome in output["corpus_category_outcomes"]:
            lines.append(
                f"| {outcome['key']} | {outcome['outcome']} | {outcome['best_baseline']} | "
                f"{outcome['correct_delta']} | {outcome['primary_metric_delta']:.3f} |"
            )
    if output.get("stability"):
        stability = output["stability"]
        lines.extend(["", "## Cross-Corpus Stability", ""])
        lines.append(f"Classification: `{stability.get('classification', 'unknown')}`")
        lines.append("")
        lines.append("| Metric | Best corpus | Worst corpus | Range |")
        lines.append("| --- | --- | --- | ---: |")
        for item in stability.get("metrics", []):
            lines.append(
                f"| {item['metric']} | {item['best_corpus']} | {item['worst_corpus']} | {item['range']:.3f} |"
            )
    if output.get("failure_clusters"):
        lines.extend(["", "## Failure Clusters", ""])
        lines.append("| Cluster | Rows affected | Corpora affected | Possible cause |")
        lines.append("| --- | ---: | --- | --- |")
        for cluster in output["failure_clusters"]:
            lines.append(
                f"| {cluster['cluster']} | {cluster['rows_affected']} | "
                f"{', '.join(cluster['corpora_affected']) or 'None'} | {cluster['possible_cause']} |"
            )
    if output.get("baseline_comparisons"):
        lines.extend(["", "## Question Movement", ""])
        for name, comparison in output["baseline_comparisons"].items():
            lines.extend([f"### Compared With {name}", ""])
            for key, label in [
                ("questions_improved", "Improved"),
                ("questions_unchanged", "Unchanged"),
                ("questions_degraded", "Degraded"),
            ]:
                lines.append(f"#### {label}")
                questions = comparison.get(key, [])
                if questions:
                    for question in questions:
                        lines.append(f"- {question['id']}: {question['question']} ({question['delta']:.3f})")
                else:
                    lines.append("- None")
                lines.append("")
    path.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")


def metric_table(values: dict[str, Any]) -> list[str]:
    rows = ["| Metric | Value |", "| --- | ---: |"]
    answerable_count = int(values.get("answerable_question_count", 0))
    refusal_count = int(values.get("refusal_question_count", 0))
    refused_count = int(values.get("refused_count", 0))
    for key in [
        "question_count",
        "retrieval_recall",
        "evidence_recall",
        "answerable_question_accuracy",
        "refusal_precision",
        "refusal_recall",
        "grounding_coverage",
        "mrr",
        "citation_accuracy",
        "avg_latency_ms",
        "avg_cost_usd",
        "correctness_rate",
    ]:
        if key in values:
            if key in {"retrieval_recall", "evidence_recall", "grounding_coverage", "mrr", "citation_accuracy"} and answerable_count == 0:
                continue
            if key == "answerable_question_accuracy" and answerable_count == 0:
                continue
            if key in {"refusal_precision", "refusal_recall"} and refusal_count == 0 and refused_count == 0:
                continue
            rows.append(f"| {key} | {format_metric(values[key])} |")
    return rows


def delta_table(values: dict[str, Any]) -> list[str]:
    rows = ["| Metric | Delta |", "| --- | ---: |"]
    for key, value in values.items():
        rows.append(f"| {key} | {format_metric(value)} |")
    return rows


def format_metric(value: Any) -> str:
    if isinstance(value, float):
        return f"{value:.3f}"
    return str(value)


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
        "id": row.get("id", row["question"]),
        "category": row.get("category", "uncategorized"),
        "corpus": row.get("corpus", ""),
        "question": row["question"],
        "expected_files": row.get("expected_files", []),
        "retrieved_files": retrieved_files,
        "expected_symbols": row.get("expected_symbols", []),
        "retrieved_symbols": retrieved_symbols,
        "expected_answer_facts": row.get("expected_answer_facts", []),
        "matched_answer_facts": response.get("matched_answer_facts", []),
        "required_evidence_groups": row.get("required_evidence_groups", []),
        "retrieved_evidence_groups": retrieved.get("retrieved_evidence_groups", retrieval.get("matched_evidence", [])),
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
    refused_total = 0
    refused_correct = 0
    refusal_correct = 0
    answerable_total = 0
    answerable_correct = 0
    correct_total = 0
    totals = {
        "file_coverage": 0.0,
        "symbol_coverage": 0.0,
        "evidence_recall": 0.0,
        "fact_coverage": 0.0,
        "line_range_accuracy": 0.0,
        "citation_coverage": 0.0,
        "evidence_coverage": 0.0,
        "mrr": 0.0,
        "latency_ms": 0.0,
        "cost_usd": 0.0,
    }
    for result in results:
        file_coverage = coverage(result.get("expected_files", []), result.get("retrieved_files", []))
        totals["file_coverage"] += file_coverage
        if result.get("expected_files") and file_coverage > 0:
            file_hits += 1
        totals["symbol_coverage"] += coverage(result.get("expected_symbols", []), result.get("retrieved_symbols", []))
        totals["evidence_recall"] += coverage(result.get("required_evidence_groups", []), evidence_groups(result))
        totals["fact_coverage"] += coverage(result.get("expected_answer_facts", []), result.get("matched_answer_facts", []))
        totals["line_range_accuracy"] += line_coverage(result.get("expected_line_ranges", []), result.get("citation_line_ranges", []))
        totals["citation_coverage"] += 1.0 if result.get("citation_line_ranges") or result.get("retrieved_files") else 0.0
        totals["evidence_coverage"] += ratio(result.get("supported_claim_count", 0), result.get("total_claim_count", 0), default=1.0)
        totals["mrr"] += reciprocal_rank(result.get("expected_files", []), result.get("retrieved_files", []))
        totals["latency_ms"] += float(result.get("latency_ms", 0))
        totals["cost_usd"] += float(result.get("cost_usd", 0))
        if result.get("refused"):
            refused_total += 1
            if result.get("should_refuse"):
                refused_correct += 1
        if result.get("should_refuse"):
            refusal_total += 1
            if result.get("refused"):
                refusal_correct += 1
        else:
            answerable_total += 1
            if not result.get("refused") and answerable_question_correct(result):
                answerable_correct += 1
        if row_correct(result):
            correct_total += 1
    return {
        "question_count": question_count,
        "correct_count": correct_total,
        "correctness_rate": correct_total / question_count,
        "answerable_question_count": answerable_total,
        "refusal_question_count": refusal_total,
        "refused_count": refused_total,
        "file_hit_rate": file_hits / question_count,
        "retrieval_recall": totals["file_coverage"] / question_count,
        "file_coverage": totals["file_coverage"] / question_count,
        "symbol_coverage": totals["symbol_coverage"] / question_count,
        "evidence_recall": totals["evidence_recall"] / question_count,
        "answer_fact_coverage": totals["fact_coverage"] / question_count,
        "line_range_accuracy": totals["line_range_accuracy"] / question_count,
        "citation_accuracy": totals["line_range_accuracy"] / question_count,
        "citation_coverage": totals["citation_coverage"] / question_count,
        "grounding_coverage": totals["evidence_coverage"] / question_count,
        "evidence_coverage": totals["evidence_coverage"] / question_count,
        "mrr": totals["mrr"] / question_count,
        "refusal_accuracy": ratio(refusal_correct, refusal_total),
        "refusal_precision": ratio(refused_correct, refused_total),
        "refusal_recall": ratio(refusal_correct, refusal_total),
        "answerable_question_accuracy": ratio(answerable_correct, answerable_total),
        "avg_latency_ms": totals["latency_ms"] / question_count,
        "avg_cost_usd": totals["cost_usd"] / question_count,
    }


def category_metrics(results: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    by_category: dict[str, list[dict[str, Any]]] = {}
    for result in results:
        by_category.setdefault(result.get("category", "uncategorized"), []).append(result)
    return {category: metrics(category_results) for category, category_results in sorted(by_category.items())}


def corpus_metrics(results: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    by_corpus: dict[str, list[dict[str, Any]]] = {}
    for result in results:
        by_corpus.setdefault(result.get("corpus") or "unknown", []).append(result)
    return {corpus: metrics(corpus_results) for corpus, corpus_results in sorted(by_corpus.items())}


def corpus_category_metrics(results: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    grouped: dict[str, list[dict[str, Any]]] = {}
    for result in results:
        key = f"{result.get('corpus') or 'unknown'} / {result.get('category', 'uncategorized')}"
        grouped.setdefault(key, []).append(result)
    return {key: metrics(group_results) for key, group_results in sorted(grouped.items())}


def comparison_summary(baseline: list[dict[str, Any]], candidate: list[dict[str, Any]]) -> dict[str, Any]:
    before_by_question = {result_id(result): result for result in baseline}
    questions = {"improved": [], "unchanged": [], "degraded": []}
    for result in candidate:
        previous = before_by_question.get(result_id(result))
        if not previous:
            continue
        delta = question_score(result) - question_score(previous)
        status = "unchanged"
        if delta > 0.001:
            status = "improved"
        elif delta < -0.001:
            status = "degraded"
        questions[status].append({"id": result_id(result), "question": result["question"], "status": status, "delta": delta})

    before_metrics = metrics(baseline)
    after_metrics = metrics(candidate)
    metric_delta = {
        key: after_metrics.get(key, 0.0) - before_metrics.get(key, 0.0)
        for key in [
            "correctness_rate",
            "retrieval_recall",
            "file_hit_rate",
            "file_coverage",
            "symbol_coverage",
            "evidence_recall",
            "answer_fact_coverage",
            "line_range_accuracy",
            "citation_accuracy",
            "citation_coverage",
            "grounding_coverage",
            "evidence_coverage",
            "refusal_accuracy",
            "refusal_precision",
            "refusal_recall",
            "answerable_question_accuracy",
            "mrr",
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


def category_outcomes(baselines: dict[str, list[dict[str, Any]]], candidate: list[dict[str, Any]]) -> list[dict[str, Any]]:
    candidate_categories = category_metrics(candidate)
    baseline_categories = {name: category_metrics(rows) for name, rows in baselines.items()}
    outcomes: list[dict[str, Any]] = []
    for category, candidate_metrics in candidate_categories.items():
        best_name = ""
        best_baseline_score = -1.0
        best_baseline_correct = 0
        status = "IMPROVED"
        for name, category_by_name in baseline_categories.items():
            baseline_metrics = category_by_name.get(category, {"correct_count": 0, "question_count": 0})
            baseline_score = primary_metric_score(baseline_metrics)
            if baseline_score > best_baseline_score:
                best_name = name
                best_baseline_score = baseline_score
                best_baseline_correct = int(baseline_metrics.get("correct_count", 0))
            correct_delta = int(candidate_metrics.get("correct_count", 0)) - int(baseline_metrics.get("correct_count", 0))
            metric_delta = primary_metric_score(candidate_metrics) - baseline_score
            if metric_delta < -0.001:
                status = "DEGRADED"
            elif correct_delta < 3 and relative_improvement(primary_metric_score(baseline_metrics), primary_metric_score(candidate_metrics)) < 0.10:
                status = "UNCHANGED" if status != "DEGRADED" else status
        candidate_score = primary_metric_score(candidate_metrics)
        outcomes.append(
            {
                "category": category,
                "outcome": status,
                "best_baseline": best_name,
                "correct_delta": int(candidate_metrics.get("correct_count", 0)) - best_baseline_correct,
                "primary_metric_delta": candidate_score - best_baseline_score,
            }
        )
    return outcomes


def corpus_outcomes(baselines: dict[str, list[dict[str, Any]]], candidate: list[dict[str, Any]]) -> list[dict[str, Any]]:
    candidate_corpora = corpus_metrics(candidate)
    baseline_corpora = {name: corpus_metrics(rows) for name, rows in baselines.items()}
    outcomes: list[dict[str, Any]] = []
    for corpus, candidate_metric in candidate_corpora.items():
        best_name = ""
        best_baseline_score = -1.0
        best_baseline_correct = 0
        status = "IMPROVED"
        for name, corpus_by_name in baseline_corpora.items():
            baseline_metric = corpus_by_name.get(corpus, {"correct_count": 0, "question_count": 0})
            baseline_score = primary_metric_score(baseline_metric)
            if baseline_score > best_baseline_score:
                best_name = name
                best_baseline_score = baseline_score
                best_baseline_correct = int(baseline_metric.get("correct_count", 0))
            correct_delta = int(candidate_metric.get("correct_count", 0)) - int(baseline_metric.get("correct_count", 0))
            metric_delta = primary_metric_score(candidate_metric) - baseline_score
            if metric_delta < -0.001:
                status = "DEGRADED"
            elif correct_delta < 3 and relative_improvement(primary_metric_score(baseline_metric), primary_metric_score(candidate_metric)) < 0.10:
                status = "UNCHANGED" if status != "DEGRADED" else status
        candidate_score = primary_metric_score(candidate_metric)
        outcomes.append(
            {
                "corpus": corpus,
                "outcome": status,
                "best_baseline": best_name,
                "correct_delta": int(candidate_metric.get("correct_count", 0)) - best_baseline_correct,
                "primary_metric_delta": candidate_score - best_baseline_score,
            }
        )
    return outcomes


def corpus_category_outcomes(baselines: dict[str, list[dict[str, Any]]], candidate: list[dict[str, Any]]) -> list[dict[str, Any]]:
    candidate_groups = corpus_category_metrics(candidate)
    baseline_groups = {name: corpus_category_metrics(rows) for name, rows in baselines.items()}
    outcomes: list[dict[str, Any]] = []
    for key, candidate_metric in candidate_groups.items():
        best_name = ""
        best_baseline_score = -1.0
        best_baseline_correct = 0
        status = "IMPROVED"
        for name, group_by_name in baseline_groups.items():
            baseline_metric = group_by_name.get(key, {"correct_count": 0, "question_count": 0})
            baseline_score = primary_metric_score(baseline_metric)
            if baseline_score > best_baseline_score:
                best_name = name
                best_baseline_score = baseline_score
                best_baseline_correct = int(baseline_metric.get("correct_count", 0))
            correct_delta = int(candidate_metric.get("correct_count", 0)) - int(baseline_metric.get("correct_count", 0))
            metric_delta = primary_metric_score(candidate_metric) - baseline_score
            if metric_delta < -0.001:
                status = "DEGRADED"
            elif correct_delta < 3 and relative_improvement(primary_metric_score(baseline_metric), primary_metric_score(candidate_metric)) < 0.10:
                status = "UNCHANGED" if status != "DEGRADED" else status
        candidate_score = primary_metric_score(candidate_metric)
        outcomes.append(
            {
                "key": key,
                "outcome": status,
                "best_baseline": best_name,
                "correct_delta": int(candidate_metric.get("correct_count", 0)) - best_baseline_correct,
                "primary_metric_delta": candidate_score - best_baseline_score,
            }
        )
    return outcomes


def stability(corpus_metric_values: dict[str, dict[str, Any]], category_outcome_values: list[dict[str, Any]]) -> dict[str, Any]:
    tracked = [
        "retrieval_recall",
        "evidence_recall",
        "answerable_question_accuracy",
        "refusal_precision",
        "refusal_recall",
        "grounding_coverage",
    ]
    metric_ranges: list[dict[str, Any]] = []
    max_range = 0.0
    for metric in tracked:
        values = {
            corpus: float(corpus_metrics.get(metric, 0.0))
            for corpus, corpus_metrics in corpus_metric_values.items()
            if metric in corpus_metrics
        }
        if not values:
            continue
        best_corpus = max(values, key=values.get)
        worst_corpus = min(values, key=values.get)
        metric_range = values[best_corpus] - values[worst_corpus]
        max_range = max(max_range, metric_range)
        metric_ranges.append(
            {
                "metric": metric,
                "best_corpus": best_corpus,
                "worst_corpus": worst_corpus,
                "range": metric_range,
            }
        )
    any_corpus_degraded = any(outcome.get("outcome") == "DEGRADED" for outcome in category_outcome_values)
    major_failure = any(
        outcome.get("outcome") == "DEGRADED"
        and outcome.get("category") in {"architecture_implementation", "dependency_impact_testing", "deep_dive_grounding_architecture_evidence"}
        for outcome in category_outcome_values
    )
    if max_range <= 0.10 and not any_corpus_degraded:
        classification = "Stable"
    elif max_range <= 0.20 and not major_failure:
        classification = "Moderately Stable"
    else:
        classification = "Unstable"
    return {"classification": classification, "max_range": max_range, "metrics": metric_ranges}


def failure_clusters(results: list[dict[str, Any]]) -> list[dict[str, Any]]:
    clusters = {
        "missing symbol retrieval": {"rows": [], "cause": "Expected symbols were not retrieved."},
        "missing architecture evidence": {"rows": [], "cause": "Architecture evidence groups or files were missing."},
        "multi-hop dependency reasoning": {"rows": [], "cause": "Dependency/impact questions missed required multi-file evidence."},
        "impact analysis": {"rows": [], "cause": "Impact questions did not retrieve all affected files or facts."},
        "refusal classification": {"rows": [], "cause": "Unsupported questions were answered or supported questions were refused."},
        "grounding gaps": {"rows": [], "cause": "Claim grounding coverage was incomplete."},
        "citation gaps": {"rows": [], "cause": "Expected citation or file evidence was missing."},
    }
    for result in results:
        if row_correct(result):
            continue
        if result.get("should_refuse") and not result.get("refused"):
            clusters["refusal classification"]["rows"].append(result)
        if not result.get("should_refuse") and result.get("refused"):
            clusters["refusal classification"]["rows"].append(result)
        if result.get("expected_symbols") and coverage(result.get("expected_symbols", []), result.get("retrieved_symbols", [])) < 1.0:
            clusters["missing symbol retrieval"]["rows"].append(result)
        if result.get("category") == "architecture_implementation" and (
            coverage(result.get("expected_files", []), result.get("retrieved_files", [])) < 1.0
            or coverage(result.get("required_evidence_groups", []), evidence_groups(result)) < 1.0
        ):
            clusters["missing architecture evidence"]["rows"].append(result)
        if result.get("category") == "dependency_impact_testing" and coverage(result.get("required_evidence_groups", []), evidence_groups(result)) < 1.0:
            clusters["multi-hop dependency reasoning"]["rows"].append(result)
            clusters["impact analysis"]["rows"].append(result)
        if ratio(result.get("supported_claim_count", 0), result.get("total_claim_count", 0), default=1.0) < 1.0:
            clusters["grounding gaps"]["rows"].append(result)
        if result.get("expected_files") and coverage(result.get("expected_files", []), result.get("retrieved_files", [])) < 1.0:
            clusters["citation gaps"]["rows"].append(result)
    output: list[dict[str, Any]] = []
    for cluster, data in clusters.items():
        rows = data["rows"]
        output.append(
            {
                "cluster": cluster,
                "rows_affected": len({result_id(row) for row in rows}),
                "corpora_affected": sorted({row.get("corpus") or "unknown" for row in rows}),
                "row_ids": sorted({result_id(row) for row in rows}),
                "possible_cause": data["cause"],
            }
        )
    return sorted(output, key=lambda item: (-int(item["rows_affected"]), item["cluster"]))


def primary_metric_score(values: dict[str, Any]) -> float:
    present = [
        float(values.get("retrieval_recall", 0.0)),
        float(values.get("evidence_recall", 0.0)),
        float(values.get("grounding_coverage", 0.0)),
    ]
    if int(values.get("answerable_question_count", 0)) > 0:
        present.append(float(values.get("answerable_question_accuracy", 0.0)))
    if int(values.get("refusal_question_count", 0)) > 0 or int(values.get("refused_count", 0)) > 0:
        present.append(float(values.get("refusal_precision", 0.0)))
        present.append(float(values.get("refusal_recall", 0.0)))
    if not present:
        return 0.0
    return sum(present) / len(present)


def relative_improvement(before: float, after: float) -> float:
    if before == 0:
        return 1.0 if after > 0 else 0.0
    return (after - before) / abs(before)


def question_score(result: dict[str, Any]) -> float:
    if result.get("should_refuse"):
        return 1.0 if result.get("refused") else 0.0
    return (
        coverage(result.get("expected_files", []), result.get("retrieved_files", []))
        + coverage(result.get("expected_symbols", []), result.get("retrieved_symbols", []))
        + coverage(result.get("required_evidence_groups", []), evidence_groups(result))
        + coverage(result.get("expected_answer_facts", []), result.get("matched_answer_facts", []))
        + ratio(result.get("supported_claim_count", 0), result.get("total_claim_count", 0), default=1.0)
    ) / 5.0


def answerable_question_correct(result: dict[str, Any]) -> bool:
    expected_files = result.get("expected_files", [])
    expected_symbols = result.get("expected_symbols", [])
    if expected_files and coverage(expected_files, result.get("retrieved_files", [])) == 0:
        return False
    if expected_symbols and coverage(expected_symbols, result.get("retrieved_symbols", [])) == 0:
        return False
    required_evidence = result.get("required_evidence_groups", [])
    if required_evidence and coverage(required_evidence, evidence_groups(result)) < 0.5:
        return False
    expected_facts = result.get("expected_answer_facts", [])
    if expected_facts and coverage(expected_facts, result.get("matched_answer_facts", [])) < 0.5:
        return False
    return True


def row_correct(result: dict[str, Any]) -> bool:
    if result.get("should_refuse"):
        return bool(result.get("refused"))
    return not result.get("refused") and answerable_question_correct(result)


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


def reciprocal_rank(expected: list[str], actual: list[str]) -> float:
    if not expected:
        return 1.0
    expected_set = set(expected)
    for index, item in enumerate(actual, start=1):
        if item in expected_set:
            return 1.0 / index
    return 0.0


def evidence_groups(result: dict[str, Any]) -> list[str]:
    return result.get("retrieved_evidence_groups") or result.get("matched_evidence") or []


def result_id(result: dict[str, Any]) -> str:
    return result.get("id") or result.get("question", "")


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
    parser.add_argument("--baseline", action="append", default=[], help="Named baseline as name=path; may be repeated")
    parser.add_argument("--report-output", default="", help="Optional Markdown report output path")
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
    output = {
        "report_title": "Phase 18.5 Multi-Corpus Benchmark Report",
        "metrics": metrics(results),
        "category_metrics": category_metrics(results),
        "corpus_metrics": corpus_metrics(results),
        "corpus_category_metrics": corpus_category_metrics(results),
        "failure_clusters": failure_clusters(results),
        "results": results,
    }
    named_baselines: dict[str, list[dict[str, Any]]] = {}
    if args.baseline_input:
        baseline_rows = load_jsonl(Path(args.baseline_input))
        named_baselines["baseline"] = baseline_rows
        output["comparison"] = comparison_summary(baseline_rows, results)
    for baseline in args.baseline:
        if "=" not in baseline:
            raise SystemExit("--baseline must use name=path")
        name, path = baseline.split("=", 1)
        named_baselines[name] = load_jsonl(Path(path))
    if named_baselines:
        output["baseline_comparisons"] = {
            name: comparison_summary(baseline_rows, results)
            for name, baseline_rows in named_baselines.items()
        }
        output["baseline_metrics"] = {name: metrics(baseline_rows) for name, baseline_rows in named_baselines.items()}
        output["baseline_category_metrics"] = {name: category_metrics(baseline_rows) for name, baseline_rows in named_baselines.items()}
        output["baseline_corpus_metrics"] = {name: corpus_metrics(baseline_rows) for name, baseline_rows in named_baselines.items()}
        output["baseline_corpus_category_metrics"] = {name: corpus_category_metrics(baseline_rows) for name, baseline_rows in named_baselines.items()}
        output["category_outcomes"] = category_outcomes(named_baselines, results)
        output["corpus_outcomes"] = corpus_outcomes(named_baselines, results)
        output["corpus_category_outcomes"] = corpus_category_outcomes(named_baselines, results)
        output["stability"] = stability(output["corpus_metrics"], output["category_outcomes"])
    write_json(Path(args.output), output)
    if args.report_output:
        write_markdown(Path(args.report_output), output)


if __name__ == "__main__":
    main()
