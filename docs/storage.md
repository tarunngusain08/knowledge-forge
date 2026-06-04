# Storage Notes: BYTEA vs GCS

v1 stores uploaded file bytes in PostgreSQL `BYTEA`.

Benefits:

- Transactional metadata + raw file writes
- Small local development footprint
- No extra object-storage setup for the portfolio demo
- Easier cleanup during document deletion

Tradeoffs:

- Large files increase database size quickly
- Backups become heavier
- Lifecycle policies and object-level retention are weaker than GCS
- Serving or reprocessing large files can pressure database I/O

Recommended production evolution:

- Store metadata, checksum, owner, and indexing state in PostgreSQL
- Store raw files in GCS
- Keep `gcs_uri`, `sha256`, and file size in the `documents` table
- Use GCS lifecycle policies for retention and archival

