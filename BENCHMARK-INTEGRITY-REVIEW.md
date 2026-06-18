# Phase 18.5 Benchmark Integrity Review

Status: PASS

Benchmark corpus:

```text
eval-runner/benchmarks/phase18_5_multi_corpus.jsonl
```

## Row Counts

| Check | Result |
| --- | ---: |
| Total rows | 70 |
| Unique IDs | 70 |
| Missing required fields | 0 |
| Required `expected_line_ranges` | No |

## Corpus Counts

| Corpus | Rows |
| --- | ---: |
| synthetic_enterprise_monolith | 30 |
| helm | 20 |
| otel-collector | 20 |

## Category Counts

| Category | Rows |
| --- | ---: |
| architecture_implementation | 24 |
| dependency_impact_testing | 20 |
| unsupported_refusal_prompt_injection | 15 |
| deep_dive_grounding_architecture_evidence | 11 |

## New Corpus Distribution

| Corpus | Architecture / Implementation | Dependency / Impact / Testing | Unsupported / Refusal / Prompt Injection | Deep-Dive / Grounding / Architecture Evidence |
| --- | ---: | ---: | ---: | ---: |
| helm | 7 | 6 | 4 | 3 |
| otel-collector | 7 | 6 | 4 | 3 |

## Label Completeness

Every row includes:

- `id`
- `corpus`
- `category`
- `question`
- `expected_files`
- `expected_symbols`
- `expected_answer_facts`
- `required_evidence_groups`
- `should_refuse`

Unsupported/refusal rows were reviewed and intentionally contain empty evidence
expectations with `should_refuse=true`.

## Freeze Statement

The benchmark corpus is frozen before baseline generation and Knowledge Forge
candidate generation.

After this point, labels, expected facts, evidence groups, and refusal
expectations must not change unless a documented labeling defect is discovered.
