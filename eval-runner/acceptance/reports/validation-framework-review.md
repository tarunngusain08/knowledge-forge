# Validation Framework Review

## Implemented Gates

| Gate | Status |
| --- | --- |
| Gate 1 Refusal Matrix | pass |
| Gate 2 Answer Relevance | pass |
| Gate 3 Architecture Evidence | pass |
| Gate 4 Metric Integrity | pass |
| Gate 5 Benchmark Label Completeness | pass |
| Gate 6 Adversarial Benchmark | pass |

## Fixture Counts

- Refusal matrix rows: 11
- False refusal catchers: 5
- False answer catchers: 6
- Answer relevance rows: 4
- Architecture fixtures: 3
- Metric audit rows: 5

## Benchmark Counts

- Candidate refusal rows: 11
- Candidate answer relevance rows: 4
- Candidate architecture rows: 3
- Candidate metric rows: 5

## Passing Examples

- RF-001: answerable RAG retrieval question requires retrieval/RAG evidence.
- FA-001: unsupported revenue API question must refuse despite API path matches.
- ARCH-NEG-001: README-only architecture evidence must not pass.
- MET-004: claim grounding cannot pass when claim-to-citation labels are unavailable.

## Rejected Negative-Control Examples

- RF-001/RF-002: answerable RAG and HTTP API questions refused by exact-identifier logic.
- FA-001/FA-002/FA-004: revenue, payroll, and external-fact questions answered from weak or missing evidence.
- AR-001/AR-002: authentication answers graded relevant despite unrelated or partial evidence.
- ARCH-NEG-001/ARCH-NEG-002: README-only and directory-only fixtures reported High-confidence layers.
- MET-003/MET-004: section-support coverage used as grounding and unavailable claim grounding treated as pass.
- Gate 5: incomplete labels allowed into acceptance results.

## Failing Examples

No failures.

## Coverage

- Refusal decision matrix covers answerable acronym questions and unsupported business/external/prompt-injection questions.
- Answer relevance covers expected files, symbols, evidence groups, and answer facts.
- Architecture validation covers source-code positive, docs-only negative, and directory-only negative fixtures.
- Metric validation covers metric purpose, limitations, anti-gaming checks, grounding availability, section-support misuse, and label completeness.

## Result

pass
