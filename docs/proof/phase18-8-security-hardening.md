# Phase 18.8 Security Hardening

## Decision

Phase 18.8 validates and remediates five medium-severity findings from the
Deep Security Scan at:

```text
/tmp/codex-security-scans/knowledge-forge-deep-security-scan/1b7b9cde66bb_20260619T065053/report.md
```

This branch does not start Phase 19 and does not address lower-priority
evaluation, prompt-hardening, cost-governance, or error-redaction findings.

## Finding Disposition

| Finding | Disposition | Fixed | Regression Test | Notes |
| --- | --- | --- | --- | --- |
| Ingestion job IDOR | REPRODUCED | YES | `TestGetRepositoryIngestionRequiresOwner` | Ingestion job reads now require requester ownership. |
| Feedback trace ownership bypass | REPRODUCED | YES | `TestCreateRepositoryFeedbackRequiresTraceOwner` | Feedback creation now verifies trace ownership before insert. |
| Refusal-path evidence exposure | REPRODUCED | YES | `TestAskRefusesUnsupportedQuestionBeforeGeneration` | Refused answers no longer return hydrated retrieval hits or prompt previews. |
| Deleted-document reindex race | REPRODUCED | YES | `TestProcessJobDoesNotPersistChunksWhenDocumentDeletedDuringIndexing` | Deleted documents cannot be moved back to indexed by worker status transitions. |
| Multipart upload/body DoS | REPRODUCED | YES | `TestUploadDocumentRejectsOversizedMultipartBeforeFullParse` | Multipart parsing is now protected by an outer request body cap. |

## Reproduction And Fix Summary

### Ingestion Job IDOR

Attack path:

```text
User B
-> GET /v1/ingestions/{user_a_job_id}
-> handler loads job by UUID only
-> User B receives User A ingestion metadata
```

Fix boundary:

```text
HTTP handler -> repository service -> repository store query
```

The handler now requires the authenticated user and calls an owner-scoped
ingestion lookup. The store loads ingestion jobs by job ID plus repository
owner, returning no row for foreign jobs.

Remaining risk:

```text
None known for this read path. Worker/internal job processing remains separate.
```

### Feedback Trace Ownership Bypass

Attack path:

```text
User B
-> POST /v1/feedback { trace_id: user_a_trace_id }
-> feedback inserted without trace ownership check
-> User B can corrupt User A trace feedback
```

Fix boundary:

```text
HTTP handler before feedback insert
```

The handler now verifies the trace with the existing owner-scoped trace lookup
before inserting feedback.

Remaining risk:

```text
None known for the feedback API path.
```

### Refusal-Path Evidence Exposure

Attack path:

```text
Unsupported repository question
-> retrieval returns hydrated chunks
-> support gate refuses answer
-> response/trace still exposes retrieval hits and prompt preview
```

Fix boundary:

```text
Repository QA response and trace assembly
```

When the support gate refuses, the service now redacts hydrated retrieval hits,
retrieved chunk IDs, context token count, and prompt preview before returning the
response or persisting the trace. Support-gate diagnostics remain in retrieval
configuration.

Remaining risk:

```text
Refused traces still keep non-sensitive diagnostics such as query category,
retrieval path, support reason, and latency.
```

### Deleted-Document Reindex Race

Attack path:

```text
Worker loads document bytes
-> user deletes document
-> worker continues indexing
-> worker marks document indexed again and persists retrieval material
```

Fix boundary:

```text
Document status transition query and worker pre-persistence check
```

Document status updates now refuse to update rows already marked deleted. The
worker also rechecks that the document remains indexable before persisting
chunks and vectors, and attempts cleanup if final indexed status cannot be
committed.

Remaining risk:

```text
Full transactional cancellation of in-flight workers remains future hardening,
but deleted rows are no longer overwritten back to indexed by the status update.
```

### Multipart Upload/Body DoS

Attack path:

```text
Large multipart upload
-> ParseMultipartForm runs before hard request body cap
-> server can parse/spool body before application rejection
```

Fix boundary:

```text
HTTP upload handler and API server timeout configuration
```

The upload handler now wraps the request body with `http.MaxBytesReader` before
multipart parsing. The API server also sets read, write, and idle timeouts in
addition to the existing header timeout.

Remaining risk:

```text
Reverse proxy limits should still be configured in deployment for defense in
depth.
```

## Validation Commands

Focused regression suite:

```bash
env GOCACHE=/Users/radhakrishna/Documents/Learning/knowledge-forge-security-hardening-18-8/.cache/go-build \
  go test ./internal/httpapi ./internal/codeqa ./internal/worker ./internal/repositories
```

Result:

```text
PASS
```

Full validation:

```bash
git diff --check
env GOCACHE=/tmp/knowledge-forge-m18-8-go-cache go test ./...
env GOCACHE=/tmp/knowledge-forge-m18-8-go-cache go vet ./...
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
cd ui/web && npm test
cd ui/web && npm run lint
cd ui/web && npm run build
docker compose config
make validate-acceptance
```

Result:

```text
PASS
```
