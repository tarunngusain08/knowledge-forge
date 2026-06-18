# Consistency Proof Results

## Implemented Checks

The validator now writes raw evaluator state to:

```text
validation-state.json
```

After report generation, `validate_report_consistency` verifies:

- raw state verdict matches evaluator verdict
- raw state gate statuses match evaluator gate statuses
- review package final verdict matches evaluator verdict
- review package gate table matches evaluator gate statuses

CLI exits nonzero if report consistency fails.

## Proof

`test_report_consistency_detects_review_verdict_drift` mutates a generated review verdict after report generation and asserts that consistency validation detects the drift.

The generated review package now includes:

```text
Evaluator Authority
- Gate statuses are derived from evaluator issues.
- Reports are generated from evaluator state and checked for verdict consistency.
- Review text is not allowed to override evaluator pass/fail.
```

## Result

Pass. Report/review verdict drift is now executable validation failure.
