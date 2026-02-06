package models

import "time"

// Repository represents a GitHub repository
type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Owner         string    `json:"owner"`
	DefaultBranch string    `json:"default_branch"`
	LastCommit    string    `json:"last_commit"`
	UpdatedAt     time.Time `json:"updated_at"`
	Private       bool      `json:"private"`
}

// FileChange represents a changed file in a repository
type FileChange struct {
	Repository   string    `json:"repository"`
	FilePath     string    `json:"file_path"`
	Content      string    `json:"content"`
	CommitSHA    string    `json:"commit_sha"`
	LastModified time.Time `json:"last_modified"`
	ChangeType   string    `json:"change_type"` // added, modified, deleted
	Size         int64     `json:"size"`
}

// Document represents a processed document chunk
type Document struct {
	ID           string            `json:"id"`
	Repository   string            `json:"repository"`
	FilePath     string            `json:"file_path"`
	Content      string            `json:"content"`
	ChunkIndex   int               `json:"chunk_index"`
	TotalChunks  int               `json:"total_chunks"`
	Metadata     map[string]string `json:"metadata"`
	CommitSHA    string            `json:"commit_sha"`
	LastModified time.Time         `json:"last_modified"`
}

// Embedding represents a vector embedding
type Embedding struct {
	ID         string            `json:"id"`
	Vector     []float32         `json:"vector"`
	Metadata   map[string]string `json:"metadata"`
	Repository string            `json:"repository"`
	FilePath   string            `json:"file_path"`
	Namespace  string            `json:"namespace"`
}

// SyncMetadata tracks synchronization state
type SyncMetadata struct {
	ID             int64     `json:"id"`
	ProjectID      string    `json:"project_id"`
	Repository     string    `json:"repository"`
	FilePath       string    `json:"file_path"`
	LastCommitSHA  string    `json:"last_commit_sha"`
	LastSyncedAt   time.Time `json:"last_synced_at"`
	EmbeddingCount int       `json:"embedding_count"`
	Status         string    `json:"status"`
}

// Project represents a multi-project configuration
type Project struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Organization      string    `json:"organization"`
	FilterKeyword     string    `json:"filter_keyword"`
	Namespace         string    `json:"namespace"`
	Enabled           bool      `json:"enabled"`
	AllowedExtensions []string  `json:"allowed_extensions"`
	ExcludePatterns   []string  `json:"exclude_patterns"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SyncResult represents the outcome of a sync operation
type SyncResult struct {
	ProjectID           string        `json:"project_id"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	RepositoriesScanned int           `json:"repositories_scanned"`
	FilesDiscovered     int           `json:"files_discovered"`
	FilesChanged        int           `json:"files_changed"`
	FilesProcessed      int           `json:"files_processed"`
	ChunksCreated       int           `json:"chunks_created"`
	EmbeddingsGenerated int           `json:"embeddings_generated"`
	VectorsUpserted     int           `json:"vectors_upserted"`
	VectorsDeleted      int           `json:"vectors_deleted"`
	Errors              []string      `json:"errors"`
	Warnings            []string      `json:"warnings"`
	Success             bool          `json:"success"`
}

// NotificationPayload represents data for notifications
type NotificationPayload struct {
	Type      string      `json:"type"` // success, error, warning
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Result    *SyncResult `json:"result,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthStatus represents service health
type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"` // healthy, unhealthy, degraded
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
