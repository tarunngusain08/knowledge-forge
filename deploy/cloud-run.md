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
  --set-secrets DATABASE_URL=DATABASE_URL:latest,JWT_SECRET=JWT_SECRET:latest

gcloud run deploy knowledge-forge-worker \
  --source . \
  --region us-central1 \
  --set-env-vars SERVICE_NAME=knowledge-forge-worker,PROVIDER_MODE=cloud

gcloud run deploy knowledge-forge-ui \
  --source . \
  --region us-central1 \
  --set-env-vars API_BASE_URL=https://knowledge-forge-api-url
```

In a hardened setup, Cloud Tasks should call
`POST /internal/jobs/{job_id}/process` using an OIDC token. The endpoint should
be restricted to service-to-service traffic.
