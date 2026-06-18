# Architecture Proof Results

## Implemented Checks

Negative architecture fixtures now fail if they detect forbidden architecture layers at any confidence, not only `High` confidence.

README-only architecture must fail.
Directory-only architecture must fail.
Positive architecture fixtures still require source-code evidence, files, packages, and line ranges.

## Report Hardening

The architecture report no longer uses the ambiguous columns:

```text
Expected = fail
Status = pass
```

It now uses:

```text
Fixture Expectation = negative fixture must not detect layers
Validation Result = pass/fail
```

## Proof

`test_negative_architecture_rejects_any_detected_docs_layer` injects a Low-confidence docs-only API layer into `ARCH-NEG-001`; the validator fails it with:

```text
negative fixture produced api layer from doc evidence
```

`red-team-repeat-candidate.json` also fails README-only and directory-only architecture checks.

## Result

Pass. README-only and directory-only architecture bypasses are eliminated, and report wording no longer contradicts execution semantics.
