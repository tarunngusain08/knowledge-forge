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

## Validated Next

### Phase 18: Benchmark Proof Pack

Status:

```text
Next
Not started in this branch
```

Goal:

Publish concise benchmark evidence that proves where Knowledge Forge improves
over naive semantic retrieval and where it does not.

Expected outputs:

- benchmark JSON
- human-readable benchmark report
- a small set of high-quality case studies
- improved/unchanged/degraded question analysis
- latency and cost comparison
- decision record for retrieval components that should remain enabled

Boundary:

Phase 18 should measure the existing validated system. It should not add static
intelligence, graph retrieval, agents, code generation, or PR review.

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
