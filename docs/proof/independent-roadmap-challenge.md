# Independent Roadmap Challenge Review

## Executive Verdict

Current roadmap decision:

```text
Proceed with Larger Corpus Expansion
```

Board decision:

```text
APPROVE WITH RESERVATIONS
```

Challenge result:

```text
SURVIVED
```

Decision confidence:

```text
Medium-High
```

This review attempted to overturn the Phase 19 Planning Review using only
repository evidence. The strongest challenge is external validity: current
benchmark evidence covers a synthetic Go monolith plus curated Helm and
OpenTelemetry Collector subsets, so it does not prove broad software-repository
generalization. That challenge weakens confidence, but it does not overturn the
decision because Larger Corpus Expansion is the lowest-cost way to test the
specific missing evidence.

Evidence:

- Phase 18 result: `Partially Proven`, with material gains in architecture,
  dependency/impact, and grounding categories, but not a universal win:
  `docs/evaluations/phase18-benchmark-proof.md`.
- Phase 18.5 result: `Generalized` and `Moderately Stable` within
  infrastructure/platform/developer-tooling scope:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- Phase 18.5 corpus coverage and failure clusters are documented in
  `CORPUS-COVERAGE-REPORT.md` and `FAILURE-CLUSTER-ANALYSIS.md`.
- Phase 19 Planning Review selected Larger Corpus Expansion and explicitly
  required human review before implementation:
  `docs/proof/phase19-planning-review.md`.

Conclusion status:

```text
SUPPORTED
```

No material conclusion in this review is used for decision-making without a
repository artifact citation.

## Reviewer Independence Requirement

This review treats all roadmap options as equally plausible at the start. Prior
`Proceed`, `Investigate`, and `Reject` labels are treated as evidence to be
challenged, not conclusions to preserve.

Independence method:

- Reconstruct evidence from repository artifacts before accepting roadmap
  labels.
- Treat a roadmap-overturning result as a valid outcome.
- Penalize unsupported preservation of the existing roadmap.
- Mark unsupported claims as `SPECULATIVE` or `UNSUPPORTED`.

Evidence:

- The current selected decision is documented in
  `docs/proof/phase19-planning-review.md`.
- The benchmark and security evidence used by that decision exists in
  `docs/evaluations/phase18-benchmark-proof.md`,
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`,
  `docs/proof/phase18-6-security-remediation.md`,
  `docs/proof/phase18-7-release-readiness.md`, and
  `docs/proof/phase18-8-security-hardening.md`.

## Evidence Rule And Decision Overturn Threshold

Evidence rule:

- Every criticism, risk, recommendation, and conclusion must cite an existing
  repository artifact.
- If evidence is missing, the statement is marked `SPECULATIVE` and the missing
  artifact is named.
- If a conclusion cannot be supported by repository evidence, it is marked
  `Conclusion Status: UNSUPPORTED` and excluded from decision-making.

Decision overturn threshold:

The current roadmap decision may be overturned only if at least one condition is
met:

1. Existing benchmark evidence directly contradicts Larger Corpus Expansion as
   the highest-information-gain next step.
2. Existing evidence strongly supports Repository Structure Indexing or Static
   Code Intelligence as a better next investment.
3. Existing evidence is too weak or incomplete to justify any implementation
   recommendation.

Result:

```text
Threshold not met.
```

Evidence:

- Phase 18.5 explicitly says Larger Corpus may `Proceed` because results are
  stable or moderately stable across corpora:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.
- Phase 18.5 recommends Repository Structure Indexing and Static Code
  Intelligence only as `Investigate`, not `Proceed`, because failures are narrow
  and limited:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.
- Phase 18.5 documents a scope limitation that results do not generalize to all
  repository types:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.

## Evidence Chain Dependency Audit

| Phase | Evidence Produced | Conclusion Drawn | Assumptions Introduced | Later Validation | Still Unvalidated |
| --- | --- | --- | --- | --- | --- |
| Phase 18 | 30-row synthetic benchmark, keyword baseline, retrieval-only baseline, Knowledge Forge candidate results. Artifact: `docs/evaluations/phase18-benchmark-proof.md`. | Knowledge Forge is `Partially Proven` and improves architecture, dependency/impact, and grounding over baselines. | Synthetic monolith is useful as a controlled first benchmark. | Phase 18.5 adds Helm and OTel corpora. Artifact: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`. | Performance on non-Go, frontend-heavy, data-science, messy application, and poorly organized repos. |
| Phase 18.5 | 70-row multi-corpus benchmark, coverage, stability, failure clusters, leakage review. Artifacts: `BENCHMARK-INTEGRITY-REVIEW.md`, `CORPUS-COVERAGE-REPORT.md`, `CROSS-CORPUS-STABILITY.md`, `FAILURE-CLUSTER-ANALYSIS.md`, `BENCHMARK-LEAKAGE-REVIEW.md`. | Result is `Generalized` and `Moderately Stable` within infrastructure/tooling scope. | Helm and OTel are enough to test infrastructure/tooling generalization. | Stability range max is 0.125 and no major category degraded. Artifact: `CROSS-CORPUS-STABILITY.md`. | Broad software-repository generalization and language diversity. |
| Phase 18.6 | Tenant isolation and deployment trust-boundary remediation proof. Artifact: `docs/proof/phase18-6-security-remediation.md`. | Security risk areas reduced before roadmap investment. | Security state is adequate to proceed with roadmap planning after remediation. | Phase 18.8 closes additional medium findings. Artifact: `docs/proof/phase18-8-security-hardening.md`. | Production telemetry and long-running operational evidence. |
| Phase 18.7 | Release readiness scoring table. Artifact: `docs/proof/phase18-7-release-readiness.md`. | Release readiness status is `READY`. | Docs, onboarding, deployment, benchmark, and security evidence are coherent enough for planning. | Phase 19 Planning Review depends on this readiness state. Artifact: `docs/proof/phase19-planning-review.md`. | Independent operator onboarding in a fresh environment. |
| Phase 18.8 | Five medium findings reproduced and fixed with regression tests. Artifact: `docs/proof/phase18-8-security-hardening.md`. | Remaining concrete medium-severity findings are closed. | No unresolved medium finding should block planning. | Phase 19 Planning Review cites Phase 18.8 as complete. Artifact: `docs/proof/phase19-planning-review.md`. | Future scan variance and production abuse telemetry. |
| Phase 19 Planning Review | Decision doc selecting Larger Corpus Expansion. Artifact: `docs/proof/phase19-planning-review.md`. | Larger Corpus Expansion is the next selected direction. | External validity is the largest remaining uncertainty. | This independent challenge review tests that conclusion. | Whether additional repo families preserve the Phase 18.5 result. |

## Assumption Dependency Map

| Assumption | Origin Phase | Validation Status | Impact If Wrong |
| --- | --- | --- | --- |
| Knowledge Forge materially improves repository-understanding tasks over simple baselines. | Phase 18 | Partial | High |
| Gains generalize beyond the synthetic monolith within infrastructure/tooling repositories. | Phase 18.5 | Validated | High |
| Gains generalize to all repository types. | Phase 18.5 | Unvalidated | High |
| Benchmark labels and outputs are not leaked into runtime product behavior. | Phase 18.5 | Validated | High |
| The current architecture does not yet require graph retrieval. | Phase 18.5 | Partial | Medium |
| Repository Structure Indexing may help if failures expand around architecture/dependency navigation. | Phase 18.5 | Partial | Medium |
| Static Code Intelligence may help if high retrieval recall coexists with low answer/reasoning accuracy. | Phase 18.5 | Partial | Medium |
| The security and release-readiness state is sufficient for planning. | Phase 18.6-18.8 | Validated | Medium |
| Larger Corpus Expansion is cheaper than new retrieval architecture. | Phase 19 Planning Review | Partial | Medium |

Evidence:

- Phase 18 outcome and limitations:
  `docs/evaluations/phase18-benchmark-proof.md`.
- Phase 18.5 result, scope limitation, and Phase 19 decision thresholds:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- Benchmark leakage result:
  `BENCHMARK-LEAKAGE-REVIEW.md`.
- Phase 19 decision:
  `docs/proof/phase19-planning-review.md`.

## Benchmark Confidence Stress Test

### Corpus Selection Risk

Classification:

```text
MEDIUM
```

Evidence supporting the risk:

- Phase 18.5 corpora are synthetic enterprise monolith, Helm, and
  OpenTelemetry Collector:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- Helm and OTel are both Go infrastructure/platform/developer-tooling
  repositories:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- The Phase 18.5 scope limitation explicitly says results do not prove
  performance on messy application repos, frontend-heavy repos, data-science
  repos, or poorly organized repos:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.

Evidence reducing the risk:

- Helm and OTel are different infrastructure repositories with different
  shapes: Helm has command-to-action layering, release lifecycle behavior,
  storage abstraction, repository metadata, and Kubernetes integration; OTel has
  component factories, pipeline assembly, runtime lifecycle, and telemetry
  concerns:
  `CORPUS-COVERAGE-REPORT.md`.

Conclusion:

```text
The result is credible within the documented infrastructure/tooling scope, but
not broad enough for universal repository claims.
```

### Benchmark Size Risk

Classification:

```text
MODERATE
```

Evidence:

- Phase 18.5 includes 70 total rows: 30 synthetic, 20 Helm, and 20 OTel:
  `BENCHMARK-INTEGRITY-REVIEW.md`.
- Category counts are 24 architecture, 20 dependency/impact, 15 refusal, and 11
  deep-dive/grounding rows:
  `BENCHMARK-INTEGRITY-REVIEW.md`.
- Coverage touches 32 of 44 Helm files and 32 of 49 OTel files, but only within
  curated subsets:
  `CORPUS-COVERAGE-REPORT.md`.

Stress-test conclusion:

The benchmark is strong enough to justify the next measurement phase. It is not
large enough to justify declaring the architecture broadly proven across
software repositories.

### Stability Confidence Review

Classification:

```text
MODERATE
```

Evidence:

- Cross-corpus stability is `Moderately Stable`, not `Stable`, with maximum
  primary metric range 0.125:
  `CROSS-CORPUS-STABILITY.md`.
- OTel is the worst corpus for retrieval recall and answerable-question
  accuracy:
  `CROSS-CORPUS-STABILITY.md`.
- Only two failing rows appear, both in OTel:
  `FAILURE-CLUSTER-ANALYSIS.md`.

Challenge:

Because there are only 20 rows per new corpus, a small number of additional OTel
failures could materially affect the narrative. This does not overturn the
decision because the proposed next step is to collect more corpus evidence.

### Benchmark Integrity And Leakage

Classification:

```text
STRONG
```

Evidence:

- Phase 18.5 integrity review reports 70 unique rows, no missing required
  fields, and no required `expected_line_ranges`:
  `BENCHMARK-INTEGRITY-REVIEW.md`.
- Phase 18.5 leakage review reports no runtime product references to benchmark
  row IDs, result paths, output files, or benchmark report names:
  `BENCHMARK-LEAKAGE-REVIEW.md`.
- The evaluator uses expected labels from result rows as the benchmark scoring
  contract, and reports do not feed back into scoring:
  `BENCHMARK-LEAKAGE-REVIEW.md`.

Conclusion:

No repository evidence shows benchmark leakage or label/result feedback loops.

## Strongest Argument That The Roadmap Is Wrong

The strongest evidence-backed argument against Larger Corpus Expansion is:

```text
The project may be mistaking infrastructure/tooling generalization for broader
software-repository generalization.
```

Evidence:

- Phase 18.5 explicitly limits results to infrastructure, platform, and
  developer-tooling repositories:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- Helm and OTel are both Go infrastructure/tooling repositories:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- Coverage only claims coverage of curated subsets, not full upstream repos:
  `CORPUS-COVERAGE-REPORT.md`.
- Cross-corpus stability is `Moderately Stable`, not `Stable`:
  `CROSS-CORPUS-STABILITY.md`.

Does the argument overturn the roadmap?

```text
NO
```

Reason:

The argument shows exactly why Larger Corpus Expansion is the correct next
measurement step. It weakens confidence in universal claims, but it strengthens
the need to test additional repository families before architecture work.

## Generalization Confidence

| Claim | Status | Evidence | Decision Use |
| --- | --- | --- | --- |
| Knowledge Forge improves over baselines on the synthetic monolith. | PROVEN | `docs/evaluations/phase18-benchmark-proof.md` | Supports moving beyond synthetic-only proof. |
| Knowledge Forge improves across Helm and OTel curated subsets. | PROVEN | `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md` | Supports infrastructure/tooling generalization. |
| Knowledge Forge generalizes across infrastructure/tooling repos. | LIKELY | `docs/evaluations/phase18-5-multi-corpus-benchmark.md`, `CROSS-CORPUS-STABILITY.md` | Supports further corpus expansion. |
| Knowledge Forge generalizes to all software repositories. | SPECULATIVE | Missing evidence: non-Go, frontend-heavy, data-science, poorly organized, and application-centric corpora. | Not used to justify implementation. |

## Repository Structure Indexing Challenge

Strongest argument for Repository Structure Indexing first:

- The Phase 18.5 decision table says Repository Structure Indexing meets the
  threshold that architecture/dependency failures are more than 25% of total
  failures:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.
- Failure clusters include missing symbol retrieval, impact analysis, and
  multi-hop dependency reasoning:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Repository Structure Indexing is plausibly cheaper than full Static Code
  Intelligence:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.

Strongest counterargument:

- The failure count is only two rows, both limited to OTel:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Missing architecture evidence is zero:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Helm is 20/20 correct and OTel is 18/20 correct:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- The current benchmark already improved architecture/dependency categories
  without Repository Structure Indexing:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.

Verdict:

```text
Investigate, but do not rank above Larger Corpus Expansion.
```

Conclusion status:

```text
SUPPORTED
```

## Static Code Intelligence Challenge

Strongest argument for Static Code Intelligence first:

- OTel retrieval recall is high while answerable accuracy is lower than Helm:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.
- Remaining OTel failures include missing symbol retrieval and grounding gaps:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Phase 18.5 says evidence suggests symbol/reference help may matter:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.

Strongest counterargument:

- Static Code Intelligence is not proven as the dominant bottleneck because
  there are only two failing rows:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Missing symbol retrieval is tied with citation and grounding gaps at two rows,
  not dominant:
  `FAILURE-CLUSTER-ANALYSIS.md`.
- Architecture and dependency categories improved over baselines in both Helm
  and OTel:
  `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`.
- Graph/static architecture work would be new capability work, while the current
  missing evidence is broader corpus coverage:
  `docs/proof/phase19-planning-review.md`.

Verdict:

```text
Investigate, but do not rank above Larger Corpus Expansion.
```

Conclusion status:

```text
SUPPORTED
```

## Opportunity Cost Analysis

| Option | Implementation Cost | Validation Cost | Expected Information Gain | Roadmap Risk | Evidence |
| --- | --- | --- | --- | --- | --- |
| Larger Corpus Expansion | Medium | Medium | High | Medium | Existing benchmark methodology already supports corpus expansion; external validity is the largest documented gap. Artifacts: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`, `docs/proof/phase19-planning-review.md`. |
| Repository Structure Indexing | Medium | Medium | Medium | Medium | It may address architecture/dependency evidence, but current failures are narrow. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |
| Static Code Intelligence | High | High | Medium | High | Evidence suggests possible symbol/reference help but not dominance. Artifacts: `FAILURE-CLUSTER-ANALYSIS.md`, `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`. |
| Maintain Architecture | Low | Low | Low | Medium | Current architecture is ready and benchmarked, but stopping leaves external validity unresolved. Artifacts: `docs/proof/phase18-7-release-readiness.md`, `docs/proof/phase19-planning-review.md`. |

Relative-cost conclusion:

```text
Larger Corpus Expansion has the highest information gain before new retrieval
architecture because it directly tests the documented missing evidence.
```

## Stop / Maintain Challenge And Kill Criteria

Strongest argument to stop or maintain current architecture:

- Phase 18.7 declares release readiness `READY`:
  `docs/proof/phase18-7-release-readiness.md`.
- Phase 18.8 closes five additional security findings:
  `docs/proof/phase18-8-security-hardening.md`.
- Phase 18.5 reports 68/70 correct with no major category failure:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.

Strongest counterargument:

- Phase 18.5 explicitly says results do not generalize to all repository types:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.
- The Phase 19 Planning Review identifies external validity as the largest
  remaining uncertainty:
  `docs/proof/phase19-planning-review.md`.

Kill criteria for future investment:

- Additional repository families show no material improvement over baselines.
- Cross-corpus stability becomes `Unstable`.
- Failure clusters do not indicate any practical architecture improvement path.
- Latency or cost becomes unacceptable relative to benchmark gains.

Current stop verdict:

```text
No repository evidence currently supports stopping as the #1 roadmap direction.
```

## Missing Evidence Ledger

| Option | Evidence Available | Evidence Missing | Cheapest Measurement Work |
| --- | --- | --- | --- |
| Larger Corpus Expansion | Phase 18.5 shows `Generalized` and `Moderately Stable` across synthetic, Helm, and OTel. Artifacts: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`, `CROSS-CORPUS-STABILITY.md`. | Non-Go, frontend-heavy, application-centric, messy, and data-science repositories. | Benchmark at least three additional repository families and recompute stability. |
| Repository Structure Indexing | Narrow OTel failures include missing symbols, grounding gaps, and one impact/multi-hop row. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Evidence that architecture/dependency failures dominate across more corpora. | Reuse larger corpus benchmark and classify whether structure/navigation failures dominate. |
| Static Code Intelligence | OTel has high retrieval recall but lower answerable accuracy than Helm. Artifact: `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`. | Evidence that symbol/reference failures dominate after retrieval succeeds. | Measure symbol/reference failure share on the next larger corpus before building static analysis. |
| Maintain Architecture | Release readiness is `READY`; security hardening complete. Artifacts: `docs/proof/phase18-7-release-readiness.md`, `docs/proof/phase18-8-security-hardening.md`. | Evidence that further benchmark expansion has low information gain. | Run a bounded corpus expansion; if gains remain stable and failures remain non-actionable, maintenance becomes more credible. |
| Graph Retrieval | Multi-hop dependency reasoning affects one row. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Evidence that graph-specific failures dominate. | Continue failure-cluster measurement; do not implement graph retrieval until graph-specific failures dominate. |

## Evidence Sufficiency Test

| Option | Supporting Evidence | Contradicting Evidence | Missing Evidence | Sufficiency |
| --- | --- | --- | --- | --- |
| Larger Corpus Expansion | Phase 18.5 stability is `Moderately Stable`; Helm and OTel improve over baselines. Artifacts: `CROSS-CORPUS-STABILITY.md`, `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`. | Scope limitation says results do not prove all repository types. Artifact: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`. | Additional repository families. | SUFFICIENT |
| Repository Structure Indexing | Architecture/dependency failure threshold is met in Phase 18.5 decision table. Artifact: `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`. | Failures are only two OTel rows; missing architecture evidence is zero. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Cross-corpus evidence that structure failures dominate. | PARTIAL |
| Static Code Intelligence | Missing symbol retrieval affects two OTel rows. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Failure count is small and tied with citation/grounding gaps. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Evidence that symbol/reference failures dominate across corpora. | PARTIAL |
| Stop / Maintain Current Architecture | Readiness is `READY`; security hardening is complete. Artifacts: `docs/proof/phase18-7-release-readiness.md`, `docs/proof/phase18-8-security-hardening.md`. | External validity remains unresolved. Artifact: `docs/proof/phase19-planning-review.md`. | Evidence that more corpus measurement has low value. | INSUFFICIENT |
| Graph Retrieval | One multi-hop dependency failure exists. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Graph-specific failures do not dominate. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. | Dominant graph-specific failures. | INSUFFICIENT |

## What Would Change The Decision

| Shift | Evidence Required |
| --- | --- |
| Larger Corpus Expansion -> Repository Structure Indexing | Larger corpus results show architecture/dependency navigation failures dominate across multiple corpora, especially with missing architecture evidence or cross-package path misses. |
| Larger Corpus Expansion -> Static Code Intelligence | Larger corpus results show retrieval recall remains high while answerable accuracy and symbol/reference correctness remain low across multiple corpora. |
| Larger Corpus Expansion -> Maintain Architecture | Larger corpus results remain stable, failure clusters stay small and non-actionable, and additional architecture work has low expected information gain. |

Evidence basis:

- Phase 18.5 decision thresholds define when Repository Structure Indexing,
  Static Code Intelligence, Graph Retrieval, and Larger Corpus should proceed:
  `docs/evaluations/phase18-5-multi-corpus-benchmark.md`.

## Decision Survival Test

| Challenge | Survived? | Reason |
| --- | --- | --- |
| External validity challenge | YES | The challenge is real and lowers confidence, but it argues for corpus expansion rather than against it. Artifact: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`. |
| Repository Structure Indexing challenge | YES | Current evidence supports investigation, but failures are narrow and OTel-only. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |
| Static Code Intelligence challenge | YES | Current evidence supports investigation, but symbol/reference failures do not dominate. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |
| Maintain Architecture challenge | YES | Readiness is strong, but external validity remains unresolved. Artifacts: `docs/proof/phase18-7-release-readiness.md`, `docs/proof/phase19-planning-review.md`. |

## Final Ranked Recommendation

| Rank | Option | Recommendation | Reason |
| ---: | --- | --- | --- |
| 1 | Larger Corpus Expansion | Proceed | Best evidence-supported way to resolve the largest documented uncertainty without adding retrieval architecture. Artifacts: `docs/proof/phase19-planning-review.md`, `docs/evaluations/phase18-5-multi-corpus-benchmark.md`. |
| 2 | Repository Structure Indexing | Investigate | Structure/navigation failures may grow on broader corpora, but current evidence is narrow. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |
| 3 | Static Code Intelligence | Investigate | Symbol/reference help may matter, but current failures are too few to justify implementation. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |
| 4 | Stop / Maintain Current Architecture | Reject for now | Release-ready state is valuable, but stopping leaves external validity unresolved. Artifacts: `docs/proof/phase18-7-release-readiness.md`, `docs/proof/phase19-planning-review.md`. |
| 5 | Graph Retrieval | Reject | Graph-specific failures do not dominate. Artifact: `FAILURE-CLUSTER-ANALYSIS.md`. |

## Decision Quality Gate

| Question | Answer |
| --- | --- |
| If all roadmap options were reset today, would repository evidence independently select the same #1 recommendation? | YES. Larger Corpus Expansion has the strongest combination of evidence support, information gain, and lower architecture risk. |
| Strongest single evidence item for Larger Corpus Expansion | Artifact: `eval-runner/benchmarks/results/phase18_5/phase18_5-benchmark.md`. Finding: Helm and OTel both improved over baselines and the result was `Generalized` with `Moderately Stable` stability. Why it matters: it proves the next uncertainty is breadth of generalization, not whether the existing method can ever work. |
| Strongest single evidence item against Larger Corpus Expansion | Artifact: `docs/evaluations/phase18-5-multi-corpus-benchmark.md`. Finding: results do not prove performance for messy application, frontend-heavy, data-science, or poorly organized repos. Why it matters: the current corpus may be too narrow to support broad roadmap confidence. |
| Could the current decision be the result of missing evidence rather than positive evidence? | PARTIALLY. Positive evidence supports infrastructure/tooling generalization, but missing evidence drives the need for corpus expansion. |
| With six months available, what should the first month target? | Larger Corpus Expansion. It validates or falsifies external validity before higher-cost architecture work. |
| If Larger Corpus Expansion succeeds perfectly, what decision is most likely next? | Maintain Architecture or Repository Structure Indexing, depending on failure clusters. Evidence basis: Phase 18.5 says structure indexing should proceed only if architecture/dependency failures dominate. |
| If Larger Corpus Expansion fails, what decision is most likely next? | Repository Structure Indexing if failures cluster around architecture/dependency navigation; Static Code Intelligence if symbol/reference failures dominate. Evidence basis: Phase 18.5 decision thresholds. |

## Independence Check

| Question | Answer |
| --- | --- |
| Did the review discover any evidence that weakens the current roadmap decision? | YES |
| Did the review discover any evidence that strengthens the current roadmap decision? | YES |
| Could the final recommendation have been written before performing this review? | NO |

Why the answer is `NO`:

- The review lowered confidence from `High` to `Medium-High` because corpus
  selection risk is material.
- The review approved the roadmap with reservations instead of giving an
  unconditional approval.
- The review identified external validity as the highest-risk assumption and
  tied the cheapest next evidence directly to that assumption.

## Final Independent Board Verdict

Current roadmap decision:

```text
Proceed with Larger Corpus Expansion
```

Board decision:

```text
APPROVE WITH RESERVATIONS
```

Confidence:

```text
Medium-High
```

Challenge result:

```text
SURVIVED
```

Has the review found evidence strong enough to overturn the current decision?

```text
NO
```

Replacement direction:

```text
None
```

Highest-risk remaining assumption:

```text
Infrastructure/tooling results will remain informative when additional
repository families are introduced.
```

Cheapest next evidence:

```text
Benchmark at least three additional repository families and recompute
cross-corpus stability and failure clusters.
```

If the project stopped today, would existing evidence support publication of
current benchmark conclusions?

```text
PARTIALLY
```

Justification:

The current evidence supports publishing benchmark conclusions within the
documented infrastructure/platform/developer-tooling scope. It does not support
publishing broad claims about all repository types.

Reasons:

- Phase 18 establishes baseline advantage on a synthetic corpus.
- Phase 18.5 extends that advantage to Helm and OTel curated subsets.
- Benchmark integrity and leakage reviews pass.
- Corpus coverage is meaningful inside curated subsets.
- Cross-corpus stability is moderate, not strong.
- Failures are narrow and OTel-only.
- Repository Structure Indexing has partial evidence but not enough for
  implementation.
- Static Code Intelligence has partial evidence but not enough for
  implementation.
- Graph Retrieval lacks dominant graph-specific failure evidence.
- The strongest remaining uncertainty is external validity, which Larger Corpus
  Expansion directly measures.
