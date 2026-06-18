# Symbol Proof Results

## Implemented Checks

Answer relevance now enforces `expected_symbols` from the benchmark fixtures.

A candidate must provide the required symbols through either:

- `actual_symbols`
- `cited_symbols`

Missing required symbols fail answer relevance.

## New Fixture

`eval-runner/acceptance/candidates/symbol-bypass-candidate.json`

This candidate keeps expected files, groups, and facts but removes all symbol evidence.

## Proof

The hardened validator rejects it with missing-symbol failures for all four answer relevance rows:

```text
Gate 2 Answer Relevance [AR-001]: missing expected symbols: ['JWTManager', 'Middleware']
Gate 2 Answer Relevance [AR-002]: missing expected symbols: ['Connect', 'JWTManager']
Gate 2 Answer Relevance [AR-003]: missing expected symbols: ['CodeRetrievalService', 'FuseRRF']
Gate 2 Answer Relevance [AR-004]: missing expected symbols: ['DeepDiveReport', 'ReportSection']
Gate 6 Adversarial Benchmark [suite]: adversarial behavior failures detected: 4 evaluator issues
```

## Result

Pass. The missing-symbol-evidence bypass is eliminated.
