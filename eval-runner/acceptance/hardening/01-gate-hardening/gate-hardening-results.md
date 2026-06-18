# Gate Hardening Results

## Implemented Checks

- Gate statuses now derive from evaluator issues, not from narrow summary counters.
- Gate 1 fails when any Gate 1 evaluator issue exists, including support-trace and evidence-group failures.
- Gate 6 now validates adversarial behavior by checking evaluator issues for adversarial fixture rows.
- Final CLI verdict still derives from `state.passed`.
- Reports are consistency-checked against evaluator state after generation.

## Proof

`bad-missing-support-trace` previously produced Gate 1 issues while Gate 1 still reported pass. The hardened test `test_gate_status_is_derived_from_gate_issues` now asserts:

- `state.passed == false`
- `Gate 1 Refusal Matrix == false`
- `Gate 6 Adversarial Benchmark == false`

`red-team-repeat-candidate.json` now fails Gate 6 with:

```text
Gate 6 Adversarial Benchmark [suite]: adversarial behavior failures detected: 58 evaluator issues
```

## Result

Pass. Gate-level summaries no longer override evaluator failures.
