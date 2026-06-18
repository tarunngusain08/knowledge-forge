# Phase 18.5 Failure Cluster Analysis

Status: PASS

| Cluster | Rows affected | Corpora affected | Row IDs | Possible cause |
| --- | ---: | --- | --- | --- |
| citation gaps | 2 | otel-collector | otel-deep-003, otel-dep-006 | Expected citation or file evidence was missing. |
| grounding gaps | 2 | otel-collector | otel-deep-003, otel-dep-006 | Claim grounding coverage was incomplete. |
| missing symbol retrieval | 2 | otel-collector | otel-deep-003, otel-dep-006 | Expected symbols were not retrieved. |
| impact analysis | 1 | otel-collector | otel-dep-006 | Impact questions did not retrieve all affected files or facts. |
| multi-hop dependency reasoning | 1 | otel-collector | otel-dep-006 | Dependency/impact questions missed required multi-file evidence. |
| missing architecture evidence | 0 | None | None | Architecture evidence groups or files were missing. |
| refusal classification | 0 | None | None | Unsupported questions were answered or supported questions were refused. |

Top failure cluster: citation gaps, grounding gaps, and missing symbol retrieval are tied at 2 rows affected, both limited to OpenTelemetry Collector.

Phase 19 implication: failures are narrow and do not meet the threshold for graph retrieval. Repository Structure Indexing remains worth investigating only as a cheaper way to improve architecture/dependency evidence on larger corpora.
