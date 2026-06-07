# Knowledge Forge

Production-style company-document RAG assistant built with Go, PostgreSQL,
Pinecone, Vertex AI, and LangChainGo behind internal provider interfaces.

## Target Retrieval Flow

```text
Question
-> Question Rewriter
-> Vertex Query Embedding
-> Pinecone Dense Retrieval
+
PostgreSQL FTS Retrieval
-> Reciprocal Rank Fusion
-> Vertex Ranking API
-> Gemini
-> Grounded Response + Citations
```

## Local Development

```bash
cp .env.example .env
make tidy
make migrate-up
make test
docker compose up --build
```

The API exposes `GET /healthz` on port `8080`.

The Streamlit demo UI runs on port `8501` when using Docker Compose.

## API Surface

- `POST /auth/login`
- `GET /me`
- `POST /documents`
- `GET /documents`
- `GET /documents/{id}`
- `DELETE /documents/{id}`
- `POST /chat/sessions`
- `GET /chat/sessions/{id}`
- `POST /chat/sessions/{id}/messages`
- `GET /debug/retrieval`
- `POST /eval/runs`
- `GET /eval/runs/{id}`
- `POST /internal/jobs/{job_id}/process`

## Validation

```bash
make test
make vet
python3 -m pytest eval-runner
python3 -m py_compile ui/streamlit/app.py
```

## Documentation

- [Docs Index](docs/README.md)
- [FR, NFR, and Scale Estimation](docs/01-fr-nfr-scale-estimation.md)
- [HLD and Component Design](docs/02-hld-component-design.md)
- [Use Cases and Sequence Diagrams](docs/03-usecases-sequence-diagrams.md)
- [LLD](docs/04-lld.md)
- [DB Design and Schema](docs/05-db-design-schema.md)
- [Architecture](docs/architecture.md)
- [Implementation Plan](docs/implementation-plan.md)
- [Deployment](deploy/README.md)
- [Cloud Run Deployment](deploy/cloud-run.md)
- [Storage Notes](docs/storage.md)
- [Evaluation](docs/evaluation.md)
- [Future Multi-Tenancy](docs/multitenancy.md)
