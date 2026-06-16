package codeintel

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	LanguageGo         = "go"
	LanguageTypeScript = "typescript"
	LanguageTSX        = "tsx"
	LanguageText       = "text"
	LanguageMarkdown   = "markdown"

	ChunkFile = "file"
)

type Repository struct {
	ID            uuid.UUID `json:"id"`
	OwnerUserID   uuid.UUID `json:"owner_user_id"`
	Name          string    `json:"name"`
	RemoteURL     string    `json:"remote_url"`
	LocalPath     string    `json:"local_path"`
	DefaultBranch string    `json:"default_branch"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Snapshot struct {
	ID           uuid.UUID  `json:"id"`
	RepositoryID uuid.UUID  `json:"repository_id"`
	BranchName   string     `json:"branch_name"`
	CommitSHA    string     `json:"commit_sha"`
	Status       string     `json:"status"`
	FileCount    int        `json:"file_count"`
	SymbolCount  int        `json:"symbol_count"`
	ChunkCount   int        `json:"chunk_count"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	IndexedAt    *time.Time `json:"indexed_at,omitempty"`
}

type IngestionJob struct {
	ID           uuid.UUID  `json:"id"`
	RepositoryID uuid.UUID  `json:"repository_id"`
	BranchName   string     `json:"branch_name"`
	CommitSHA    string     `json:"commit_sha"`
	Status       string     `json:"status"`
	Attempts     int        `json:"attempts"`
	MaxAttempts  int        `json:"max_attempts"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	LockedBy     *string    `json:"locked_by,omitempty"`
	RequestedBy  *uuid.UUID `json:"requested_by,omitempty"`
	SnapshotID   *uuid.UUID `json:"snapshot_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type FileVersion struct {
	ID           uuid.UUID `json:"id"`
	RepositoryID uuid.UUID `json:"repository_id"`
	SnapshotID   uuid.UUID `json:"snapshot_id"`
	FileID       uuid.UUID `json:"file_id"`
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	ContentHash  string    `json:"content_hash"`
	Content      string    `json:"content,omitempty"`
	SizeBytes    int64     `json:"size_bytes"`
	IsSupported  bool      `json:"is_supported"`
	IsTest       bool      `json:"is_test"`
}

type Symbol struct {
	ID            uuid.UUID      `json:"id"`
	RepositoryID  uuid.UUID      `json:"repository_id"`
	SnapshotID    uuid.UUID      `json:"snapshot_id"`
	FileID        uuid.UUID      `json:"file_id"`
	Name          string         `json:"name"`
	QualifiedName string         `json:"qualified_name"`
	Kind          string         `json:"kind"`
	Language      string         `json:"language"`
	Signature     string         `json:"signature"`
	Content       string         `json:"content,omitempty"`
	StartLine     int            `json:"start_line"`
	EndLine       int            `json:"end_line"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Chunk struct {
	ID            uuid.UUID      `json:"id"`
	RepositoryID  uuid.UUID      `json:"repository_id"`
	SnapshotID    uuid.UUID      `json:"snapshot_id"`
	FileID        uuid.UUID      `json:"file_id"`
	FileVersionID uuid.UUID      `json:"file_version_id"`
	SymbolID      *uuid.UUID     `json:"symbol_id,omitempty"`
	ChunkIndex    int            `json:"chunk_index"`
	ChunkType     string         `json:"chunk_type"`
	Path          string         `json:"path"`
	Language      string         `json:"language"`
	StartLine     int            `json:"start_line"`
	EndLine       int            `json:"end_line"`
	Content       string         `json:"content"`
	TokenCount    int            `json:"token_count"`
	Metadata      map[string]any `json:"metadata"`
}

type Worktree struct {
	Path      string
	Branch    string
	CommitSHA string
	Cleanup   func() error
}

type GitCommit struct {
	SHA             string           `json:"sha"`
	ParentSHAs      []string         `json:"parent_shas"`
	AuthorName      string           `json:"author_name"`
	AuthorEmailHash string           `json:"author_email_hash"`
	Message         string           `json:"message"`
	CommittedAt     time.Time        `json:"committed_at"`
	Files           []GitChangedFile `json:"files"`
}

type GitChangedFile struct {
	Path       string `json:"path"`
	ChangeType string `json:"change_type"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
}

type GitProvider interface {
	ResolveWorktree(ctx context.Context, repo Repository, branchName string, commitSHA string) (Worktree, error)
	RecentCommits(ctx context.Context, worktree Worktree, limit int) ([]GitCommit, error)
}
