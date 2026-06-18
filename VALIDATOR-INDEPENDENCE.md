# Validator Independence

Branch: `tgusain/m17-product-conformance-r2`

Purpose: prove the product remediation does not depend on validator fixtures, benchmark labels, candidate JSON, or validation outputs.

## Searched Paths

Runtime/product paths:

```text
cmd
internal
ui
deploy
```

Repository-wide search excluding validator fixtures:

```text
.
excluding eval-runner/acceptance/**
excluding node_modules, dist, and .git
```

## Searched Terms

```text
eval-runner/acceptance
acceptance-suite
passing-candidate
red-team-repeat-candidate
validation-state
validator-reports
RF-004
RF-005
AR-004
```

## Product Runtime Search Result

Command:

```bash
rg -n "eval-runner/acceptance|acceptance-suite|passing-candidate|red-team-repeat-candidate|validation-state|validator-reports|RF-004|RF-005|AR-004" cmd internal ui deploy --glob '!**/node_modules/**' --glob '!**/dist/**'
```

Result:

```text
No matches.
```

## Repository Search Result

Command:

```bash
rg -n "eval-runner/acceptance|acceptance-suite|passing-candidate|red-team-repeat-candidate|validation-state|validator-reports|RF-004|RF-005|AR-004" . --glob '!eval-runner/acceptance/**' --glob '!**/node_modules/**' --glob '!**/dist/**' --glob '!**/.git/**'
```

Result:

```text
Makefile:37: python3 eval-runner/acceptance/validation_runner.py
Makefile:38: --fixtures eval-runner/acceptance/fixtures/acceptance-suite.json
Makefile:39: --candidate eval-runner/acceptance/candidates/passing-candidate.json
Makefile:40: --output eval-runner/acceptance/reports
Makefile:41: python3 -m unittest discover eval-runner/acceptance -p 'test_*.py'
```

Interpretation:

- The only repository-wide matches outside the validator directory are the `make validate-acceptance` command wiring.
- Product runtime code under `cmd`, `internal`, `ui`, and `deploy` has no references to validator fixtures, validator candidates, row IDs, report paths, or validation outputs.

## Conclusion

Product behavior is derived from:

- Repository retrieval hits.
- Evidence groups inferred from hit metadata, paths, and content.
- Support-gate metadata.
- Repository Q&A policy and context token budget.

Product behavior is not derived from:

- Validator fixtures.
- Validator candidates.
- Benchmark labels.
- Validation outputs.
- Previous audit artifacts.

Validator independence conclusion:

```text
PASS
```
