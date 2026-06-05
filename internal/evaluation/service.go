package evaluation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type Store interface {
	CreateEvalRun(ctx context.Context, arg db.CreateEvalRunParams) (db.EvalRun, error)
	UpdateEvalRun(ctx context.Context, arg db.UpdateEvalRunParams) (db.EvalRun, error)
	GetEvalRun(ctx context.Context, arg db.GetEvalRunParams) (db.EvalRun, error)
}

type Service struct {
	store     Store
	retriever rag.Retriever
}

type Question struct {
	Question          string   `json:"question"`
	ExpectedVectorIDs []string `json:"expected_vector_ids"`
}

type RunRequest struct {
	UserID          uuid.UUID  `json:"user_id"`
	Name            string     `json:"name"`
	TopK            int        `json:"top_k"`
	RerankerEnabled bool       `json:"reranker_enabled"`
	Questions       []Question `json:"questions"`
}

type RunResponse struct {
	Run     db.EvalRun       `json:"run"`
	Metrics Metrics          `json:"metrics"`
	Results []QuestionResult `json:"results"`
}

func NewService(store Store, retriever rag.Retriever) *Service {
	return &Service{store: store, retriever: retriever}
}

func (s *Service) Run(ctx context.Context, req RunRequest) (RunResponse, error) {
	if req.TopK == 0 {
		req.TopK = 5
	}
	config := mustJSON(map[string]any{"top_k": req.TopK, "reranker_enabled": req.RerankerEnabled})
	run, err := s.store.CreateEvalRun(ctx, db.CreateEvalRunParams{
		UserID: req.UserID,
		Name:   req.Name,
		Config: config,
	})
	if err != nil {
		return RunResponse{}, fmt.Errorf("create eval run: %w", err)
	}
	results := make([]QuestionResult, 0, len(req.Questions))
	for _, question := range req.Questions {
		retrieval, err := s.retriever.Retrieve(ctx, rag.RetrievalRequest{
			UserID:          req.UserID,
			Query:           question.Question,
			TopK:            req.TopK,
			RerankerEnabled: req.RerankerEnabled,
		})
		if err != nil {
			_, _ = s.store.UpdateEvalRun(ctx, db.UpdateEvalRunParams{
				ID:           run.ID,
				Metrics:      []byte(`{}`),
				Status:       "failed",
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			})
			return RunResponse{}, err
		}
		results = append(results, QuestionResult{
			Question:           question.Question,
			ExpectedVectorIDs:  question.ExpectedVectorIDs,
			RetrievedVectorIDs: vectorIDs(retrieval.RerankedHits),
			LatencyMs:          retrieval.Latency.Milliseconds(),
		})
	}
	metrics := Compute(results)
	metricsJSON := mustJSON(map[string]any{"retrieval": metrics, "results": results})
	run, err = s.store.UpdateEvalRun(ctx, db.UpdateEvalRunParams{
		ID:           run.ID,
		Metrics:      metricsJSON,
		Status:       "succeeded",
		ErrorMessage: pgtype.Text{Valid: false},
	})
	if err != nil {
		return RunResponse{}, err
	}
	return RunResponse{Run: run, Metrics: metrics, Results: results}, nil
}

func (s *Service) Get(ctx context.Context, userID, runID uuid.UUID) (db.EvalRun, error) {
	return s.store.GetEvalRun(ctx, db.GetEvalRunParams{ID: runID, UserID: userID})
}

func vectorIDs(hits []rag.RetrievalHit) []string {
	ids := make([]string, 0, len(hits))
	for _, hit := range hits {
		ids = append(ids, hit.Chunk.VectorID)
	}
	return ids
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}
