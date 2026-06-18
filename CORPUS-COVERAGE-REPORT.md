# Phase 18.5 Corpus Coverage Report

Status: PASS

Coverage measures how much of each curated corpus is touched by benchmark labels. It does not claim coverage of the full upstream repository.

| Corpus | Files indexed | Files benchmarked | File coverage | Symbols benchmarked | Evidence groups benchmarked |
| --- | ---: | ---: | ---: | ---: | ---: |
| synthetic_enterprise_monolith | 10 | 10 | 100.0% | 31 | 10 |
| helm | 44 | 32 | 72.7% | 38 | 25 |
| otel-collector | 49 | 32 | 65.3% | 25 | 28 |

## Corpus Notes

### synthetic_enterprise_monolith

- Source: local synthetic fixture
- Snapshot: not_applicable
- Rationale: Retained as the controlled baseline corpus from Phase 18.
- Exclusions: No additional areas excluded in Phase 18.5.

### helm

- Source: https://github.com/helm/helm
- Snapshot: fa9efb07d9d8debbb4306d72af76a383895aa8c4
- Rationale: Helm is a mature Go developer tool with command-to-action layering, release lifecycle behavior, storage abstraction, repository metadata, and Kubernetes integration.
- Exclusions: vendor, generated code, large testdata, plugin implementation details, search UI helpers, unrelated documentation, and broad package tests.

### otel-collector

- Source: https://github.com/open-telemetry/opentelemetry-collector
- Snapshot: a082f2e439e8f77a9a9503d54d8afea576f2d08c
- Rationale: OpenTelemetry Collector is a modular infrastructure repository with component factories, pipeline assembly, runtime lifecycle, and telemetry concerns.
- Exclusions: vendor, generated code-heavy pdata packages, large testdata, broad test suites, connector variants, most helper packages, and unrelated documentation.
