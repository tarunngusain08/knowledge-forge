# Phase 18.6 Security Remediation

## Decision

Phase 18.6 addresses tenant isolation and deployment trust-boundary findings before any Phase 19 work.

Security findings are not marked fixed because benchmarks, acceptance validation, or scans pass. Each reproduced issue has an exploit-oriented regression test and a product-side fix.

## Findings Disposition

| Finding | Reproduced | Severity | Fixed | Test |
| --- | --- | --- | --- | --- |
| Dense retrieval leakage | YES | P0 | YES | `TestRetrieveScopesDenseHydrationToRequester` |
| FTS leakage | YES | P0 | YES | `TestPostgresFTSSearchScopesToRequester` |
| Trace IDOR | YES | P0 | YES | `TestGetRepositoryRetrievalTraceRequiresOwner` |
| Symlink escape | NO | P1 | YES | `TestBuildIndexSkipsSymlinkEscape` |
| local_path abuse | YES | P1 | YES | `TestCreateRejectsLocalPathByDefault`, `TestResolveWorktreeRejectsLocalPathWhenDisabled` |
| remote_url SSRF | YES | P1 | YES | `TestRepositoryPolicyRejectsUnsafeRemotes`, `TestResolveWorktreeRejectsUnsafeRemoteBeforeClone` |
| Internal job auth | YES | P1 | YES | `TestRequireInternalWorkerTokenFailsClosedWhenUnset`, `TestRequireInternalWorkerTokenRejectsMissingOrInvalidToken`, `TestRequireInternalWorkerTokenAcceptsBearerOrHeaderToken` |

## Attack Paths And Fix Boundaries

### Dense Retrieval Leakage

Attack path:

```text
User B query
-> Pinecone dense candidates
-> vector ID hydration by document_id/chunk_index
-> chunk from User A document becomes answer evidence
```

Fix boundary:

- Dense vector search now includes `owner_user_id` metadata filtering for newly indexed vectors.
- Dense hydration now requires `owner_user_id` and `documents.status = 'indexed'` in the database query.
- Stale or unauthorized vector IDs are skipped rather than hydrated.

Remaining risk:

- Existing vectors without owner metadata should be reindexed for best candidate recall. Database hydration still prevents unauthorized answer context.

### PostgreSQL FTS Leakage

Attack path:

```text
User B lexical query
-> SearchChunksFTS
-> chunks joined with indexed documents across all owners
-> User A chunk appears in lexical/fused evidence
```

Fix boundary:

- `SearchChunksFTS` now requires `owner_user_id`.
- Retrieval passes the authenticated requester ID into lexical search.

Remaining risk:

- None known for document FTS ownership filtering.

### Retrieval Trace IDOR

Attack path:

```text
User B obtains/guesses User A trace UUID
-> GET /v1/retrieval-traces/{trace_id}
-> trace loaded by UUID only
-> prompt preview, evidence, timings, and cost metadata leak
```

Fix boundary:

- Trace lookup now uses `trace_id + user_id`.
- Handler requires authenticated user context before lookup.
- Cross-user reads return not found.

Remaining risk:

- None known for trace reads. Feedback writes are not evidence disclosure, but can be separately hardened if product requirements need trace ownership validation there too.

### Symlink Escape

Attack path:

```text
Repository contains symlink to outside file
-> indexer walks repository
-> os.ReadFile follows symlink
-> outside file content enters repository index
```

Current proof:

- Current indexer code skips symlink entries before reading.
- Regression test proves escaped symlink content is not indexed.

Remaining risk:

- None known for file symlinks in repository indexing.

### local_path Abuse

Attack path:

```text
Hosted user submits local_path
-> repository registration stores server path
-> git provider resolves local path
-> indexer can read server-local repository/files
```

Fix boundary:

- Repository registration rejects `local_path` by default.
- Git provider also fails closed on legacy `local_path` rows unless trusted operator config enables local paths.

Remaining risk:

- `ALLOW_LOCAL_REPOSITORY_PATHS=true` should be used only for trusted local/operator environments.

### remote_url SSRF

Attack path:

```text
Hosted user submits unsafe remote_url
-> git clone reaches local/private/metadata target
-> server-side network access is attacker-directed
```

Fix boundary:

- Remote URLs must use HTTPS.
- Default hosts are `github.com` and `gitlab.com`.
- Approved enterprise hosts must be explicitly configured.
- `file://`, SSH, localhost, loopback, private, link-local, and metadata-service targets are rejected before clone.

Remaining risk:

- Hostname DNS resolution is not performed in v1; configured enterprise hosts should be reviewed by operators.

### Internal Job Auth

Attack path:

```text
Unauthenticated request
-> POST /internal/jobs/{job_id}/process
-> worker processes arbitrary job ID
```

Fix boundary:

- Internal job routes require `INTERNAL_WORKER_TOKEN`.
- Missing server token fails closed.
- Missing request token returns unauthorized.
- Invalid request token returns forbidden.
- Deployment docs require Secret Manager-backed token configuration and restart-based rotation.

Remaining risk:

- Cloud Run IAM and ingress restrictions should still be applied as defense in depth.

## Benchmark Regression Guardrail

Security changes did not regenerate benchmark candidates. The evaluator was rerun against committed Phase 18 and Phase 18.5 outputs.

| Benchmark | Result |
| --- | --- |
| Phase 18 | Primary metrics and category outcomes unchanged |
| Phase 18.5 | Regenerated JSON byte-identical to committed result |

Guardrail status:

```text
PASS
```

No primary metric regression above the 2% absolute threshold was observed.

## Validation Commands

```bash
GOCACHE=/Users/radhakrishna/Documents/Learning/knowledge-forge-security-remediation/.gocache go test ./internal/retrieval ./internal/httpapi ./internal/repositories ./internal/codeintel ./internal/providers/git ./internal/indexing ./internal/worker
GOCACHE=/private/tmp/knowledge-forge-go-cache go test ./...
GOCACHE=/private/tmp/knowledge-forge-go-cache go vet ./...
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
docker compose config
cd ui/web && npm test && npm run lint && npm run build
python3 eval-runner/acceptance/validation_runner.py --fixtures eval-runner/acceptance/fixtures/acceptance-suite.json --candidate eval-runner/acceptance/candidates/passing-candidate.json --output eval-runner/acceptance/reports
python3 -m unittest discover eval-runner/acceptance -p 'test_*.py'
python3 eval-runner/repo_benchmark_runner.py --input eval-runner/benchmarks/results/phase18/knowledge_forge_candidate.jsonl --baseline keyword=eval-runner/benchmarks/results/phase18/keyword_baseline.jsonl --baseline retrieval_only=eval-runner/benchmarks/results/phase18/retrieval_only_baseline.jsonl --output /tmp/phase18-security-guardrail.json --report-output /tmp/phase18-security-guardrail.md
python3 eval-runner/repo_benchmark_runner.py --input eval-runner/benchmarks/results/phase18_5/knowledge_forge_candidate.jsonl --baseline keyword=eval-runner/benchmarks/results/phase18_5/keyword_baseline.jsonl --baseline retrieval_only=eval-runner/benchmarks/results/phase18_5/retrieval_only_baseline.jsonl --output /tmp/phase18_5-security-guardrail.json --report-output /tmp/phase18_5-security-guardrail.md
```
