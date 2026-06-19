package retrieval

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/providers/mock"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

func TestRetrieveScopesDenseHydrationToRequester(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	docA := uuid.New()
	docB := uuid.New()
	store := &tenantChunkStore{
		allowedUserID: userB,
		rows: map[uuid.UUID]db.GetChunkByVectorIDRow{
			docB: vectorRow(docB, "User B handbook content"),
		},
	}
	vector := &capturingVector{
		hits: []rag.RetrievalHit{
			{Chunk: rag.Chunk{VectorID: docA.String() + ":0"}, DenseScore: 0.99},
			{Chunk: rag.Chunk{VectorID: docB.String() + ":0"}, DenseScore: 0.98},
		},
	}
	lexical := &capturingLexical{}
	service := NewService(store, mock.Embeddings{Dimension: 8}, vector, lexical, nil)

	result, err := service.Retrieve(context.Background(), rag.RetrievalRequest{
		UserID: userB,
		Query:  "PROJECT_FALCON",
		TopK:   5,
	})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if got := vector.filter["owner_user_id"].(map[string]any)["$eq"]; got != userB.String() {
		t.Fatalf("dense owner filter = %v, want %s", got, userB)
	}
	if store.seenOwner != userB {
		t.Fatalf("hydration owner = %s, want %s", store.seenOwner, userB)
	}
	if lexical.userID != userB {
		t.Fatalf("lexical owner = %s, want %s", lexical.userID, userB)
	}
	if len(result.DenseHits) != 1 {
		t.Fatalf("expected one requester-owned dense hit, got %d", len(result.DenseHits))
	}
	if result.DenseHits[0].Chunk.DocumentID != docB {
		t.Fatalf("hydrated document = %s, want %s", result.DenseHits[0].Chunk.DocumentID, docB)
	}
	if userA == userB {
		t.Fatal("test setup error: users must differ")
	}
}

func TestRetrieveSkipsDeletedOrUnauthorizedStaleVectors(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	store := &tenantChunkStore{allowedUserID: userID, missing: true}
	vector := &capturingVector{
		hits: []rag.RetrievalHit{
			{Chunk: rag.Chunk{VectorID: docID.String() + ":0"}, DenseScore: 0.99},
		},
	}
	service := NewService(store, mock.Embeddings{Dimension: 8}, vector, &capturingLexical{}, nil)

	result, err := service.Retrieve(context.Background(), rag.RetrievalRequest{
		UserID: userID,
		Query:  "deleted policy",
		TopK:   5,
	})
	if err != nil {
		t.Fatalf("retrieve should skip stale vector without failing: %v", err)
	}
	if len(result.DenseHits) != 0 || len(result.FusedHits) != 0 || len(result.RerankedHits) != 0 {
		t.Fatalf("expected stale vector to produce no hydrated hits: %#v", result)
	}
}

func TestPostgresFTSSearchScopesToRequester(t *testing.T) {
	userID := uuid.New()
	store := &capturingFTSStore{}
	fts := NewPostgresFTS(store)

	if _, err := fts.Search(context.Background(), userID, "PROJECT_FALCON", 3); err != nil {
		t.Fatalf("fts search: %v", err)
	}
	if store.params.OwnerUserID != userID {
		t.Fatalf("fts owner = %s, want %s", store.params.OwnerUserID, userID)
	}
	if store.params.WebsearchToTsquery != "PROJECT_FALCON" || store.params.Limit != 3 {
		t.Fatalf("unexpected fts params: %#v", store.params)
	}
}

type tenantChunkStore struct {
	allowedUserID uuid.UUID
	rows          map[uuid.UUID]db.GetChunkByVectorIDRow
	seenOwner     uuid.UUID
	missing       bool
}

func (s *tenantChunkStore) GetChunkByVectorID(_ context.Context, arg db.GetChunkByVectorIDParams) (db.GetChunkByVectorIDRow, error) {
	s.seenOwner = arg.OwnerUserID
	if s.missing || arg.OwnerUserID != s.allowedUserID {
		return db.GetChunkByVectorIDRow{}, pgx.ErrNoRows
	}
	row, ok := s.rows[arg.DocumentID]
	if !ok {
		return db.GetChunkByVectorIDRow{}, pgx.ErrNoRows
	}
	return row, nil
}

type capturingVector struct {
	hits   []rag.RetrievalHit
	filter map[string]any
}

func (v *capturingVector) UpsertChunks(context.Context, []rag.VectorRecord) error { return nil }
func (v *capturingVector) Search(_ context.Context, _ []float32, _ int, filter map[string]any) ([]rag.RetrievalHit, error) {
	v.filter = filter
	return v.hits, nil
}
func (v *capturingVector) DeleteDocument(context.Context, uuid.UUID) error { return nil }
func (v *capturingVector) Healthcheck(context.Context) error               { return nil }

type capturingLexical struct {
	userID uuid.UUID
}

func (l *capturingLexical) Search(_ context.Context, userID uuid.UUID, _ string, _ int) ([]rag.RetrievalHit, error) {
	l.userID = userID
	return nil, nil
}

type capturingFTSStore struct {
	params db.SearchChunksFTSParams
}

func (s *capturingFTSStore) SearchChunksFTS(_ context.Context, arg db.SearchChunksFTSParams) ([]db.SearchChunksFTSRow, error) {
	s.params = arg
	return nil, nil
}

func vectorRow(documentID uuid.UUID, content string) db.GetChunkByVectorIDRow {
	return db.GetChunkByVectorIDRow{
		ID:         uuid.New(),
		DocumentID: documentID,
		ChunkIndex: 0,
		Content:    content,
		PageNumber: pgtype.Int4{Valid: false},
		TokenCount: 4,
		Metadata:   []byte(`{}`),
		CreatedAt:  pgtype.Timestamptz{},
		Filename:   "handbook.md",
	}
}
