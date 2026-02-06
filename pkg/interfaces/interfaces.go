package interfaces

import (
	"context"

	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
)

// RepositoryClient defines the interface for GitHub operations (SOLID: Interface Segregation)
type RepositoryClient interface {
	// ListRepositories finds all repositories matching the filter
	ListRepositories(ctx context.Context, org, keyword string) ([]*models.Repository, error)

	// GetChangedFiles detects files that changed since last sync
	GetChangedFiles(ctx context.Context, repo *models.Repository, lastCommitSHA string) ([]*models.FileChange, error)

	// GetFileContent retrieves content of a specific file
	GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error)

	// GetLatestCommitSHA gets the latest commit SHA for a repository
	GetLatestCommitSHA(ctx context.Context, owner, repo, branch string) (string, error)
}

// DocumentProcessor defines the interface for document processing (SOLID: Single Responsibility)
type DocumentProcessor interface {
	// ChunkDocument splits a document into smaller chunks
	ChunkDocument(ctx context.Context, fileChange *models.FileChange, maxSize, overlap int) ([]*models.Document, error)

	// ValidateDocument checks if document should be processed
	ValidateDocument(fileChange *models.FileChange, allowedExtensions []string, excludePatterns []string) bool

	// CleanContent cleans and normalizes document content
	CleanContent(content string) string
}

// EmbeddingService defines the interface for generating embeddings (SOLID: Dependency Inversion)
type EmbeddingService interface {
	// GenerateEmbedding creates a vector embedding for text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateBatchEmbeddings creates embeddings for multiple texts
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// GetDimension returns the dimension of embeddings
	GetDimension() int
}

// VectorStore defines the interface for vector database operations (SOLID: Open/Closed)
type VectorStore interface {
	// UpsertVectors inserts or updates vectors
	UpsertVectors(ctx context.Context, embeddings []*models.Embedding) error

	// DeleteVectors removes vectors by IDs
	DeleteVectors(ctx context.Context, ids []string, namespace string) error

	// QueryVectors searches for similar vectors
	QueryVectors(ctx context.Context, vector []float32, topK int, namespace string) ([]*models.Embedding, error)

	// DescribeIndex gets index statistics
	DescribeIndex(ctx context.Context) (map[string]interface{}, error)

	// Health checks the connection health
	Health(ctx context.Context) error
}

// MetadataStore defines the interface for metadata persistence (SOLID: Single Responsibility)
type MetadataStore interface {
	// SaveSyncMetadata stores sync state for a file
	SaveSyncMetadata(ctx context.Context, metadata *models.SyncMetadata) error

	// GetSyncMetadata retrieves sync state for a file
	GetSyncMetadata(ctx context.Context, projectID, repository, filePath string) (*models.SyncMetadata, error)

	// ListSyncMetadata lists all sync metadata for a project
	ListSyncMetadata(ctx context.Context, projectID string) ([]*models.SyncMetadata, error)

	// DeleteSyncMetadata removes sync metadata
	DeleteSyncMetadata(ctx context.Context, projectID, repository, filePath string) error

	// SaveProject stores project configuration
	SaveProject(ctx context.Context, project *models.Project) error

	// GetProject retrieves project configuration
	GetProject(ctx context.Context, projectID string) (*models.Project, error)

	// ListProjects lists all projects
	ListProjects(ctx context.Context) ([]*models.Project, error)

	// DeleteProject removes a project
	DeleteProject(ctx context.Context, projectID string) error
}

// NotificationService defines the interface for sending notifications (SOLID: Interface Segregation)
type NotificationService interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, payload *models.NotificationPayload) error

	// SendSlack sends a Slack notification
	SendSlack(ctx context.Context, payload *models.NotificationPayload) error
}

// SyncOrchestrator defines the interface for orchestrating sync operations (SOLID: Liskov Substitution)
type SyncOrchestrator interface {
	// SyncProject synchronizes a single project
	SyncProject(ctx context.Context, projectID string, incremental bool) (*models.SyncResult, error)

	// SyncAllProjects synchronizes all enabled projects
	SyncAllProjects(ctx context.Context, incremental bool) ([]*models.SyncResult, error)

	// ValidateSync validates the sync result
	ValidateSync(ctx context.Context, result *models.SyncResult) error
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	// Check performs a health check
	Check(ctx context.Context) (*models.HealthStatus, error)
}
