package rag

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TaskRetrievalDocument = "RETRIEVAL_DOCUMENT"
	TaskRetrievalQuery    = "RETRIEVAL_QUERY"
)

type Chunk struct {
	ID         uuid.UUID      `json:"id"`
	DocumentID uuid.UUID      `json:"document_id"`
	VectorID   string         `json:"vector_id"`
	Index      int            `json:"index"`
	Content    string         `json:"content"`
	PageNumber *int           `json:"page_number,omitempty"`
	TokenCount int            `json:"token_count"`
	Metadata   map[string]any `json:"metadata"`
}

type ChunkInput struct {
	DocumentID uuid.UUID      `json:"document_id"`
	Filename   string         `json:"filename"`
	Content    string         `json:"content"`
	Metadata   map[string]any `json:"metadata"`
}

type EmbeddingResult struct {
	Vector      []float32 `json:"vector"`
	InputTokens int       `json:"input_tokens"`
	Model       string    `json:"model"`
}

type VectorRecord struct {
	ID       string         `json:"id"`
	Values   []float32      `json:"values"`
	Metadata map[string]any `json:"metadata"`
}

type RetrievalHit struct {
	Chunk       Chunk    `json:"chunk"`
	DenseScore  float64  `json:"dense_score,omitempty"`
	LexicalRank int      `json:"lexical_rank,omitempty"`
	FusedRank   int      `json:"fused_rank,omitempty"`
	RerankScore float64  `json:"rerank_score,omitempty"`
	Source      string   `json:"source"`
	Reasons     []string `json:"reasons,omitempty"`
}

type RerankDocument struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type RerankResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
	Rank  int     `json:"rank"`
}

type GenerateRequest struct {
	Query          string         `json:"query"`
	RewrittenQuery string         `json:"rewritten_query"`
	Context        []RetrievalHit `json:"context"`
	ChatHistory    []Message      `json:"chat_history"`
}

type GenerateResponse struct {
	Answer       string     `json:"answer"`
	InputTokens  int        `json:"input_tokens"`
	OutputTokens int        `json:"output_tokens"`
	Model        string     `json:"model"`
	Citations    []Citation `json:"citations"`
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Citation struct {
	ChunkID      uuid.UUID      `json:"chunk_id"`
	DocumentID   uuid.UUID      `json:"document_id"`
	Document     string         `json:"document"`
	RepositoryID uuid.UUID      `json:"repository_id,omitempty"`
	SnapshotID   uuid.UUID      `json:"snapshot_id,omitempty"`
	BranchName   string         `json:"branch_name,omitempty"`
	CommitSHA    string         `json:"commit_sha,omitempty"`
	Path         string         `json:"path,omitempty"`
	StartLine    int            `json:"start_line,omitempty"`
	EndLine      int            `json:"end_line,omitempty"`
	PageNumber   *int           `json:"page_number,omitempty"`
	Excerpt      string         `json:"excerpt"`
	DenseScore   float64        `json:"dense_score,omitempty"`
	LexicalRank  int            `json:"lexical_rank,omitempty"`
	FusedRank    int            `json:"fused_rank,omitempty"`
	RerankScore  float64        `json:"rerank_score,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type RetrievalRequest struct {
	UserID             uuid.UUID      `json:"user_id"`
	SessionID          uuid.UUID      `json:"session_id"`
	RepositoryID       uuid.UUID      `json:"repository_id,omitempty"`
	BranchName         string         `json:"branch_name,omitempty"`
	Query              string         `json:"query"`
	TopK               int            `json:"top_k"`
	CandidateK         int            `json:"candidate_k,omitempty"`
	QueryCategory      string         `json:"query_category,omitempty"`
	RetrievalPath      []string       `json:"retrieval_path,omitempty"`
	RetrievalConfig    map[string]any `json:"retrieval_config,omitempty"`
	ContextTokenBudget int            `json:"context_token_budget,omitempty"`
	RerankerEnabled    bool           `json:"reranker_enabled"`
}

type RetrievalResult struct {
	OriginalQuery      string         `json:"original_query"`
	RewrittenQuery     string         `json:"rewritten_query"`
	RepositoryID       uuid.UUID      `json:"repository_id,omitempty"`
	SnapshotID         uuid.UUID      `json:"snapshot_id,omitempty"`
	BranchName         string         `json:"branch_name,omitempty"`
	CommitSHA          string         `json:"commit_sha,omitempty"`
	DenseHits          []RetrievalHit `json:"dense_hits"`
	LexicalHits        []RetrievalHit `json:"lexical_hits"`
	SymbolHits         []RetrievalHit `json:"symbol_hits,omitempty"`
	GraphHits          []RetrievalHit `json:"graph_hits,omitempty"`
	FusedHits          []RetrievalHit `json:"fused_hits"`
	RerankedHits       []RetrievalHit `json:"reranked_hits"`
	QueryCategory      string         `json:"query_category,omitempty"`
	RetrievalPath      []string       `json:"retrieval_path,omitempty"`
	RetrievalConfig    map[string]any `json:"retrieval_config,omitempty"`
	ContextTokenCount  int            `json:"context_token_count,omitempty"`
	RetrievedChunkIDs  []uuid.UUID    `json:"retrieved_chunk_ids,omitempty"`
	StageContributions map[string]int `json:"stage_contributions,omitempty"`
	Latency            time.Duration  `json:"latency"`
}

type LLMProvider interface {
	GenerateAnswer(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	RewriteQuestion(ctx context.Context, question string, history []Message) (string, error)
}

type EmbeddingProvider interface {
	EmbedDocuments(ctx context.Context, texts []string) ([]EmbeddingResult, error)
	EmbedQuery(ctx context.Context, text string) (EmbeddingResult, error)
}

type VectorStoreProvider interface {
	UpsertChunks(ctx context.Context, records []VectorRecord) error
	Search(ctx context.Context, vector []float32, topK int, filter map[string]any) ([]RetrievalHit, error)
	DeleteDocument(ctx context.Context, documentID uuid.UUID) error
	Healthcheck(ctx context.Context) error
}

type RerankerProvider interface {
	Rerank(ctx context.Context, query string, documents []RerankDocument, topN int) ([]RerankResult, error)
}

type LexicalSearchProvider interface {
	Search(ctx context.Context, query string, topK int) ([]RetrievalHit, error)
}

type ChunkingProvider interface {
	Split(ctx context.Context, input ChunkInput) ([]Chunk, error)
}

type DocumentExtractor interface {
	Extract(ctx context.Context, filename string, content []byte) (string, error)
}

type Retriever interface {
	Retrieve(ctx context.Context, req RetrievalRequest) (RetrievalResult, error)
}
