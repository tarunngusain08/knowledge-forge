# Deployment

The target production deployment is Google Cloud Run with:

- Cloud SQL PostgreSQL for relational data and document `BYTEA` storage
- Secret Manager for database URL, JWT secret, Pinecone credentials, and Google settings
- Cloud Tasks for durable indexing job dispatch
- Cloud Trace/Monitoring for OpenTelemetry exports

Detailed deployment manifests and commands are added in the deployment milestone.

