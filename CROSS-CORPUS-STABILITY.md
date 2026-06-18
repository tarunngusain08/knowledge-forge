# Phase 18.5 Cross-Corpus Stability

Classification: Moderately Stable

Maximum primary metric range: 0.125

| Metric | Best corpus | Worst corpus | Range |
| --- | --- | --- | ---: |
| retrieval_recall | helm | otel-collector | 0.063 |
| evidence_recall | helm | synthetic_enterprise_monolith | 0.052 |
| answerable_question_accuracy | helm | otel-collector | 0.125 |
| refusal_precision | helm | helm | 0.000 |
| refusal_recall | helm | helm | 0.000 |
| grounding_coverage | helm | otel-collector | 0.083 |

Interpretation: the benchmark is moderately stable because primary metric ranges stay below 0.20 and no major category is degraded.
