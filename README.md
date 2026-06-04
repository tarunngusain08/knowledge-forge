# RAG-bot

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
make test
docker compose up --build
```

The API exposes `GET /healthz` on port `8080`.

## Documentation

- [Architecture](docs/architecture.md)
- [Implementation Plan](docs/implementation-plan.md)
- [Deployment](deploy/README.md)
