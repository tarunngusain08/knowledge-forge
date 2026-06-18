# Phase 17 Retrospective

## Decision

Phase 17 is closed.

The accepted source of truth is the Phase 17 proof package summarized in
[Phase 17 Validation Proof](../proof/phase17-validation.md). This retrospective
does not reopen Phase 17, add a new gate, or replace the proof document. It only
captures the operating lessons that should shape Phase 18.

## What Found Real Defects

The highest-value activities were the ones that tested real product behavior or
proved a specific causal chain.

| Activity | Value |
| --- | --- |
| Red-team audit | Found false answers, false refusals, weak architecture evidence, and metric-gaming risks that normal happy-path validation missed. |
| Reality audit | Proved the actual product still failed the hardened validator, separating product failures from validator failures. |
| Meta-validation audit | Found places where the validator itself could pass inconsistent or weak evidence. This prevented false confidence. |
| Product conformance reruns | Showed measurable progress from failing gates to `6/6` gates and `0` evaluator issues. |
| R2 root-cause investigation | Prevented blind patching by proving which failures shared a support-gate cause and which failure came from context retention. |

The strongest pattern was:

```text
observed failure
-> code-level root cause
-> bounded remediation
-> fresh reality evidence
```

That pattern should remain mandatory for future quality fixes.

## What Created Little Value

Some work was useful once but should not become routine.

| Activity | Lesson |
| --- | --- |
| Repeated audit loops after root cause was known | Low value. Once a cause is proven, move to bounded remediation instead of producing another report. |
| Intermediate report packages | Useful while debugging, but most are superseded by the final proof and should stay out of the repo. |
| Broad retrospective or timeline reconstruction | Low value for Phase 18. The project already has proof, readiness, roadmap, and repository-health docs. |
| Validation without a stop rule | Risky. It encourages continuous patching instead of decision-driven remediation. |

Phase 18 should avoid creating process artifacts unless they change an
engineering decision.

## Standard Operating Procedure

Future phases should use this sequence:

```text
Build the smallest useful artifact
-> run existing acceptance validation
-> run a reality check against fresh product output
-> investigate only proven failures
-> classify root causes as primary, secondary, or downstream
-> remediate with a bounded cycle count
-> consolidate proof once
```

The default should be one remediation cycle. A second cycle is allowed only when
the first cycle improves the result and leaves a clearly bounded residual
failure. More cycles require explicit human review.

## Anti-Patterns To Avoid

- Auditing without a decision or stop condition.
- Expanding remediation into new product scope.
- Treating generated reports as proof when evaluator output disagrees.
- Teaching the product about validator fixtures or row IDs.
- Adding new retrieval machinery before measuring whether existing retrieval is
  actually insufficient.
- Creating duplicate docs for proof, roadmap, or acceptance status.
- Starting Phase 19 static intelligence before Phase 18 identifies a measured
  weakness.

## Phase 18 Entry Criteria

Phase 18 may start because:

- Phase 17 is accepted and closed.
- Documentation has been consolidated.
- The acceptance validator passes.
- Reality validation passed with `6/6` gates and `0` evaluator issues.
- The roadmap clearly identifies Phase 18 as the validated next step.

Phase 18 should not begin by designing a new audit framework. Its first
meaningful deliverable should be a Benchmark Proof Pack that measures retrieval
and report quality against a small, high-quality benchmark set.

## Phase 18 Operating Rule

Phase 18 should prove measurable value before adding capability.

Recommended rule:

```text
30 excellent benchmark questions > 200 weak benchmark questions
```

The benchmark should show:

- where Knowledge Forge beats naive semantic retrieval
- where it is unchanged
- where it gets worse
- which retrieval layers earned their place
- what weaknesses justify future Phase 19 static intelligence

If a proposed Phase 18 activity does not improve benchmark proof, it should be
deferred.
