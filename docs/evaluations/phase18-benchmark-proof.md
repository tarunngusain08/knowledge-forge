# Phase 18 Benchmark Proof

Phase 18 answers one question:

```text
Does Knowledge Forge retrieve and ground repository evidence better than simple baselines?
```

Result:

```text
Partially Proven
```

Knowledge Forge materially outperformed both baselines in architecture,
dependency/impact, and grounding-oriented categories. It did not materially
outperform the stronger retrieval-only baseline in the unsupported/refusal
category, so the result is not presented as a universal win.

## Benchmark Design Summary

The benchmark uses the existing synthetic enterprise monolith only:

```text
eval-runner/fixtures/synthetic-enterprise-monolith
```

Corpus:

```text
30 frozen rows
10 architecture / implementation
8 dependency / impact / testing
7 unsupported / refusal / prompt-injection
5 Deep-Dive / grounding / architecture evidence
```

Design details:

- [Phase 18 Benchmark Design](phase18-benchmark-design.md)

## Benchmark Freeze

The corpus was finalized before any result files were generated.

Git history preserves the sequence:

| Step | Commit | Meaning |
| --- | --- | --- |
| Corpus freeze | `08c5028` | Fixture expansion, 30 benchmark rows, and design review were committed before result generation. |
| Baseline preservation | `0217a87` | Keyword and retrieval-only baseline outputs were committed before candidate output. |

After benchmark execution, labels, expected facts, evidence groups, and refusal
expectations must not change unless a labeling defect is documented in git
history.

## Baselines

| Baseline | Definition | Output |
| --- | --- | --- |
| Keyword search | Simple lexical/file-content matching over the fixture. | `eval-runner/benchmarks/results/phase18/keyword_baseline.jsonl` |
| Retrieval only | Retrieved files/evidence without grounded generation. | `eval-runner/benchmarks/results/phase18/retrieval_only_baseline.jsonl` |

Knowledge Forge candidate output:

```text
eval-runner/benchmarks/results/phase18/knowledge_forge_candidate.jsonl
```

Generated reports:

```text
eval-runner/benchmarks/results/phase18/phase18-benchmark.json
eval-runner/benchmarks/results/phase18/phase18-benchmark.md
```

## Overall Metrics

| Metric | Keyword | Retrieval Only | Knowledge Forge |
| --- | ---: | ---: | ---: |
| Retrieval recall | 0.481 | 0.819 | 0.948 |
| Evidence recall | 0.483 | 0.824 | 0.948 |
| Answerable accuracy | 0.261 | 0.000 | 1.000 |
| Refusal precision | 0.400 | 1.000 | 1.000 |
| Refusal recall | 0.571 | 0.714 | 1.000 |
| Grounding coverage | 0.167 | 0.000 | 0.717 |
| MRR | 0.800 | 1.000 | 1.000 |
| Average latency ms | 5.533 | 22.600 | 464.067 |
| Average cost USD | 0.000 | 0.000 | 0.003 |

Primary metrics improved substantially over keyword search. Against
retrieval-only, Knowledge Forge improved answerable accuracy, grounding, refusal
recall, evidence recall, and retrieval recall, but at higher latency and cost.

## Per-Category Outcomes

| Category | Outcome | Best Baseline | Correct Delta | Primary Metric Delta |
| --- | --- | --- | ---: | ---: |
| Architecture / implementation | IMPROVED | retrieval_only | 10 | 0.571 |
| Dependency / impact / testing | IMPROVED | retrieval_only | 8 | 0.542 |
| Deep-Dive / grounding / architecture evidence | IMPROVED | retrieval_only | 5 | 0.595 |
| Unsupported / refusal / prompt-injection | UNCHANGED | retrieval_only | 2 | 0.057 |

Material improvement means Knowledge Forge beats both baselines by at least 10%
relative improvement or at least three additional correct rows. The refusal
category improved over keyword search but did not clear the material improvement
threshold against retrieval-only.

## Improved, Unchanged, And Degraded Questions

Compared with keyword search:

```text
26 improved
4 unchanged
0 degraded
```

Compared with retrieval-only:

```text
25 improved
5 unchanged
0 degraded
```

The unchanged rows are unsupported/refusal questions where the stronger baseline
already refused correctly. No rows degraded in this saved proof pack.

## Benchmark Integrity Verification

Runtime/product paths searched:

```text
cmd
internal
ui
deploy
```

Terms searched:

```text
arch-impl
dep-impact
unsupported-
deep-ground
synthetic_enterprise_monolith
keyword_baseline
retrieval_only_baseline
knowledge_forge_candidate
phase18-benchmark
```

Result:

```text
No matches.
```

Conclusion:

```text
PASS
```

Runtime product code does not reference benchmark rows, labels, outputs, or
result files.

## Known Limitations

- The corpus is synthetic and intentionally small.
- Candidate output is a saved proof artifact; live API benchmarking remains
  optional and environment-gated.
- `expected_line_ranges` are optional, so citation accuracy is not the strongest
  signal in this phase.
- Knowledge Forge is slower and more expensive than both simple baselines.
- Refusal quality is strong, but the refusal category did not materially beat
  the stronger retrieval-only baseline under the Phase 18 threshold.

## Phase 19 Decision Table

| Candidate | Evidence | Recommendation | Reason |
| --- | --- | --- | --- |
| Static Code Intelligence | Dependency / impact category improved, but retrieval and evidence recall were lower than architecture questions. | Investigate | Static symbol/reference mapping may help impact questions, but Phase 18 does not yet prove it is mandatory. |
| Graph Retrieval | No category showed a graph-specific failure pattern. Existing retrieval improved without graph traversal. | Reject | Current evidence does not support investing in graph retrieval before stronger benchmark evidence exists. |
| Larger Corpus | Phase 18 proof is synthetic-only and successful enough to justify broader validation. | Proceed | A larger or more realistic corpus is the highest-confidence next investment before adding new retrieval machinery. |

Definitions:

- `Proceed`: evidence clearly supports investment.
- `Investigate`: evidence suggests potential value but is inconclusive.
- `Reject`: evidence does not currently support investment.

## Final Outcome

```text
Phase 18 outcome: Partially Proven
```

Knowledge Forge has measurable evidence of better repository-intelligence
quality than simple baselines on the synthetic enterprise monolith. The next
highest-value step is a larger benchmark corpus, not Phase 19 static intelligence
by default.
