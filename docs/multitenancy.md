# Future Multi-Tenant Design

v1 intentionally supports a single workspace.

Future multi-tenant changes:

- Add `tenants` table and `tenant_id` to users, documents, chunks, chat sessions, eval runs, traces, and cost events
- Enforce tenant filters in every SQL query
- Use Pinecone namespace per tenant
- Add per-tenant quotas for uploaded bytes, documents, questions, and token spend
- Roll up costs per tenant, user, document, and evaluation run
- Include tenant ID in OpenTelemetry attributes, but never document text
- Consider PostgreSQL row-level security once the model stabilizes

