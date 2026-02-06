package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/errors"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
)

// MetadataService implements interfaces.MetadataStore
type MetadataService struct {
	db *sql.DB
}

// NewMetadataService creates a new metadata service
func NewMetadataService(dbPath string) (*MetadataService, error) {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	service := &MetadataService{db: db}
	if err := service.initSchema(); err != nil {
		return nil, err
	}

	return service, nil
}

// initSchema creates database tables
func (s *MetadataService) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sync_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id TEXT NOT NULL,
		repository TEXT NOT NULL,
		file_path TEXT NOT NULL,
		last_commit_sha TEXT NOT NULL,
		last_synced_at DATETIME NOT NULL,
		embedding_count INTEGER DEFAULT 0,
		status TEXT DEFAULT 'synced',
		UNIQUE(project_id, repository, file_path)
	);
	
	CREATE INDEX IF NOT EXISTS idx_sync_project ON sync_metadata(project_id);
	CREATE INDEX IF NOT EXISTS idx_sync_repo ON sync_metadata(repository);
	
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		organization TEXT NOT NULL,
		filter_keyword TEXT,
		namespace TEXT NOT NULL,
		enabled BOOLEAN DEFAULT 1,
		allowed_extensions TEXT,
		exclude_patterns TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Implement interfaces.MetadataStore methods

func (s *MetadataService) SaveSyncMetadata(ctx context.Context, metadata *models.SyncMetadata) error {
	query := `
		INSERT INTO sync_metadata (project_id, repository, file_path, last_commit_sha, last_synced_at, embedding_count, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, repository, file_path) DO UPDATE SET
			last_commit_sha = excluded.last_commit_sha,
			last_synced_at = excluded.last_synced_at,
			embedding_count = excluded.embedding_count,
			status = excluded.status
	`

	_, err := s.db.ExecContext(ctx, query,
		metadata.ProjectID, metadata.Repository, metadata.FilePath,
		metadata.LastCommitSHA, metadata.LastSyncedAt, metadata.EmbeddingCount, metadata.Status)

	if err != nil {
		return errors.Database("failed to save sync metadata", err)
	}

	return nil
}

func (s *MetadataService) GetSyncMetadata(ctx context.Context, projectID, repository, filePath string) (*models.SyncMetadata, error) {
	query := `SELECT id, project_id, repository, file_path, last_commit_sha, last_synced_at, embedding_count, status 
		FROM sync_metadata WHERE project_id = ? AND repository = ? AND file_path = ?`

	var metadata models.SyncMetadata
	err := s.db.QueryRowContext(ctx, query, projectID, repository, filePath).Scan(
		&metadata.ID, &metadata.ProjectID, &metadata.Repository, &metadata.FilePath,
		&metadata.LastCommitSHA, &metadata.LastSyncedAt, &metadata.EmbeddingCount, &metadata.Status)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("sync metadata")
	}
	if err != nil {
		return nil, errors.Database("failed to get sync metadata", err)
	}

	return &metadata, nil
}

func (s *MetadataService) ListSyncMetadata(ctx context.Context, projectID string) ([]*models.SyncMetadata, error) {
	query := `SELECT id, project_id, repository, file_path, last_commit_sha, last_synced_at, embedding_count, status 
		FROM sync_metadata WHERE project_id = ?`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, errors.Database("failed to list sync metadata", err)
	}
	defer rows.Close()

	var results []*models.SyncMetadata
	for rows.Next() {
		var metadata models.SyncMetadata
		if err := rows.Scan(&metadata.ID, &metadata.ProjectID, &metadata.Repository, &metadata.FilePath,
			&metadata.LastCommitSHA, &metadata.LastSyncedAt, &metadata.EmbeddingCount, &metadata.Status); err != nil {
			return nil, errors.Database("failed to scan sync metadata", err)
		}
		results = append(results, &metadata)
	}

	return results, nil
}

func (s *MetadataService) DeleteSyncMetadata(ctx context.Context, projectID, repository, filePath string) error {
	query := `DELETE FROM sync_metadata WHERE project_id = ? AND repository = ? AND file_path = ?`
	_, err := s.db.ExecContext(ctx, query, projectID, repository, filePath)
	if err != nil {
		return errors.Database("failed to delete sync metadata", err)
	}
	return nil
}

func (s *MetadataService) SaveProject(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, organization, filter_keyword, namespace, enabled, allowed_extensions, exclude_patterns, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			organization = excluded.organization,
			filter_keyword = excluded.filter_keyword,
			namespace = excluded.namespace,
			enabled = excluded.enabled,
			allowed_extensions = excluded.allowed_extensions,
			exclude_patterns = excluded.exclude_patterns,
			updated_at = excluded.updated_at
	`

	allowedExt := ""
	if len(project.AllowedExtensions) > 0 {
		data, _ := json.Marshal(project.AllowedExtensions)
		allowedExt = string(data)
	}

	excludePat := ""
	if len(project.ExcludePatterns) > 0 {
		data, _ := json.Marshal(project.ExcludePatterns)
		excludePat = string(data)
	}

	_, err := s.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Organization, project.FilterKeyword,
		project.Namespace, project.Enabled, allowedExt, excludePat, time.Now())

	if err != nil {
		return errors.Database("failed to save project", err)
	}

	return nil
}

func (s *MetadataService) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	query := `SELECT id, name, organization, filter_keyword, namespace, enabled, allowed_extensions, exclude_patterns, created_at, updated_at 
		FROM projects WHERE id = ?`

	var project models.Project
	var allowedExt, excludePat string

	err := s.db.QueryRowContext(ctx, query, projectID).Scan(
		&project.ID, &project.Name, &project.Organization, &project.FilterKeyword,
		&project.Namespace, &project.Enabled, &allowedExt, &excludePat,
		&project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("project")
	}
	if err != nil {
		return nil, errors.Database("failed to get project", err)
	}

	if allowedExt != "" {
		json.Unmarshal([]byte(allowedExt), &project.AllowedExtensions)
	}
	if excludePat != "" {
		json.Unmarshal([]byte(excludePat), &project.ExcludePatterns)
	}

	return &project, nil
}

func (s *MetadataService) ListProjects(ctx context.Context) ([]*models.Project, error) {
	query := `SELECT id, name, organization, filter_keyword, namespace, enabled, allowed_extensions, exclude_patterns, created_at, updated_at 
		FROM projects`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Database("failed to list projects", err)
	}
	defer rows.Close()

	var results []*models.Project
	for rows.Next() {
		var project models.Project
		var allowedExt, excludePat string

		if err := rows.Scan(&project.ID, &project.Name, &project.Organization, &project.FilterKeyword,
			&project.Namespace, &project.Enabled, &allowedExt, &excludePat,
			&project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, errors.Database("failed to scan project", err)
		}

		if allowedExt != "" {
			json.Unmarshal([]byte(allowedExt), &project.AllowedExtensions)
		}
		if excludePat != "" {
			json.Unmarshal([]byte(excludePat), &project.ExcludePatterns)
		}

		results = append(results, &project)
	}

	return results, nil
}

func (s *MetadataService) DeleteProject(ctx context.Context, projectID string) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return errors.Database("failed to delete project", err)
	}
	return nil
}

func (s *MetadataService) Close() error {
	return s.db.Close()
}

// HTTP Handlers
func (s *MetadataService) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "metadata-service"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Metadata Service on port %d", cfg.Services.MetadataServicePort)

	// Create metadata service
	service, err := NewMetadataService(cfg.Database.MetadataDBPath)
	if err != nil {
		logger.Fatal("Failed to create metadata service: %v", err)
	}
	defer service.Close()

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.MetadataServicePort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down metadata service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Metadata Service listening on port %d", cfg.Services.MetadataServicePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
