# Adversarial Proof Results

## Implemented Checks

Gate 6 is no longer only a benchmark-structure gate.

It now fails when evaluator issues are present for adversarial fixture rows across:

- refusal matrix
- answer relevance
- architecture fixtures
- metric integrity fixtures

## Proof

`red-team-repeat-candidate.json` fails with 58 evaluator issues and Gate 6 now fails explicitly:

```text
Gate 6 Adversarial Benchmark [suite]: adversarial behavior failures detected: 58 evaluator issues
```

Previously, this candidate failed overall while Gate 6 still reported pass.

## Result

Pass. Known red-team failures now fail Gate 6 automatically.
