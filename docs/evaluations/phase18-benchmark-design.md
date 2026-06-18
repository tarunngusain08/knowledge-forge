# Phase 18 Benchmark Design

Phase 18 measures whether Knowledge Forge retrieves and grounds repository
evidence better than simple baselines. It does not introduce new product
capability, new retrieval systems, or another audit framework.

## Corpus

The benchmark uses only the existing synthetic enterprise monolith fixture:

```text
eval-runner/fixtures/synthetic-enterprise-monolith
```

The frozen benchmark file contains exactly 30 rows:

| Category | Count | Purpose |
| --- | ---: | --- |
| `architecture_implementation` | 10 | Measures implementation-location and architecture-understanding questions. |
| `dependency_impact_testing` | 8 | Measures dependency, impact, and test-evidence questions. |
| `unsupported_refusal_prompt_injection` | 7 | Measures refusal precision and recall. |
| `deep_dive_grounding_architecture_evidence` | 5 | Measures evidence-backed architecture and deep-dive grounding questions. |

## Row Schema

Every row includes:

- `id`
- `category`
- `question`
- `expected_files`
- `expected_symbols`
- `expected_answer_facts`
- `required_evidence_groups`
- `should_refuse`

`expected_line_ranges` are optional and are not required for Phase 18.

## Benchmark Authoring Gate

The benchmark corpus is authored before any baseline or Knowledge Forge result
is generated.

Integrity review:

- category counts are `10/8/7/5`
- labels are complete
- evidence groups are present for answerable rows
- refusal labels are reviewed
- `expected_line_ranges` are not required

Once the first benchmark result is generated, labels, expected facts, evidence
groups, and refusal expectations are frozen unless a documented labeling defect
is discovered and preserved in git history.

## Baselines

Phase 18 compares Knowledge Forge with two baselines:

| Baseline | Definition |
| --- | --- |
| Keyword search | Simple lexical/file-content matching over the fixture. |
| Retrieval only | Retrieved files and evidence without grounded answer generation. |

Baseline independence rule:

1. Generate keyword baseline.
2. Generate retrieval-only baseline.
3. Preserve baseline outputs.
4. Generate Knowledge Forge candidate output.
5. Do not modify baseline outputs after candidate results exist.

## Metrics

Primary metrics:

- retrieval recall / file coverage
- evidence recall
- answerable-question accuracy
- refusal precision
- refusal recall
- grounding coverage

Secondary metrics:

- MRR
- citation accuracy when optional line-range labels exist
- average latency
- average cost

Per-category breakdowns are required so aggregate scores do not hide weak areas.

## Material Improvement Definition

A category is improved only when Knowledge Forge beats both baselines by:

```text
>= 10% relative improvement
```

or:

```text
at least 3 additional benchmark rows answered correctly
```

Otherwise the category is `UNCHANGED`. If Knowledge Forge underperforms a
baseline, the category is `DEGRADED`.

## Failure Criteria

Phase 18 may be inconclusive.

If Knowledge Forge does not materially outperform both baselines in at least one
major category, the proof report must state:

```text
Inconclusive
```

The report must not manufacture Phase 19 justification.

## Limitations

- The corpus is synthetic and deliberately small.
- Saved result files are acceptable for this proof pack; live API runs remain
  optional and environment-gated.
- The benchmark proves relative behavior on the fixture, not enterprise-scale
  production quality.
- Phase 19 static intelligence is justified only if the metrics expose a
  measured weakness it can plausibly fix.
