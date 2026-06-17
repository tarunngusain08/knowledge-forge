from repo_benchmark_runner import comparison_summary, latency_ms, metrics


def test_repository_metrics_include_failure_quality() -> None:
    output = metrics(
        [
            {
                "question": "Where is auth implemented?",
                "expected_files": ["internal/auth/service.go", "cmd/api/main.go"],
                "retrieved_files": ["internal/auth/service.go"],
                "expected_symbols": ["AuthService"],
                "retrieved_symbols": ["AuthService"],
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
    assert output["file_coverage"] == 0.75
    assert output["symbol_coverage"] == 1.0
    assert output["line_range_accuracy"] == 1.0
    assert output["refusal_accuracy"] == 1.0
    assert output["answerable_question_accuracy"] == 1.0


def test_comparison_summary_reports_quality_latency_cost_and_stages() -> None:
    baseline = [
        {
            "question": "Where is auth implemented?",
            "expected_files": ["internal/auth/service.go"],
            "retrieved_files": ["cmd/api/main.go"],
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
    assert output["metric_delta"]["avg_latency_ms"] == 30
    assert output["stage_contribution_delta"]["symbol"] == 1


def test_latency_prefers_explicit_milliseconds_and_converts_legacy_nanoseconds() -> None:
    assert latency_ms({"latency_ms": 623.18, "latency": 999999999}) == 623.18
    assert latency_ms({"latency": 623180}) == 0.62318
