# Roadmap

This roadmap separates implemented and validated work from future candidates.
No item should be marked complete unless it is implemented, validated, and
merged.

## Completed

### Phase 17: Repository Intelligence And Deep-Dive Report Conformance

Status:

```text
Complete
Accepted
Merged
```

Validated result:

```text
6/6 acceptance gates pass
0 evaluator issues
```

Scope:

- repository Q&A conformance
- Deep-Dive Report conformance
- architecture evidence validation
- claim grounding validation
- validator independence proof

Proof:

- [Phase 17 Validation Proof](proof/phase17-validation.md)

### Phase 18: Benchmark Proof Pack

Status:

```text
Complete
Partially Proven
```

Goal:

Publish concise benchmark evidence that proves where Knowledge Forge improves
over naive semantic retrieval and where it does not.

Validated outputs:

- benchmark JSON
- human-readable benchmark report
- improved/unchanged/degraded question analysis
- latency and cost comparison
- Phase 19 justification decision table

Proof:

- [Phase 18 Benchmark Design](evaluations/phase18-benchmark-design.md)
- [Phase 18 Benchmark Proof](evaluations/phase18-benchmark-proof.md)

Result:

```text
Partially Proven
```

Knowledge Forge materially outperformed keyword and retrieval-only baselines in
architecture, dependency/impact, and grounding categories. Refusal remained
unchanged against the stronger retrieval-only baseline.

## Validated Next

The next validated investment is a larger benchmark corpus. Phase 19 static code
intelligence should remain an investigation candidate, not an automatic build.

## Future Candidate

### Phase 19: Static Code Intelligence v1

Status:

```text
Candidate
Not started
```

Allowed only if Phase 18 identifies a measured weakness that static
intelligence can address.

Candidate scope:

- interfaces
- structs
- constructors
- imports
- calls
- simple implements mapping

Not allowed by default:

- full call graph
- runtime graph
- dependency-injection inference
- reflection tracking
- multi-repo graph traversal

## Not Yet Started

Anything beyond Phase 19 is not started and should not be described as
implemented or validated.

Examples:

- multi-repository intelligence
- GitHub PR ingestion
- GitHub issue creation
- autonomous code changes
- code generation
- PR review workflow
- architecture diagram generation
- repository memory system
- runtime tracing
- reflection-aware call graph
- full SaaS multi-tenancy

## Roadmap Rule

Every future phase must strengthen the North-Star workflow:

```text
Index repository
-> Ask architecture/code question
-> Inspect cited evidence
-> Generate Deep-Dive Report
-> Generate implementation plan
-> Generate impact analysis
```

Features that do not improve this workflow should remain out of scope.
