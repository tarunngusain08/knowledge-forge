#!/usr/bin/env python3
"""Executable acceptance gates for Knowledge Forge validation hardening.

This runner intentionally validates candidate output artifacts instead of
changing product behavior. It encodes the acceptance-redesign package as CI
checks that fail on false refusals, false answers, irrelevant answers, fake
architecture evidence, metric gaming, and incomplete benchmark labels.
"""

from __future__ import annotations

import argparse
import json
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Iterable


Decision = str


@dataclass
class RowIssue:
    gate: str
    row_id: str
    severity: str
    message: str


@dataclass
class ValidationState:
    fixture: dict[str, Any]
    candidate: dict[str, Any]
    issues: list[RowIssue] = field(default_factory=list)
    refusal_rows: list[dict[str, Any]] = field(default_factory=list)
    relevance_rows: list[dict[str, Any]] = field(default_factory=list)
    architecture_rows: list[dict[str, Any]] = field(default_factory=list)
    metric_rows: list[dict[str, Any]] = field(default_factory=list)
    label_rows: list[dict[str, Any]] = field(default_factory=list)
    gate_status: dict[str, bool] = field(default_factory=dict)

    def fail(self, gate: str, row_id: str, message: str) -> None:
        self.issues.append(RowIssue(gate, row_id, "fail", message))

    @property
    def passed(self) -> bool:
        return not self.issues


def load_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def as_set(value: Any) -> set[str]:
    if value is None:
        return set()
    if isinstance(value, str):
        return {value}
    if isinstance(value, Iterable):
        return {str(item) for item in value}
    return {str(value)}


def index_by_id(rows: Iterable[dict[str, Any]], key: str = "id") -> dict[str, dict[str, Any]]:
    return {str(row.get(key, "")): row for row in rows if row.get(key)}


def support_gate_has_trace(result: dict[str, Any]) -> bool:
    gate = result.get("support_gate") or {}
    required = {"answerable", "reason", "matched_evidence", "missing_evidence"}
    return required.issubset(gate.keys())


def support_gate_values(result: dict[str, Any], field: str) -> set[str]:
    gate = result.get("support_gate") or {}
    return as_set(gate.get(field))


def normalize_decision(value: Any) -> Decision:
    value = str(value or "").strip().lower()
    if value in {"answered", "answer"}:
        return "answer"
    if value in {"refused", "refuse"}:
        return "refuse"
    return value


def evaluate_refusal_matrix(state: ValidationState) -> None:
    gate_name = "Gate 1 Refusal Matrix"
    results = index_by_id(state.candidate.get("refusal_results", []))
    false_refusals = 0
    false_answers = 0

    for fixture_row in state.fixture.get("refusal_matrix", []):
        row_id = fixture_row["id"]
        result = results.get(row_id)
        expected = fixture_row["expected_decision"]
        category = fixture_row["category"]
        row_summary = {
            "id": row_id,
            "category": category,
            "question": fixture_row["question"],
            "expected_decision": expected,
            "actual_decision": None,
            "status": "pass",
            "issues": []
        }

        if not result:
            msg = "candidate output missing refusal result row"
            state.fail(gate_name, row_id, msg)
            row_summary["status"] = "fail"
            row_summary["issues"].append(msg)
            state.refusal_rows.append(row_summary)
            continue

        actual = normalize_decision(result.get("decision"))
        row_summary["actual_decision"] = actual

        if not support_gate_has_trace(result):
            msg = "support-gate trace lacks answerable, reason, matched_evidence, or missing_evidence"
            state.fail(gate_name, row_id, msg)
            row_summary["status"] = "fail"
            row_summary["issues"].append(msg)

        if expected == "answer" and actual == "refuse":
            false_refusals += 1
            msg = "answerable row was refused"
            state.fail(gate_name, row_id, msg)
            row_summary["status"] = "fail"
            row_summary["issues"].append(msg)

        if expected == "refuse" and actual == "answer":
            false_answers += 1
            msg = "unsupported row was answered"
            state.fail(gate_name, row_id, msg)
            row_summary["status"] = "fail"
            row_summary["issues"].append(msg)

        if expected == "answer" and actual == "answer":
            matched = support_gate_values(result, "matched_evidence")
            missing_groups = set(fixture_row.get("required_evidence_groups", [])) - matched
            if missing_groups:
                msg = f"answerable row answered without required evidence groups: {sorted(missing_groups)}"
                state.fail(gate_name, row_id, msg)
                row_summary["status"] = "fail"
                row_summary["issues"].append(msg)

        if expected == "refuse":
            reason = str(result.get("refusal_reason") or (result.get("support_gate") or {}).get("reason") or "")
            expected_reason = str(fixture_row.get("expected_refusal_reason") or "")
            missing = support_gate_values(result, "missing_evidence")
            matched = support_gate_values(result, "matched_evidence")
            shortcuts = set(fixture_row.get("forbidden_evidence_shortcuts", []))

            if actual == "refuse" and expected_reason and reason != expected_reason:
                msg = f"refusal reason mismatch: expected {expected_reason!r}, got {reason!r}"
                state.fail(gate_name, row_id, msg)
                row_summary["status"] = "fail"
                row_summary["issues"].append(msg)

            if actual == "answer" and shortcuts.intersection(matched):
                msg = f"answered using forbidden shortcut evidence: {sorted(shortcuts.intersection(matched))}"
                state.fail(gate_name, row_id, msg)
                row_summary["status"] = "fail"
                row_summary["issues"].append(msg)

            required_missing = set(fixture_row.get("required_evidence_groups", []))
            if actual == "refuse" and required_missing and not required_missing.intersection(missing):
                msg = f"refusal trace does not name missing required evidence: {sorted(required_missing)}"
                state.fail(gate_name, row_id, msg)
                row_summary["status"] = "fail"
                row_summary["issues"].append(msg)

        state.refusal_rows.append(row_summary)

    state.gate_status[gate_name] = not any(issue.gate == gate_name for issue in state.issues)


def classify_relevance(fixture_row: dict[str, Any], result: dict[str, Any] | None) -> tuple[str, list[str]]:
    if not result:
        return "unscored", ["candidate output missing answer relevance result row"]

    if normalize_decision(result.get("decision")) == "refuse":
        return "unsupported", ["answerable relevance row was refused"]

    retrieved = as_set(result.get("retrieved_files")) | as_set(result.get("cited_files"))
    expected_files = set(fixture_row.get("expected_files", []))
    expected_symbols = set(fixture_row.get("expected_symbols", []))
    required_groups = set(fixture_row.get("required_evidence_groups", []))
    expected_facts = list(fixture_row.get("expected_answer_facts", []))
    evidence_groups = result.get("evidence_groups") or {}
    claim_support = result.get("claim_support") or {}
    symbols = as_set(result.get("actual_symbols")) | as_set(result.get("cited_symbols"))

    issues: list[str] = []
    missing_files = sorted(expected_files - retrieved)
    if missing_files:
        issues.append(f"missing expected files: {missing_files}")

    missing_groups = [
        group for group in sorted(required_groups)
        if (evidence_groups.get(group) or {}).get("status") != "supported"
    ]
    if missing_groups:
        issues.append(f"missing required evidence groups: {missing_groups}")

    missing_symbols = sorted(expected_symbols - symbols)
    if missing_symbols:
        issues.append(f"missing expected symbols: {missing_symbols}")

    missing_facts = [
        fact for fact in expected_facts
        if not claim_support.get(fact)
    ]
    if missing_facts:
        issues.append(f"missing supported answer facts: {missing_facts}")

    if not issues:
        return "relevant", []

    if not retrieved.intersection(expected_files):
        return "irrelevant", issues

    if missing_facts:
        return "unsupported", issues

    return "partially_relevant", issues


def evaluate_answer_relevance(state: ValidationState) -> None:
    gate_name = "Gate 2 Answer Relevance"
    results = index_by_id(state.candidate.get("answer_relevance_results", []))
    classifications: list[str] = []

    for fixture_row in state.fixture.get("answer_relevance", []):
        row_id = fixture_row["id"]
        result = results.get(row_id)
        classification, issues = classify_relevance(fixture_row, result)
        classifications.append(classification)

        state.relevance_rows.append({
            "id": row_id,
            "question": fixture_row["question"],
            "classification": classification,
            "expected_files": fixture_row.get("expected_files", []),
            "actual_files": sorted(as_set((result or {}).get("retrieved_files")) | as_set((result or {}).get("cited_files"))),
            "issues": issues
        })

        if classification != "relevant":
            state.fail(gate_name, row_id, "; ".join(issues) or f"classified as {classification}")

    state.gate_status[gate_name] = classifications and all(item == "relevant" for item in classifications)


def layer_by_name(result: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {str(layer.get("name", "")).lower(): layer for layer in result.get("layers", []) if layer.get("name")}


def has_line_ranges(layer: dict[str, Any]) -> bool:
    return bool(layer.get("line_ranges"))


def is_high_confidence(layer: dict[str, Any]) -> bool:
    return str(layer.get("confidence", "")).lower() == "high"


def is_source_evidence(layer: dict[str, Any]) -> bool:
    return str(layer.get("evidence_type", "")).lower() == "source_code"


def evaluate_architecture(state: ValidationState) -> None:
    gate_name = "Gate 3 Architecture Evidence"
    results = index_by_id(state.candidate.get("architecture_results", []), key="fixture_id")

    for fixture_row in state.fixture.get("architecture_fixtures", []):
        row_id = fixture_row["id"]
        result = results.get(row_id)
        row_issues: list[str] = []

        if not result:
            row_issues.append("candidate output missing architecture fixture result")
        elif fixture_row["expected_result"] == "pass":
            layers = layer_by_name(result)
            for layer_name in fixture_row.get("required_layers", []):
                layer = layers.get(layer_name)
                if not layer:
                    row_issues.append(f"missing required layer {layer_name}")
                    continue
                if not is_source_evidence(layer):
                    row_issues.append(f"layer {layer_name} lacks source-code evidence")
                if not has_line_ranges(layer):
                    row_issues.append(f"layer {layer_name} lacks line ranges")
                if not layer.get("files") or not layer.get("packages"):
                    row_issues.append(f"layer {layer_name} lacks files or packages")
        else:
            layers = layer_by_name(result)
            for layer_name in fixture_row.get("forbidden_high_confidence_layers", []):
                layer = layers.get(layer_name)
                if layer:
                    row_issues.append(
                        f"negative fixture produced {layer_name} layer from {layer.get('evidence_type')} evidence"
                    )

        state.architecture_rows.append({
            "id": row_id,
            "name": fixture_row["name"],
            "expected_result": fixture_row["expected_result"],
            "status": "fail" if row_issues else "pass",
            "issues": row_issues
        })

        for issue in row_issues:
            state.fail(gate_name, row_id, issue)

    state.gate_status[gate_name] = not any(row["status"] == "fail" for row in state.architecture_rows)


def evaluate_metrics(state: ValidationState) -> None:
    gate_name = "Gate 4 Metric Integrity"
    results = index_by_id(state.candidate.get("metric_results", []))

    for fixture_row in state.fixture.get("metric_integrity", []):
        row_id = fixture_row["id"]
        result = results.get(row_id)
        row_issues: list[str] = []

        if not result:
            row_issues.append("candidate output missing metric audit row")
        else:
            for field_name in ("purpose", "limitations", "anti_gaming_checks"):
                if not result.get(field_name):
                    row_issues.append(f"metric audit missing {field_name}")

            metric = fixture_row["metric"]
            if metric == "refusal_accuracy":
                paired = set(result.get("paired_checks", []))
                if result.get("used_as_acceptance_pass") and not {"false_refusal_rate", "false_answer_rate"}.issubset(paired):
                    row_issues.append("refusal_accuracy used for acceptance without false-refusal and false-answer pairing")

            if metric == "answer_relevance_accuracy":
                if result.get("used_as_acceptance_pass") and int(result.get("unlabeled_rows_count", 0)) > 0:
                    row_issues.append("unlabeled rows counted toward answer relevance acceptance")
                if int(result.get("partial_relevance_accepted_rows", 0)) > 0:
                    row_issues.append("partial relevance counted as full acceptance pass")

            if metric == "section_support_coverage":
                if result.get("used_as_acceptance_pass"):
                    row_issues.append("section_support_coverage used as acceptance pass")
                if result.get("used_as_grounding"):
                    row_issues.append("section_support_coverage used as grounding")

            if metric == "claim_grounding_coverage":
                mappings = result.get("claim_grounding_mappings") or result.get("claim_to_citation_mappings") or []
                if result.get("status") == "unavailable" and result.get("used_as_acceptance_pass"):
                    row_issues.append("claim_grounding_coverage unavailable but treated as pass")
                if not result.get("claim_to_citation_labels_present"):
                    row_issues.append("claim grounding lacks claim-to-citation labels")
                if not result.get("citation_line_ranges_present"):
                    row_issues.append("claim grounding lacks citation line ranges")
                if result.get("used_as_acceptance_pass") and not mappings:
                    row_issues.append("claim grounding lacks claim-to-citation mappings")
                for index, mapping in enumerate(mappings):
                    missing_fields = [
                        field_name for field_name in ("claim", "citation_id", "file", "evidence", "line_range")
                        if not mapping.get(field_name)
                    ]
                    if missing_fields:
                        row_issues.append(
                            f"claim grounding mapping {index} missing fields: {missing_fields}"
                        )

            if metric == "architecture_layer_detection":
                required_checks = {"source_code_evidence_required", "docs_only_negative_fixture", "directory_only_negative_fixture"}
                checks = set(result.get("anti_gaming_checks", []))
                if not required_checks.issubset(checks):
                    row_issues.append("architecture metric lacks required anti-gaming fixtures")

        state.metric_rows.append({
            "id": row_id,
            "metric": fixture_row["metric"],
            "status": "fail" if row_issues else "pass",
            "issues": row_issues
        })

        for issue in row_issues:
            state.fail(gate_name, row_id, issue)

    state.gate_status[gate_name] = not any(row["status"] == "fail" for row in state.metric_rows)


def required_fixture_fields(row: dict[str, Any], kind: str) -> list[str]:
    common = ["id", "question", "required_evidence_groups", "expected_files", "expected_symbols", "expected_answer_facts", "claim_labels", "reviewer_rationale"]
    if kind == "refusal":
        return ["id", "question", "expected_decision", "required_evidence_groups", "forbidden_evidence_shortcuts", "expected_refusal_reason", "claim_labels", "reviewer_rationale"]
    if kind == "relevance":
        return common
    if kind == "architecture":
        return ["id", "name", "expected_result", "reviewer_rationale"]
    if kind == "metric":
        return ["id", "metric", "purpose", "limitations", "anti_gaming_checks"]
    return ["id"]


def evaluate_label_completeness(state: ValidationState) -> None:
    gate_name = "Gate 5 Benchmark Label Completeness"
    rows_to_check: list[tuple[str, dict[str, Any], str]] = []
    rows_to_check.extend(("refusal", row, row["id"]) for row in state.fixture.get("refusal_matrix", []))
    rows_to_check.extend(("relevance", row, row["id"]) for row in state.fixture.get("answer_relevance", []))
    rows_to_check.extend(("architecture", row, row["id"]) for row in state.fixture.get("architecture_fixtures", []))
    rows_to_check.extend(("metric", row, row["id"]) for row in state.fixture.get("metric_integrity", []))

    candidate_labels = index_by_id((state.candidate.get("benchmark_metadata") or {}).get("rows", []))

    for kind, row, row_id in rows_to_check:
        issues: list[str] = []
        for field_name in required_fixture_fields(row, kind):
            if field_name not in row:
                issues.append(f"fixture missing required label {field_name}")

        metadata = candidate_labels.get(row_id)
        if not metadata:
            issues.append("candidate benchmark metadata missing row")
        elif metadata.get("labels_complete") is not True:
            issues.append("candidate benchmark metadata marks labels incomplete")

        state.label_rows.append({
            "id": row_id,
            "kind": kind,
            "status": "fail" if issues else "pass",
            "issues": issues
        })

        for issue in issues:
            state.fail(gate_name, row_id, issue)

    state.gate_status[gate_name] = not any(row["status"] == "fail" for row in state.label_rows)


def evaluate_adversarial_benchmark(state: ValidationState) -> None:
    gate_name = "Gate 6 Adversarial Benchmark"
    adversarial_ids = {
        row.get("id") for row in state.fixture.get("refusal_matrix", [])
    } | {
        row.get("id") for row in state.fixture.get("answer_relevance", [])
    } | {
        row.get("id") for row in state.fixture.get("architecture_fixtures", [])
    } | {
        row.get("id") for row in state.fixture.get("metric_integrity", [])
    }
    categories = {row.get("category") for row in state.fixture.get("refusal_matrix", [])}
    required_categories = {"false_refusal_catcher", "false_answer_catcher"}
    missing_categories = required_categories - categories
    if missing_categories:
        state.fail(gate_name, "suite", f"adversarial benchmark missing categories: {sorted(missing_categories)}")

    if not state.fixture.get("answer_relevance"):
        state.fail(gate_name, "suite", "adversarial benchmark missing answer relevance catchers")
    if not state.fixture.get("architecture_fixtures"):
        state.fail(gate_name, "suite", "adversarial benchmark missing architecture catchers")
    if not state.fixture.get("metric_integrity"):
        state.fail(gate_name, "suite", "adversarial benchmark missing metric catchers")

    behavior_issues = [
        issue for issue in state.issues
        if issue.gate != gate_name and issue.row_id in adversarial_ids
    ]
    if behavior_issues:
        state.fail(
            gate_name,
            "suite",
            f"adversarial behavior failures detected: {len(behavior_issues)} evaluator issues"
        )

    state.gate_status[gate_name] = not any(issue.gate == gate_name for issue in state.issues)


def validate(fixture: dict[str, Any], candidate: dict[str, Any]) -> ValidationState:
    state = ValidationState(fixture=fixture, candidate=candidate)
    evaluate_refusal_matrix(state)
    evaluate_answer_relevance(state)
    evaluate_architecture(state)
    evaluate_metrics(state)
    evaluate_label_completeness(state)
    evaluate_adversarial_benchmark(state)
    return state


def state_to_dict(state: ValidationState) -> dict[str, Any]:
    return {
        "passed": state.passed,
        "gate_status": state.gate_status,
        "issues": [issue.__dict__ for issue in state.issues],
        "refusal_rows": state.refusal_rows,
        "relevance_rows": state.relevance_rows,
        "architecture_rows": state.architecture_rows,
        "metric_rows": state.metric_rows,
        "label_rows": state.label_rows,
    }


def md_escape(value: Any) -> str:
    text = str(value)
    return text.replace("|", "\\|").replace("\n", " ")


def issue_table(issues: list[RowIssue]) -> str:
    if not issues:
        return "No failures.\n"
    lines = ["| Gate | Row | Severity | Message |", "| --- | --- | --- | --- |"]
    for issue in issues:
        lines.append(f"| {md_escape(issue.gate)} | {md_escape(issue.row_id)} | {issue.severity} | {md_escape(issue.message)} |")
    return "\n".join(lines) + "\n"


def write_false_refusal_report(state: ValidationState, output_dir: Path) -> None:
    rows = [row for row in state.refusal_rows if row["category"] == "false_refusal_catcher"]
    false_rows = [row for row in rows if row["status"] == "fail"]
    lines = [
        "# False Refusal Report",
        "",
        f"Rows: {len(rows)}",
        f"False refusals: {len(false_rows)}",
        "",
        "| ID | Expected | Actual | Status | Issues |",
        "| --- | --- | --- | --- | --- |"
    ]
    for row in rows:
        lines.append(
            f"| {row['id']} | {row['expected_decision']} | {row['actual_decision']} | {row['status']} | {md_escape('; '.join(row['issues']))} |"
        )
    (output_dir / "false-refusal-report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def write_false_answer_report(state: ValidationState, output_dir: Path) -> None:
    rows = [row for row in state.refusal_rows if row["category"] == "false_answer_catcher"]
    false_rows = [row for row in rows if row["status"] == "fail"]
    lines = [
        "# False Answer Report",
        "",
        f"Rows: {len(rows)}",
        f"False answers: {len(false_rows)}",
        "",
        "| ID | Expected | Actual | Status | Issues |",
        "| --- | --- | --- | --- | --- |"
    ]
    for row in rows:
        lines.append(
            f"| {row['id']} | {row['expected_decision']} | {row['actual_decision']} | {row['status']} | {md_escape('; '.join(row['issues']))} |"
        )
    (output_dir / "false-answer-report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def write_answer_relevance_report(state: ValidationState, output_dir: Path) -> None:
    lines = [
        "# Answer Relevance Report",
        "",
        "| ID | Classification | Expected Files | Actual Files | Issues |",
        "| --- | --- | --- | --- | --- |"
    ]
    for row in state.relevance_rows:
        lines.append(
            f"| {row['id']} | {row['classification']} | {md_escape(', '.join(row['expected_files']))} | {md_escape(', '.join(row['actual_files']))} | {md_escape('; '.join(row['issues']))} |"
        )
    (output_dir / "answer-relevance-report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def architecture_expectation(row: dict[str, Any]) -> str:
    if row.get("expected_result") == "pass":
        return "source-code layers required"
    return "negative fixture must not detect layers"


def write_architecture_report(state: ValidationState, output_dir: Path) -> None:
    lines = [
        "# Architecture Validation Report",
        "",
        "| ID | Fixture | Fixture Expectation | Validation Result | Issues |",
        "| --- | --- | --- | --- | --- |"
    ]
    fixture_rows = index_by_id(state.fixture.get("architecture_fixtures", []))
    for row in state.architecture_rows:
        fixture_row = fixture_rows.get(row["id"], {})
        lines.append(
            f"| {row['id']} | {row['name']} | {architecture_expectation(fixture_row)} | {row['status']} | {md_escape('; '.join(row['issues']))} |"
        )
    (output_dir / "architecture-validation-report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def write_metric_report(state: ValidationState, output_dir: Path) -> None:
    lines = [
        "# Metric Validation Report",
        "",
        "| ID | Metric | Status | Issues |",
        "| --- | --- | --- | --- |"
    ]
    for row in state.metric_rows:
        lines.append(
            f"| {row['id']} | {row['metric']} | {row['status']} | {md_escape('; '.join(row['issues']))} |"
        )
    (output_dir / "metric-validation-report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def write_review_report(state: ValidationState, output_dir: Path) -> None:
    gate_lines = ["| Gate | Status |", "| --- | --- |"]
    for gate_name in [
        "Gate 1 Refusal Matrix",
        "Gate 2 Answer Relevance",
        "Gate 3 Architecture Evidence",
        "Gate 4 Metric Integrity",
        "Gate 5 Benchmark Label Completeness",
        "Gate 6 Adversarial Benchmark"
    ]:
        gate_lines.append(f"| {gate_name} | {'pass' if state.gate_status.get(gate_name) else 'fail'} |")

    false_refusal_count = sum(1 for row in state.refusal_rows if row["category"] == "false_refusal_catcher")
    false_answer_count = sum(1 for row in state.refusal_rows if row["category"] == "false_answer_catcher")
    lines = [
        "# Validation Framework Review",
        "",
        "## Implemented Gates",
        "",
        *gate_lines,
        "",
        "## Fixture Counts",
        "",
        f"- Refusal matrix rows: {len(state.fixture.get('refusal_matrix', []))}",
        f"- False refusal catchers: {false_refusal_count}",
        f"- False answer catchers: {false_answer_count}",
        f"- Answer relevance rows: {len(state.fixture.get('answer_relevance', []))}",
        f"- Architecture fixtures: {len(state.fixture.get('architecture_fixtures', []))}",
        f"- Metric audit rows: {len(state.fixture.get('metric_integrity', []))}",
        "",
        "## Benchmark Counts",
        "",
        f"- Candidate refusal rows: {len(state.candidate.get('refusal_results', []))}",
        f"- Candidate answer relevance rows: {len(state.candidate.get('answer_relevance_results', []))}",
        f"- Candidate architecture rows: {len(state.candidate.get('architecture_results', []))}",
        f"- Candidate metric rows: {len(state.candidate.get('metric_results', []))}",
        "",
        "## Passing Examples",
        "",
        "- RF-001: answerable RAG retrieval question requires retrieval/RAG evidence.",
        "- FA-001: unsupported revenue API question must refuse despite API path matches.",
        "- ARCH-NEG-001: README-only architecture evidence must not pass.",
        "- MET-004: claim grounding cannot pass when claim-to-citation labels are unavailable.",
        "",
        "## Failing Examples",
        "",
        issue_table(state.issues),
        "## Evaluator Authority",
        "",
        "- Gate statuses are derived from evaluator issues.",
        "- Reports are generated from evaluator state and checked for verdict consistency.",
        "- Review text is not allowed to override evaluator pass/fail.",
        "",
        "## Coverage",
        "",
        "- Refusal decision matrix covers answerable acronym questions and unsupported business/external/prompt-injection questions.",
        "- Answer relevance covers expected files, symbols, evidence groups, and answer facts.",
        "- Architecture validation covers source-code positive, docs-only negative, and directory-only negative fixtures.",
        "- Metric validation covers metric purpose, limitations, anti-gaming checks, grounding availability, section-support misuse, and label completeness.",
        "",
        "## Result",
        "",
        "pass" if state.passed else "fail"
    ]
    (output_dir / "validation-framework-review.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def write_state_json(state: ValidationState, output_dir: Path) -> None:
    (output_dir / "validation-state.json").write_text(
        json.dumps(state_to_dict(state), indent=2) + "\n",
        encoding="utf-8"
    )


def write_reports(state: ValidationState, output_dir: Path) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    write_state_json(state, output_dir)
    write_false_refusal_report(state, output_dir)
    write_false_answer_report(state, output_dir)
    write_answer_relevance_report(state, output_dir)
    write_architecture_report(state, output_dir)
    write_metric_report(state, output_dir)
    write_review_report(state, output_dir)


def parse_review_result(markdown: str) -> str | None:
    lines = [line.strip() for line in markdown.splitlines()]
    for index, line in enumerate(lines):
        if line == "## Result":
            for value in lines[index + 1:]:
                if value:
                    return value.lower()
    return None


def parse_review_gate_status(markdown: str) -> dict[str, bool]:
    statuses: dict[str, bool] = {}
    for line in markdown.splitlines():
        parts = [part.strip() for part in line.strip().strip("|").split("|")]
        if len(parts) == 2 and parts[0].startswith("Gate "):
            statuses[parts[0]] = parts[1].lower() == "pass"
    return statuses


def validate_report_consistency(state: ValidationState, output_dir: Path) -> list[str]:
    issues: list[str] = []
    state_path = output_dir / "validation-state.json"
    review_path = output_dir / "validation-framework-review.md"
    if not state_path.exists():
        issues.append("validation-state.json missing")
    else:
        raw_state = json.loads(state_path.read_text(encoding="utf-8"))
        if raw_state.get("passed") is not state.passed:
            issues.append("raw state verdict differs from evaluator verdict")
        if raw_state.get("gate_status") != state.gate_status:
            issues.append("raw state gate statuses differ from evaluator gate statuses")

    if not review_path.exists():
        issues.append("validation-framework-review.md missing")
    else:
        review = review_path.read_text(encoding="utf-8")
        review_result = parse_review_result(review)
        expected_result = "pass" if state.passed else "fail"
        if review_result != expected_result:
            issues.append(f"review verdict {review_result!r} differs from evaluator verdict {expected_result!r}")
        review_gates = parse_review_gate_status(review)
        for gate_name, expected in state.gate_status.items():
            if review_gates.get(gate_name) is not expected:
                issues.append(f"review gate {gate_name!r} differs from evaluator status")

    return issues


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run Knowledge Forge acceptance validation gates.")
    parser.add_argument("--fixtures", type=Path, required=True, help="Path to acceptance-suite.json.")
    parser.add_argument("--candidate", type=Path, required=True, help="Path to candidate output JSON.")
    parser.add_argument("--output", type=Path, required=True, help="Directory for markdown validation reports.")
    return parser.parse_args(argv)


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    fixture = load_json(args.fixtures)
    candidate = load_json(args.candidate)
    state = validate(fixture, candidate)
    write_reports(state, args.output)
    consistency_issues = validate_report_consistency(state, args.output)

    if state.passed and not consistency_issues:
        print(f"acceptance validation passed; reports written to {args.output}")
        return 0

    print(f"acceptance validation failed; reports written to {args.output}", file=sys.stderr)
    for issue in state.issues:
        print(f"{issue.gate} [{issue.row_id}]: {issue.message}", file=sys.stderr)
    for issue in consistency_issues:
        print(f"Report Consistency: {issue}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
