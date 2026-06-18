# Grounding Proof Results

## Implemented Checks

Claim-grounding booleans are no longer sufficient.

For `claim_grounding_coverage` used as an acceptance pass, the validator requires `claim_grounding_mappings` or `claim_to_citation_mappings` with each mapping containing:

- `claim`
- `citation_id`
- `file`
- `line_range`
- `evidence`

## New Fixture

`eval-runner/acceptance/candidates/grounding-bypass-candidate.json`

This candidate sets grounding booleans to pass but omits raw mappings.

## Proof

The hardened validator rejects it with:

```text
Gate 4 Metric Integrity [MET-004]: claim grounding lacks claim-to-citation mappings
Gate 6 Adversarial Benchmark [suite]: adversarial behavior failures detected: 1 evaluator issues
```

The canonical passing candidate now includes concrete claim-to-citation mappings under `MET-004`.

## Result

Pass. The booleans-only claim-grounding bypass is eliminated.
