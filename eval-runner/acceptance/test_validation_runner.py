import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "fixtures" / "acceptance-suite.json"
PASSING_CANDIDATE = ROOT / "candidates" / "passing-candidate.json"
RED_TEAM_CANDIDATE = ROOT / "candidates" / "red-team-repeat-candidate.json"


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
        self.assertIn("negative fixture produced High-confidence", messages)
        self.assertIn("section_support_coverage used as acceptance pass", messages)
        self.assertIn("claim_grounding_coverage unavailable but treated as pass", messages)
        self.assertIn("candidate benchmark metadata marks labels incomplete", messages)

    def test_reports_are_generated_for_passing_candidate(self):
        candidate = self.runner.load_json(PASSING_CANDIDATE)
        state = self.runner.validate(self.fixture, candidate)
        with tempfile.TemporaryDirectory() as tempdir:
            output = Path(tempdir)
            self.runner.write_reports(state, output)
            expected = {
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


if __name__ == "__main__":
    unittest.main()
