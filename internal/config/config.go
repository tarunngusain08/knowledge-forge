package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName           string
	Environment           string
	HTTPPort              string
	LogLevel              string
	DatabaseURL           string
	JWTSecret             string
	AdminEmail            string
	AdminPassword         string
	ProviderMode          string
	GoogleProjectID       string
	GoogleLocation        string
	VertexEmbedModel      string
	VertexChatModel       string
	VertexRankingModel    string
	VertexRankingLocation string
	PineconeAPIKey        string
	PineconeHost          string
	PineconeNamespace     string
	WorkerName            string
	WorkerBatchSize       int64
	InternalWorkerToken   string
	AllowLocalRepoPaths   bool
	AllowedGitRemoteHosts []string
	ShutdownTimeout       time.Duration
	MaxUploadBytes        int64
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:           getenv("SERVICE_NAME", "knowledge-forge-api"),
		Environment:           getenv("ENVIRONMENT", "local"),
		HTTPPort:              getenv("HTTP_PORT", "8080"),
		LogLevel:              getenv("LOG_LEVEL", "info"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		AdminEmail:            os.Getenv("ADMIN_EMAIL"),
		AdminPassword:         os.Getenv("ADMIN_PASSWORD"),
		ProviderMode:          getenv("PROVIDER_MODE", "mock"),
		GoogleProjectID:       os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GoogleLocation:        getenv("GOOGLE_CLOUD_LOCATION", "us-central1"),
		VertexEmbedModel:      getenv("VERTEX_EMBEDDING_MODEL", "gemini-embedding-001"),
		VertexChatModel:       getenv("VERTEX_CHAT_MODEL", "gemini-2.5-flash"),
		VertexRankingModel:    getenv("VERTEX_RANKING_MODEL", "semantic-ranker-default@latest"),
		VertexRankingLocation: getenv("VERTEX_RANKING_LOCATION", "global"),
		PineconeAPIKey:        os.Getenv("PINECONE_API_KEY"),
		PineconeHost:          os.Getenv("PINECONE_HOST"),
		PineconeNamespace:     getenv("PINECONE_NAMESPACE", "default"),
		WorkerName:            getenv("WORKER_NAME", "knowledge-forge-worker"),
		WorkerBatchSize:       getInt64("WORKER_BATCH_SIZE", 2),
		InternalWorkerToken:   os.Getenv("INTERNAL_WORKER_TOKEN"),
		AllowLocalRepoPaths:   getBool("ALLOW_LOCAL_REPOSITORY_PATHS", false),
		AllowedGitRemoteHosts: getCSV("ALLOWED_GIT_REMOTE_HOSTS"),
		ShutdownTimeout:       getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		MaxUploadBytes:        getInt64("MAX_UPLOAD_BYTES", 20*1024*1024),
	}

	if cfg.MaxUploadBytes <= 0 {
		return Config{}, errors.New("MAX_UPLOAD_BYTES must be positive")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getCSV(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
