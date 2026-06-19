# Cloud Run Deployment

## Services

- `knowledge-forge-migrate`: one-shot migration job from `Dockerfile.migrate`
- `knowledge-forge-api`: Go API container from `Dockerfile.api`
- `knowledge-forge-worker`: Go worker container from `Dockerfile.worker`
- `knowledge-forge-ui`: Streamlit UI container from `Dockerfile.ui`

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

## Deployment Sketch

```bash
gcloud run deploy knowledge-forge-api \
  --source . \
  --region us-central1 \
  --set-env-vars PROVIDER_MODE=cloud \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,JWT_SECRET=JWT_SECRET:latest,INTERNAL_WORKER_TOKEN=INTERNAL_WORKER_TOKEN:latest

gcloud run deploy knowledge-forge-worker \
  --source . \
  --region us-central1 \
  --set-env-vars SERVICE_NAME=knowledge-forge-worker,PROVIDER_MODE=cloud \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,INTERNAL_WORKER_TOKEN=INTERNAL_WORKER_TOKEN:latest

gcloud run deploy knowledge-forge-ui \
  --source . \
  --region us-central1 \
  --set-env-vars API_BASE_URL=https://knowledge-forge-api-url
```

Cloud Tasks or worker callers must include `Authorization: Bearer
$INTERNAL_WORKER_TOKEN` when calling `POST /internal/jobs/{job_id}/process` or
`POST /internal/repository-jobs/{job_id}/process`. Store the token in Secret
Manager, rotate it by publishing a new secret version, and restart the Cloud Run
services to pick up the new value. Keep Cloud Run ingress and IAM restricted to
service-to-service traffic where possible.

Repository ingestion is remote-only by default. `ALLOW_LOCAL_REPOSITORY_PATHS`
should remain `false` in hosted deployments. By default, remote repository URLs
must use HTTPS and target `github.com` or `gitlab.com`; set
`ALLOWED_GIT_REMOTE_HOSTS` to a comma-separated list of approved enterprise Git
hosts when needed.
