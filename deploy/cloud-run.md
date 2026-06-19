# Cloud Run Deployment

This document sketches the intended Google Cloud deployment. It is not a full
Terraform module; it records the service layout, required secrets, and security
configuration that must be preserved when deploying Knowledge Forge.

## Services

- `knowledge-forge-migrate`: one-shot migration job from `Dockerfile.migrate`
- `knowledge-forge-api`: Go API container from `Dockerfile.api`
- `knowledge-forge-worker`: Go worker container from `Dockerfile.worker`
- `knowledge-forge-ui`: React/Vite static UI container from `Dockerfile.ui`

## Managed Dependencies

- Cloud SQL PostgreSQL
- Secret Manager
- Cloud Tasks
- Cloud Trace / Cloud Monitoring
- Pinecone serverless index
- Vertex AI APIs enabled in the target Google Cloud project

## Required Secrets

- `DATABASE_URL`
- `JWT_SECRET`
- `INTERNAL_WORKER_TOKEN`
- `PINECONE_API_KEY`
- `PINECONE_HOST`
- `GOOGLE_CLOUD_PROJECT`
- `GOOGLE_CLOUD_LOCATION`

## Important Environment Variables

- `PROVIDER_MODE=cloud` for real provider integrations.
- `ALLOW_LOCAL_REPOSITORY_PATHS=false` for hosted deployments.
- `ALLOWED_GIT_REMOTE_HOSTS=github.com,gitlab.com` plus approved enterprise Git
  hosts if needed.
- `INTERNAL_WORKER_TOKEN` for internal document and repository job processing
  routes.
- `VITE_API_BASE_URL` when building the React/Vite UI image.

## Deployment Sketch

```bash
gcloud run deploy knowledge-forge-api \
  --source . \
  --region us-central1 \
  --set-env-vars PROVIDER_MODE=cloud,ALLOW_LOCAL_REPOSITORY_PATHS=false,ALLOWED_GIT_REMOTE_HOSTS=github.com,gitlab.com \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,JWT_SECRET=JWT_SECRET:latest,INTERNAL_WORKER_TOKEN=INTERNAL_WORKER_TOKEN:latest

gcloud run deploy knowledge-forge-worker \
  --source . \
  --region us-central1 \
  --set-env-vars SERVICE_NAME=knowledge-forge-worker,PROVIDER_MODE=cloud \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,INTERNAL_WORKER_TOKEN=INTERNAL_WORKER_TOKEN:latest

gcloud run deploy knowledge-forge-ui \
  --source . \
  --region us-central1 \
  --build-arg VITE_API_BASE_URL=https://knowledge-forge-api-url
```

The exact `gcloud` flags may change with the deployment pipeline. The important
contract is that API, worker, migration, and UI are separate deployable units
and secrets come from Secret Manager.

## Internal Worker Authorization

Cloud Tasks or worker callers must include:

```text
Authorization: Bearer $INTERNAL_WORKER_TOKEN
```

when calling:

- `POST /internal/jobs/{job_id}/process`
- `POST /internal/repository-jobs/{job_id}/process`

Store the token in Secret Manager. Restart-based rotation is acceptable:
publish a new secret version and restart Cloud Run services to pick it up.
Missing server token should fail closed for internal job processing.

## Repository Ingestion Boundary

Repository ingestion is remote-only by default in hosted deployments.

Rules:

- Keep `ALLOW_LOCAL_REPOSITORY_PATHS=false`.
- Allow HTTPS remotes from approved hosts only.
- Default approved public hosts are `github.com` and `gitlab.com`.
- Use `ALLOWED_GIT_REMOTE_HOSTS` for approved enterprise Git hosts.
- Block local files, SSH remotes, localhost, private-network targets, and
  metadata-service style addresses.

## Operational Notes

- Keep API and worker services on the same database and provider configuration.
- Restrict Cloud Run ingress and IAM to the narrowest practical callers.
- Configure request size limits at the edge as defense in depth for upload
  handling.
- Export OpenTelemetry traces to Cloud Trace/Monitoring for API and worker
  diagnostics.
- Keep benchmark, acceptance, and security proof documents in the repository so
  deployment decisions can be audited without raw Desktop evidence packages.
