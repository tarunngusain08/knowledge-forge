# Deployment

The target production deployment is Google Cloud Run on Google Cloud Platform.
The deployment model keeps the application small enough for a personal project
while still reflecting production SaaS boundaries.

## Services

- Cloud Run API service from `Dockerfile.api`
- Cloud Run worker service from `Dockerfile.worker`
- one-shot migration job from `Dockerfile.migrate`
- React/Vite UI service from `Dockerfile.ui`
- Cloud SQL PostgreSQL for relational data and document `BYTEA` storage
- Secret Manager for database URL, JWT secret, worker token, Pinecone
  credentials, and Google settings
- Cloud Tasks for durable indexing job dispatch
- Cloud Trace/Monitoring for OpenTelemetry exports
- Vertex AI for embeddings, reranking, and Gemini
- Pinecone for vector retrieval

## Required Security Defaults

Hosted deployments should use the hardened defaults validated during Phase 18.6
and Phase 18.8:

- `ALLOW_LOCAL_REPOSITORY_PATHS=false`
- `INTERNAL_WORKER_TOKEN` stored in Secret Manager
- `ALLOWED_GIT_REMOTE_HOSTS` restricted to approved Git hosts
- HTTPS Git remotes only unless an approved enterprise host is configured
- internal job routes protected by worker-token authentication
- Cloud Run ingress and IAM restricted where practical

## Local Vs Cloud

Local development can run with Docker Compose and mock providers. Cloud
deployment should use real managed dependencies and secrets.

Local:

```bash
docker compose up --build
```

Cloud:

- run migrations first
- deploy API, worker, and UI separately
- configure secrets through Secret Manager
- configure Cloud Tasks callers to include the internal worker token

See [Cloud Run Deployment](cloud-run.md) for the concrete command sketch.
