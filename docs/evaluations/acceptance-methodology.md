# Acceptance Methodology

Knowledge Forge uses acceptance validation to answer a simple question:

```text
Can the actual product produce evidence-grounded repository intelligence
without repeating known failure modes?
```

The acceptance suite is not a replacement for unit tests, integration tests, or
future benchmark reports. It is a guardrail that prevents known incorrect
behavior from silently passing.

## Why Acceptance Validation Exists

RAG systems can look successful while still being wrong:

- they can answer unsupported questions
- they can refuse answerable repository questions
- they can cite irrelevant files
- they can infer architecture from documentation alone
- they can report grounding metrics without claim-level evidence
- they can pass benchmarks because labels are incomplete

Knowledge Forge treats these as product failures, not just metric issues.

## The Six Gates

| Gate | Name | What It Catches |
| --- | --- | --- |
| 1 | Refusal Matrix | false refusals and false answers |
| 2 | Answer Relevance | irrelevant answers, missing evidence groups, missing expected facts |
| 3 | Architecture Evidence | README-only or directory-only architecture claims |
| 4 | Metric Integrity | grounding metrics without claim-to-evidence mappings |
| 5 | Benchmark Label Completeness | fixtures that cannot actually validate behavior |
| 6 | Adversarial Benchmark | known red-team failure patterns |

Each gate uses executable fixtures and evaluator output. Generated Markdown
reports are summaries; they do not override evaluator verdicts.

## Evidence Contract

The product must validate the difference between:

```text
evidence exists
```

and:

```text
evidence supports the question
```

For example, a payroll UI question cannot be supported by any UI file. It
requires payroll-domain evidence. A revenue API question cannot be supported by
any API handler. It requires revenue-domain endpoint evidence.

## Grounding Contract

Grounding coverage is only meaningful when each claim maps to evidence:

```text
claim
-> citation
-> file
-> line range
-> evidence excerpt
```

`section_support_coverage` is diagnostic only. It must not be treated as
claim-level grounding.

## Validator Independence

The product must not inspect:

- acceptance fixtures
- benchmark labels
- candidate JSON
- validation outputs
- previous audit artifacts

Product behavior must come from repository evidence, retrieval results, support
metadata, and product configuration only.

## Running Acceptance Validation

```bash
make validate-acceptance
```

The Makefile invokes the acceptance runner with fixture and candidate paths under
`eval-runner/acceptance`.

Normal project validation should include:

```bash
make test
make vet
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
cd ui/web && npm test && npm run lint && npm run build
docker compose config
make validate-acceptance
```

## Phase 17 Result

The accepted Phase 17 conformance result is:

```text
6/6 gates pass
0 evaluator issues
Acceptance validator: pass
```

See the canonical proof summary:

- [Phase 17 Validation Proof](../proof/phase17-validation.md)

## Limits

The acceptance framework proves conformance to the known Phase 17 failure set.
It does not prove the system is better than baseline retrieval across a broad
benchmark corpus. That belongs to Phase 18: Benchmark Proof Pack.
