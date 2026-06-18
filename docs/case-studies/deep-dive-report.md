# Case Study: Repository Deep-Dive Report

This case study shows the intended Knowledge Forge repository-intelligence
workflow without requiring a reader to inspect raw audit packages.

## Scenario

A new engineer joins a codebase and wants to understand how the system is put
together before making changes.

Instead of reading every file manually, the engineer imports the repository,
generates a Deep-Dive Report, and inspects the evidence behind each claim.

## Workflow

```text
Import repository
-> Index repository snapshot
-> Generate Deep-Dive Report
-> Review architecture overview
-> Open citations
-> Inspect missing context
-> Ask follow-up repository questions
-> Generate plan or impact analysis
```

## What The Report Produces

A Deep-Dive Report returns structured JSON and Markdown export with sections
such as:

- architecture overview
- entry points
- main packages
- authentication flow
- data layer
- external services
- testing strategy
- risk areas
- suggested improvements
- evidence quality
- missing context

Every supported claim should trace back to repository evidence. Unsupported or
weakly supported areas should be surfaced as missing context instead of being
filled in with guesses.

## Why The Shared Evidence Pass Matters

The report does not run a completely separate RAG pipeline for every section.
It starts with one broad evidence pass, then performs targeted follow-up
retrieval only when important sections are weak.

This reduces:

- latency
- cost
- duplicate retrieval
- debugging complexity

It also makes the report easier to explain: the report is built from a shared
evidence set, not a collection of unrelated prompts.

## Evidence Quality

The Evidence Quality section tells the reviewer how much evidence was available
and where the system was uncertain.

Good report behavior:

- cites source files and line ranges
- separates observed evidence from assumptions
- identifies missing context
- refuses unsupported conclusions

Bad report behavior:

- summarizes README prose as architecture
- claims unsupported systems exist
- reports high confidence without citations
- omits missing context

## Demo Story

An interview demo can follow this path:

```text
Repository: Knowledge Forge
Question: How are deep-dive reports generated?
Evidence: internal/codeqa/service.go and internal/codeqa/reports.go
Output: cited explanation plus report Markdown
Validation: accepted Phase 17 gates prove this class of question no longer false-refuses
```

The strongest part of the demo is not that the model can write a summary. It is
that the summary is tied to repository evidence and the system exposes what it
could not prove.

## Related Docs

- [Architecture](../architecture.md)
- [Phase 17 Validation Proof](../proof/phase17-validation.md)
- [Acceptance Methodology](../evaluations/acceptance-methodology.md)
- [Roadmap](../roadmap.md)
