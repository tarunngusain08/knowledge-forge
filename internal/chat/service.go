package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/RAG-bot/internal/db"
	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type Store interface {
	CreateChatSession(ctx context.Context, arg db.CreateChatSessionParams) (db.ChatSession, error)
	GetChatSession(ctx context.Context, arg db.GetChatSessionParams) (db.ChatSession, error)
	ListChatMessages(ctx context.Context, sessionID uuid.UUID) ([]db.ChatMessage, error)
	CreateChatMessage(ctx context.Context, arg db.CreateChatMessageParams) (db.ChatMessage, error)
	CreateCitation(ctx context.Context, arg db.CreateCitationParams) (db.Citation, error)
	CreateRetrievalTrace(ctx context.Context, arg db.CreateRetrievalTraceParams) (db.RetrievalTrace, error)
}

type Service struct {
	store     Store
	llm       rag.LLMProvider
	retriever rag.Retriever
}

type AskRequest struct {
	UserID          uuid.UUID `json:"user_id"`
	SessionID       uuid.UUID `json:"session_id"`
	Question        string    `json:"question"`
	TopK            int       `json:"top_k"`
	RerankerEnabled bool      `json:"reranker_enabled"`
}

type AskResponse struct {
	UserMessage      db.ChatMessage      `json:"user_message"`
	AssistantMessage db.ChatMessage      `json:"assistant_message"`
	Answer           string              `json:"answer"`
	Citations        []rag.Citation      `json:"citations"`
	Retrieval        rag.RetrievalResult `json:"retrieval"`
	Model            string              `json:"model"`
	InputTokens      int                 `json:"input_tokens"`
	OutputTokens     int                 `json:"output_tokens"`
}

func NewService(store Store, llm rag.LLMProvider, retriever rag.Retriever) *Service {
	return &Service{store: store, llm: llm, retriever: retriever}
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, title string) (db.ChatSession, error) {
	if title == "" {
		title = "New chat"
	}
	return s.store.CreateChatSession(ctx, db.CreateChatSessionParams{UserID: userID, Title: title})
}

func (s *Service) GetSession(ctx context.Context, userID, sessionID uuid.UUID) (db.ChatSession, []db.ChatMessage, error) {
	session, err := s.store.GetChatSession(ctx, db.GetChatSessionParams{ID: sessionID, UserID: userID})
	if err != nil {
		return db.ChatSession{}, nil, err
	}
	messages, err := s.store.ListChatMessages(ctx, sessionID)
	return session, messages, err
}

func (s *Service) Ask(ctx context.Context, req AskRequest) (AskResponse, error) {
	_, messages, err := s.GetSession(ctx, req.UserID, req.SessionID)
	if err != nil {
		return AskResponse{}, fmt.Errorf("get session: %w", err)
	}
	history := toRAGMessages(messages)
	rewritten, err := s.llm.RewriteQuestion(ctx, req.Question, history)
	if err != nil {
		return AskResponse{}, fmt.Errorf("rewrite question: %w", err)
	}
	userMsg, err := s.store.CreateChatMessage(ctx, db.CreateChatMessageParams{
		SessionID:      req.SessionID,
		Role:           "user",
		Content:        req.Question,
		RewrittenQuery: pgtype.Text{String: rewritten, Valid: true},
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("store user message: %w", err)
	}
	retrieval, err := s.retriever.Retrieve(ctx, rag.RetrievalRequest{
		UserID:          req.UserID,
		SessionID:       req.SessionID,
		Query:           rewritten,
		TopK:            req.TopK,
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("retrieve context: %w", err)
	}
	generation, err := s.llm.GenerateAnswer(ctx, rag.GenerateRequest{
		Query:          req.Question,
		RewrittenQuery: rewritten,
		Context:        retrieval.RerankedHits,
		ChatHistory:    history,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("generate answer: %w", err)
	}
	assistantMsg, err := s.store.CreateChatMessage(ctx, db.CreateChatMessageParams{
		SessionID:      req.SessionID,
		Role:           "assistant",
		Content:        generation.Answer,
		RewrittenQuery: pgtype.Text{Valid: false},
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("store assistant message: %w", err)
	}
	for _, citation := range generation.Citations {
		_, _ = s.store.CreateCitation(ctx, citationParams(assistantMsg.ID, citation))
	}
	_, _ = s.store.CreateRetrievalTrace(ctx, traceParams(req, rewritten, retrieval, generation))
	return AskResponse{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
		Answer:           generation.Answer,
		Citations:        generation.Citations,
		Retrieval:        retrieval,
		Model:            generation.Model,
		InputTokens:      generation.InputTokens,
		OutputTokens:     generation.OutputTokens,
	}, nil
}

func toRAGMessages(messages []db.ChatMessage) []rag.Message {
	out := make([]rag.Message, 0, len(messages))
	for _, message := range messages {
		createdAt := time.Time{}
		if message.CreatedAt.Valid {
			createdAt = message.CreatedAt.Time
		}
		out = append(out, rag.Message{Role: message.Role, Content: message.Content, CreatedAt: createdAt})
	}
	return out
}

func citationParams(messageID uuid.UUID, citation rag.Citation) db.CreateCitationParams {
	metadata := mustJSON(citation.Metadata)
	return db.CreateCitationParams{
		MessageID:    messageID,
		ChunkID:      nullableUUID(citation.ChunkID),
		DocumentID:   nullableUUID(citation.DocumentID),
		DocumentName: citation.Document,
		PageNumber:   nullableInt(citation.PageNumber),
		Excerpt:      citation.Excerpt,
		DenseScore:   nullableFloat(citation.DenseScore),
		LexicalRank:  nullableIntValue(citation.LexicalRank),
		FusedRank:    nullableIntValue(citation.FusedRank),
		RerankScore:  nullableFloat(citation.RerankScore),
		Metadata:     metadata,
	}
}

func traceParams(req AskRequest, rewritten string, retrieval rag.RetrievalResult, generation rag.GenerateResponse) db.CreateRetrievalTraceParams {
	return db.CreateRetrievalTraceParams{
		UserID:           req.UserID,
		SessionID:        nullableUUID(req.SessionID),
		OriginalQuery:    req.Question,
		RewrittenQuery:   rewritten,
		TopK:             int32(req.TopK),
		RerankerEnabled:  req.RerankerEnabled,
		DenseHits:        mustJSON(retrieval.DenseHits),
		LexicalHits:      mustJSON(retrieval.LexicalHits),
		FusedHits:        mustJSON(retrieval.FusedHits),
		RerankedHits:     mustJSON(retrieval.RerankedHits),
		PromptPreview:    rag.BuildGroundedPrompt(rag.GenerateRequest{RewrittenQuery: rewritten, Context: retrieval.RerankedHits}),
		LatencyMs:        int32(retrieval.Latency.Milliseconds()),
		EstimatedCostUsd: "0",
	}
}

func nullableUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func nullableInt(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func nullableIntValue(value int) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func nullableFloat(value float64) pgtype.Float8 {
	if value == 0 {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: value, Valid: true}
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}
