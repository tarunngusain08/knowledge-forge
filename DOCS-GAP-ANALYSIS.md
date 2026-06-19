# Documentation Gap Analysis

This analysis records the documentation review performed during the
Documentation Excellence and Evidence Consolidation refresh. It is committed
because it explains what was improved, what evidence was used, and what remains
intentionally unresolved.

## Review Scope

Reviewed areas:

- `README.md`
- `docs/README.md`
- `docs/architecture.md`
- `docs/roadmap.md`
- `docs/readiness-scorecard.md`
- `docs/evaluations/`
- `docs/proof/`
- `deploy/`
- committed Phase 18 and Phase 18.5 benchmark results
- Phase 18.6 and Phase 18.8 security proof documents

Out of scope:

- product behavior changes
- benchmark reruns
- benchmark label changes
- validator changes
- security control changes
- Phase 19 implementation

## Findings And Fixes

| Area | Before | Fix Applied |
| --- | --- | --- |
| README first impression | The README was accurate and modern, but the evidence trail was mostly text-heavy. | Added a maturity snapshot, clearer value proposition, target users, non-goals, and visual proof summaries. |
| Architecture overview | Architecture existed in prose and ASCII flow, but there was no repository-friendly visual summary. | Added `docs/images/architecture/architecture-overview.png` based on existing architecture and security docs. |
| Benchmark evidence | Phase 18 and 18.5 results were present in proof docs and JSON, but easy to miss on first read. | Added `benchmark-summary.png` and `benchmark-comparison.png` generated from committed benchmark JSON. |
| Security posture | Phase 18.6 and 18.8 proof docs were strong but buried for a new reader. | Added `security-posture-summary.png` and linked it from the README security section. |
| Evidence navigation | The README listed proof docs, but the milestone sequence required reading several sections. | Added an Evidence Trail table mapping each milestone to proof and outcome. |
| Roadmap clarity | Roadmap docs were already strong; README needed a faster summary of selected/candidate/rejected work. | Kept Larger Corpus Expansion as selected next direction and repeated not-started boundaries. |
| Screenshot inventory | UI screenshots were requested, but this checkout does not have runnable UI dependencies or a seeded backend session. | Added `docs/images/ui/README.md` to document the no-fabrication rule and future capture criteria. |

## Stale Or Risky Documentation Notes

No product claims were added without repository evidence. The main remaining
documentation risks are:

- Some historical docs are intentionally retained and may describe older phase
  expectations. The current sources of truth are `README.md`, `docs/README.md`,
  `docs/readiness-scorecard.md`, `docs/roadmap.md`, `docs/proof/`, and
  `docs/evaluations/`.
- Phase 18.5 evidence is explicitly limited to infrastructure, platform, and
  developer-tooling repositories. Documentation should not generalize it to all
  repository types.
- Repository Structure Indexing, Static Code Intelligence, Graph Retrieval, PR
  review, and autonomous code changes are not implemented. Docs should continue
  to describe them only as candidate, rejected, or not-started work.

## Screenshot Gaps

Real UI screenshots were not committed in this refresh.

Reason:

- `ui/web/node_modules` is absent in this checkout.
- The Streamlit fallback dependency is not installed in the default local
  Python environment.
- No seeded live backend session was created during this docs-only branch.

Decision:

```text
Do not fabricate product UI screenshots.
```

Future UI screenshots should be captured only from a running product with real
local or seeded-demo data. Candidate views:

- repository registration/indexing flow
- repository Q&A flow
- grounded answer with citations
- retrieval trace or debug view
- Deep-Dive Report output
- evaluation or benchmark report view if the UI exposes it

## Documentation Quality Review

Reader goals after this refresh:

| Reader | Should Be Able To Answer |
| --- | --- |
| New engineer | What Knowledge Forge does, how it works, how to run it, and where to start contributing. |
| Senior engineer | What architecture exists, what evidence supports quality, and what tradeoffs remain. |
| Security reviewer | Which tenant-isolation and trust-boundary controls exist and where proof lives. |
| Interviewer / portfolio reviewer | Why the project is more than a chatbot demo and what evidence makes it credible. |
| Future maintainer | What roadmap direction is selected next and what must remain out of scope. |

## Remaining Recommended Documentation Work

These are not blockers for this branch:

1. Capture real product UI screenshots after a seeded local demo environment is
   available.
2. Add a short demo script that creates a repository, asks a question, opens
   citations, and exports a Deep-Dive Report.
3. Add a lightweight "first contribution" guide after the next implementation
   phase begins.
4. Refresh historical docs only when they actively confuse new readers; avoid
   rewriting history just for polish.
