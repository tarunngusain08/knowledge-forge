import copy
import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "fixtures" / "acceptance-suite.json"
PASSING_CANDIDATE = ROOT / "candidates" / "passing-candidate.json"
RED_TEAM_CANDIDATE = ROOT / "candidates" / "red-team-repeat-candidate.json"
GROUNDING_BYPASS_CANDIDATE = ROOT / "candidates" / "grounding-bypass-candidate.json"
SYMBOL_BYPASS_CANDIDATE = ROOT / "candidates" / "symbol-bypass-candidate.json"


def load_runner():
    spec = importlib.util.spec_from_file_location("validation_runner", ROOT / "validation_runner.py")
    module = importlib.util.module_from_spec(spec)
    assert spec and spec.loader
    sys.modules["validation_runner"] = module
    spec.loader.exec_module(module)
    return module


class ValidationRunnerTest(unittest.TestCase):
    def setUp(self):
        self.runner = load_runner()
        self.fixture = self.runner.load_json(FIXTURES)

    def test_passing_candidate_satisfies_all_gates(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        state = self.runner.validate(self.fixture, candidate)

        self.assertTrue(state.passed, [issue.message for issue in state.issues])
        self.assertTrue(all(state.gate_status.values()))
        self.assertEqual(11, len(state.refusal_rows))
        self.assertEqual(4, len(state.relevance_rows))
        self.assertEqual(3, len(state.architecture_rows))
        self.assertEqual(5, len(state.metric_rows))

    def test_red_team_repeat_candidate_is_rejected(self):
        candidate = self.runner.load_json(RED_TEAM_CANDIDATE)
        state = self.runner.validate(self.fixture, candidate)

        self.assertFalse(state.passed)
        messages = "\n".join(issue.message for issue in state.issues)
        self.assertIn("answerable row was refused", messages)
        self.assertIn("unsupported row was answered", messages)
        self.assertIn("missing required evidence groups", messages)
        self.assertIn("negative fixture produced api layer from doc evidence", messages)
        self.assertIn("section_support_coverage used as acceptance pass", messages)
        self.assertIn("claim_grounding_coverage unavailable but treated as pass", messages)
        self.assertIn("candidate benchmark metadata marks labels incomplete", messages)
        self.assertIn("adversarial behavior failures detected", messages)
        self.assertFalse(state.gate_status["Gate 6 Adversarial Benchmark"])

    def test_reports_are_generated_for_passing_candidate(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        state = self.runner.validate(self.fixture, candidate)
        with tempfile.TemporaryDirectory() as tempdir:
            output = Path(tempdir)
            self.runner.write_reports(state, output)
            expected = {
                "validation-state.json",
                "false-refusal-report.md",
                "false-answer-report.md",
                "answer-relevance-report.md",
                "architecture-validation-report.md",
                "metric-validation-report.md",
                "validation-framework-review.md",
            }
            self.assertEqual(expected, {path.name for path in output.iterdir()})
            review = (output / "validation-framework-review.md").read_text(encoding="utf-8")
            self.assertIn("Gate 1 Refusal Matrix", review)
            self.assertIn("pass", review)
            self.assertEqual([], self.runner.validate_report_consistency(state, output))

    def test_claim_grounding_requires_raw_mappings(self):
        candidate = self.runner.load_json(GROUNDING_BYPASS_CANDIDATE)

        state = self.runner.validate(self.fixture, candidate)

        self.assertFalse(state.passed)
        self.assertFalse(state.gate_status["Gate 4 Metric Integrity"])
        messages = "\n".join(issue.message for issue in state.issues)
        self.assertIn("claim grounding lacks claim-to-citation mappings", messages)

    def test_expected_symbols_are_enforced(self):
        candidate = self.runner.load_json(SYMBOL_BYPASS_CANDIDATE)

        state = self.runner.validate(self.fixture, candidate)

        self.assertFalse(state.passed)
        self.assertFalse(state.gate_status["Gate 2 Answer Relevance"])
        messages = "\n".join(issue.message for issue in state.issues)
        self.assertIn("missing expected symbols", messages)

    def test_negative_architecture_rejects_any_detected_docs_layer(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        candidate = copy.deepcopy(candidate)
        for row in candidate["architecture_results"]:
            if row["fixture_id"] == "ARCH-NEG-001":
                row["layers"] = [{
                    "name": "api",
                    "confidence": "Low",
                    "evidence_type": "doc",
                    "files": ["cmd/api/README.md"],
                    "packages": [],
                    "line_ranges": [],
                    "sufficiency": "README path says API"
                }]

        state = self.runner.validate(self.fixture, candidate)

        self.assertFalse(state.passed)
        self.assertFalse(state.gate_status["Gate 3 Architecture Evidence"])
        messages = "\n".join(issue.message for issue in state.issues)
        self.assertIn("negative fixture produced api layer from doc evidence", messages)

    def test_gate_status_is_derived_from_gate_issues(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        candidate = copy.deepcopy(candidate)
        for row in candidate["refusal_results"]:
            if row["id"] == "RF-001":
                row.pop("support_gate", None)

        state = self.runner.validate(self.fixture, candidate)

        self.assertFalse(state.passed)
        self.assertFalse(state.gate_status["Gate 1 Refusal Matrix"])
        self.assertFalse(state.gate_status["Gate 6 Adversarial Benchmark"])

    def test_report_consistency_detects_review_verdict_drift(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        state = self.runner.validate(self.fixture, candidate)
        with tempfile.TemporaryDirectory() as tempdir:
            output = Path(tempdir)
            self.runner.write_reports(state, output)
            review_path = output / "validation-framework-review.md"
            review = review_path.read_text(encoding="utf-8")
            review_path.write_text(review.replace("\npass\n", "\nfail\n", 1), encoding="utf-8")

            issues = self.runner.validate_report_consistency(state, output)

            self.assertTrue(any("review verdict" in issue for issue in issues))


if __name__ == "__main__":
    unittest.main()
