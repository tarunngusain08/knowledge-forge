# Acceptance Framework Hardening Review

## Summary

The acceptance framework was hardened using the meta-validation audit findings as source of truth.

No product code was changed.
No retrieval code was changed.
No repository intelligence code was changed.
No Phase 18 or Phase 19 work was started.

## Implemented Checks

1. Gate Integrity
   - Gate statuses derive from evaluator issues.
   - Gate 1 no longer passes when support-trace or evidence-group issues exist.

2. Claim Grounding Proof
   - Booleans are insufficient.
   - Claim grounding requires claim-to-citation mappings with claim, citation, file, line range, and evidence.

3. Symbol Evidence Enforcement
   - Expected symbols are enforced through actual or cited symbol evidence.

4. Architecture Evidence Enforcement
   - README-only architecture fails.
   - Directory-only architecture fails.
   - Negative fixtures fail on any detected forbidden layer, regardless of confidence.

5. Adversarial Benchmark Gate
   - Gate 6 validates adversarial behavior, not only benchmark structure.

6. Consistency Enforcement
   - Raw evaluator state, generated reports, review package, and final verdict are checked for consistency.

## New Fixtures

- `eval-runner/acceptance/candidates/grounding-bypass-candidate.json`
- `eval-runner/acceptance/candidates/symbol-bypass-candidate.json`
- `claim_grounding_coverage.required_proof_fields` in `acceptance-suite.json`
- `cited_symbols` and `claim_grounding_mappings` in `passing-candidate.json`

## Bypasses Eliminated

- Gate 1 pass despite support-trace issues.
- Gate 6 pass despite red-team behavior failures.
- Booleans-only claim grounding.
- Missing expected symbol evidence.
- Low-confidence README-only architecture layer reporting.
- Ambiguous architecture report semantics.
- Review verdict drift from evaluator verdict.

## Remaining Limitations

- CI still validates candidate artifacts rather than live product-generated outputs.
- The framework proves candidate-output correctness, not product runtime behavior.
- Claim evidence is structurally validated, but citation text truth still depends on benchmark labels and reviewer-quality fixtures.

## Validation Performed

```text
make validate-acceptance
python3 eval-runner/acceptance/validation_runner.py --fixtures eval-runner/acceptance/fixtures/acceptance-suite.json --candidate eval-runner/acceptance/candidates/grounding-bypass-candidate.json --output /tmp/kf-grounding-bypass-hardening
python3 eval-runner/acceptance/validation_runner.py --fixtures eval-runner/acceptance/fixtures/acceptance-suite.json --candidate eval-runner/acceptance/candidates/symbol-bypass-candidate.json --output /tmp/kf-symbol-bypass-hardening
python3 eval-runner/acceptance/validation_runner.py --fixtures eval-runner/acceptance/fixtures/acceptance-suite.json --candidate eval-runner/acceptance/candidates/red-team-repeat-candidate.json --output /tmp/kf-red-team-repeat-hardening
```

## Result

Pass. Every bypass discovered in meta-validation now fails automatically, and report/review verdict drift is checked by executable validation.
