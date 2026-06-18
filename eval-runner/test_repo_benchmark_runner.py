from pathlib import Path

from repo_benchmark_runner import category_metrics, category_outcomes, comparison_summary, latency_ms, metrics, write_markdown


def test_repository_metrics_include_failure_quality() -> None:
    output = metrics(
        [
            {
                "question": "Where is auth implemented?",
                "expected_files": ["internal/auth/service.go", "cmd/api/main.go"],
                "retrieved_files": ["internal/auth/service.go"],
                "expected_symbols": ["AuthService"],
                "retrieved_symbols": ["AuthService"],
                "required_evidence_groups": ["auth_service"],
                "retrieved_evidence_groups": ["auth_service"],
                "expected_answer_facts": ["AuthService owns login behavior"],
                "matched_answer_facts": ["AuthService owns login behavior"],
                "expected_line_ranges": [{"path": "internal/auth/service.go", "start_line": 10, "end_line": 30}],
                "citation_line_ranges": [{"path": "internal/auth/service.go", "start_line": 20, "end_line": 25}],
                "supported_claim_count": 3,
                "total_claim_count": 4,
                "latency_ms": 100,
                "cost_usd": 0.01,
            },
            {
                "question": "What handles payroll?",
                "should_refuse": True,
                "refused": True,
                "latency_ms": 50,
                "cost_usd": 0.005,
            },
        ]
    )

    assert output["question_count"] == 2
    assert output["file_hit_rate"] == 0.5
    assert output["retrieval_recall"] == 0.75
    assert output["evidence_recall"] == 1.0
    assert output["file_coverage"] == 0.75
    assert output["symbol_coverage"] == 1.0
    assert output["line_range_accuracy"] == 1.0
    assert output["refusal_precision"] == 1.0
    assert output["refusal_recall"] == 1.0
    assert output["refusal_accuracy"] == 1.0
    assert output["answerable_question_accuracy"] == 1.0
    assert output["grounding_coverage"] == 0.875


def test_comparison_summary_reports_quality_latency_cost_and_stages() -> None:
    baseline = [
        {
            "question": "Where is auth implemented?",
            "expected_files": ["internal/auth/service.go"],
            "retrieved_files": ["cmd/api/main.go"],
            "required_evidence_groups": ["auth_service"],
            "retrieved_evidence_groups": [],
            "expected_answer_facts": ["AuthService owns login behavior"],
            "matched_answer_facts": [],
            "supported_claim_count": 1,
            "total_claim_count": 2,
            "latency_ms": 100,
            "cost_usd": 0.01,
            "stage_contributions": {"dense": 1},
        }
    ]
    candidate = [
        {
            "question": "Where is auth implemented?",
            "expected_files": ["internal/auth/service.go"],
            "retrieved_files": ["internal/auth/service.go"],
            "required_evidence_groups": ["auth_service"],
            "retrieved_evidence_groups": ["auth_service"],
            "expected_answer_facts": ["AuthService owns login behavior"],
            "matched_answer_facts": ["AuthService owns login behavior"],
            "supported_claim_count": 2,
            "total_claim_count": 2,
            "latency_ms": 130,
            "cost_usd": 0.012,
            "stage_contributions": {"dense": 1, "symbol": 1},
        }
    ]

    output = comparison_summary(baseline, candidate)

    assert len(output["questions_improved"]) == 1
    assert output["metric_delta"]["file_coverage"] > 0
    assert output["metric_delta"]["evidence_recall"] > 0
    assert output["metric_delta"]["avg_latency_ms"] == 30
    assert output["stage_contribution_delta"]["symbol"] == 1


def test_latency_prefers_explicit_milliseconds_and_converts_legacy_nanoseconds() -> None:
    assert latency_ms({"latency_ms": 623.18, "latency": 999999999}) == 623.18
    assert latency_ms({"latency": 623180}) == 0.62318


def test_category_metrics_and_outcomes_compare_against_strongest_baseline() -> None:
    candidate = [
        {
            "id": "q1",
            "category": "architecture_implementation",
            "question": "Where is auth?",
            "expected_files": ["internal/auth/service.go"],
            "retrieved_files": ["internal/auth/service.go"],
            "required_evidence_groups": ["auth_service"],
            "retrieved_evidence_groups": ["auth_service"],
            "expected_answer_facts": ["AuthService owns login behavior"],
            "matched_answer_facts": ["AuthService owns login behavior"],
            "supported_claim_count": 1,
            "total_claim_count": 1,
        }
    ]
    weak = [
        {
            "id": "q1",
            "category": "architecture_implementation",
            "question": "Where is auth?",
            "expected_files": ["internal/auth/service.go"],
            "retrieved_files": [],
            "required_evidence_groups": ["auth_service"],
            "retrieved_evidence_groups": [],
            "expected_answer_facts": ["AuthService owns login behavior"],
            "matched_answer_facts": [],
            "supported_claim_count": 0,
            "total_claim_count": 1,
        }
    ]

    assert category_metrics(candidate)["architecture_implementation"]["correct_count"] == 1
    outcomes = category_outcomes({"keyword": weak, "retrieval_only": weak}, candidate)

    assert outcomes[0]["category"] == "architecture_implementation"
    assert outcomes[0]["outcome"] == "IMPROVED"
    assert outcomes[0]["correct_delta"] == 1


def test_markdown_report_writes_metrics(tmp_path: Path) -> None:
    output = {
        "metrics": {"question_count": 1, "retrieval_recall": 1.0, "correctness_rate": 1.0},
        "category_metrics": {"architecture_implementation": {"question_count": 1, "retrieval_recall": 1.0}},
        "baseline_comparisons": {"keyword": {"metric_delta": {"retrieval_recall": 1.0}, "questions_improved": [], "questions_unchanged": [], "questions_degraded": []}},
        "category_outcomes": [{"category": "architecture_implementation", "outcome": "UNCHANGED", "best_baseline": "keyword", "correct_delta": 1, "primary_metric_delta": 1.0}],
    }
    report = tmp_path / "report.md"

    write_markdown(report, output)

    text = report.read_text()
    assert "# Phase 18 Benchmark Report" in text
    assert "Category Outcomes" in text
