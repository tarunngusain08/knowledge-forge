# Repository Health Review

Branch B reviewed the repository structure for discoverability and documentation
overlap. No files were deleted or archived in this branch.

## Top-Level Discoverability

| Path | Recommendation | Reason |
| --- | --- | --- |
| `README.md` | keep/update | Primary product entry point for new engineers and interviewers. |
| `docs/README.md` | keep/update | Documentation index and routing page. |
| `docs/proof/phase17-validation.md` | keep | Canonical current Phase 17 proof narrative. |
| `docs/roadmap.md` | keep | Canonical roadmap status and future boundary. |
| `docs/readiness-scorecard.md` | keep | Concise current-state checkpoint. |
| `R2-ROOT-CAUSE-CLOSURE.md` | keep | Root-cause source artifact linked from proof doc. |
| `VALIDATOR-INDEPENDENCE.md` | keep | Validator-independence source artifact linked from proof doc. |

## Documentation Keep List

| Path | Role |
| --- | --- |
| `docs/01-fr-nfr-scale-estimation.md` | Requirements and scale assumptions. |
| `docs/02-hld-component-design.md` | High-level design and component responsibilities. |
| `docs/03-usecases-sequence-diagrams.md` | Use cases and sequence diagrams. |
| `docs/04-lld.md` | Low-level implementation design. |
| `docs/05-db-design-schema.md` | Database design and schema reference. |
| `docs/06-ui-backend-quality.md` | UI/backend quality guidance. |
| `docs/architecture.md` | Current architecture narrative. |
| `docs/evaluation.md` | Metric definitions and benchmark runner notes. |
| `docs/evaluations/acceptance-methodology.md` | Acceptance gates and validation method. |
| `docs/case-studies/deep-dive-report.md` | Publishable product case study. |
| `docs/portfolio/README.md` | Portfolio and interview overview. |
| `docs/adr/*` | Architecture decision records. |

## Merge Candidates

| Path | Recommendation |
| --- | --- |
| `docs/phase17-hardening-review.md` | Keep as historical context, but treat `docs/proof/phase17-validation.md` as the current proof source. Future edits should merge any still-useful Phase 17 summary into the proof doc rather than extending this file. |
| `docs/evaluation.md` and `docs/evaluations/acceptance-methodology.md` | Keep both because they have distinct roles: metrics reference vs acceptance gate methodology. Avoid duplicating gate details in `docs/evaluation.md`. |
| `docs/architecture.md`, `docs/02-hld-component-design.md`, and `docs/04-lld.md` | Keep all three, but use them as different depths of the same architecture story: current overview, HLD, and LLD. |

## Archive Candidates

| Path | Recommendation |
| --- | --- |
| Raw Desktop audit folders | Do not commit. Keep local or archive outside the repo. Publish only curated summaries. |
| Generated validator report dumps | Do not commit unless they are intentionally part of the acceptance framework. Link to local evidence package paths from proof docs instead. |
| Superseded branch-local review packages | Keep only if they provide source evidence not already summarized by the Phase 17 proof doc. |

## Update Candidates

| Path | Recommendation |
| --- | --- |
| `docs/implementation-plan.md` | Future update should mark Phase 17 accepted and point to `docs/roadmap.md` for current next steps. |
| `docs/evaluation.md` | Future update should link to Phase 18 benchmark reports after they exist. |
| `deploy/*` | Future update should include any changes needed when real cloud deployment is exercised. |

## Duplicate Content Risk

Deep-Dive Reports, repository intelligence, validation, and roadmap status now
have canonical destinations:

| Topic | Canonical Destination |
| --- | --- |
| Product overview | `README.md` |
| Current validation proof | `docs/proof/phase17-validation.md` |
| Acceptance method | `docs/evaluations/acceptance-methodology.md` |
| Roadmap status | `docs/roadmap.md` |
| Interview/demo story | `docs/portfolio/README.md` |
| Detailed architecture | `docs/architecture.md`, `docs/02-hld-component-design.md`, `docs/04-lld.md` |

Future docs should link to these destinations instead of recreating the same
content.
