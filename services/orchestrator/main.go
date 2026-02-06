package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
)

// Orchestrator coordinates all microservices
type Orchestrator struct {
	githubServiceURL       string
	documentProcessorURL   string
	embeddingServiceURL    string
	vectorStorageURL       string
	notificationServiceURL string
	metadataServiceURL     string
	httpClient             *http.Client
	config                 *config.Config
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		githubServiceURL:       getServiceURL("GITHUB_SERVICE_URL", "http://localhost:8081"),
		documentProcessorURL:   getServiceURL("DOCUMENT_PROCESSOR_URL", "http://localhost:8082"),
		embeddingServiceURL:    getServiceURL("EMBEDDING_SERVICE_URL", "http://localhost:8083"),
		vectorStorageURL:       getServiceURL("VECTOR_STORAGE_URL", "http://localhost:8084"),
		notificationServiceURL: getServiceURL("NOTIFICATION_SERVICE_URL", "http://localhost:8085"),
		metadataServiceURL:     getServiceURL("METADATA_SERVICE_URL", "http://localhost:8086"),
		httpClient:             &http.Client{Timeout: 60 * time.Second},
		config:                 cfg,
	}
}

// SyncProject synchronizes a single project
func (o *Orchestrator) SyncProject(ctx context.Context, projectID string, incremental bool) (*models.SyncResult, error) {
	result := &models.SyncResult{
		ProjectID: projectID,
		StartTime: time.Now(),
		Success:   false,
	}

	logger.Info("Starting sync for project: %s (incremental: %v)", projectID, incremental)

	// Step 1: Discover repositories from GitHub
	repos, err := o.discoverRepositories(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to discover repositories: %v", err))
		o.sendNotification(ctx, result, "error")
		return result, err
	}
	result.RepositoriesScanned = len(repos)
	logger.Info("Discovered %d repositories", len(repos))

	// Step 2: Process each repository
	var allChangedFiles []*models.FileChange
	for _, repo := range repos {
		// Get last commit SHA if incremental
		lastCommitSHA := ""
		if incremental {
			lastCommitSHA, _ = o.getLastCommitSHA(ctx, projectID, repo.FullName)
		}

		// Detect changed files
		changedFiles, err := o.getChangedFiles(ctx, repo, lastCommitSHA)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to get changed files for %s: %v", repo.FullName, err))
			continue
		}

		allChangedFiles = append(allChangedFiles, changedFiles...)
	}

	result.FilesDiscovered = len(allChangedFiles)
	result.FilesChanged = len(allChangedFiles)
	logger.Info("Found %d changed files", len(allChangedFiles))

	// Step 3: Filter and process files
	validFiles := o.filterFiles(allChangedFiles)
	result.FilesProcessed = len(validFiles)

	// Step 4: Process files in batches
	embeddings, chunks, err := o.processFiles(ctx, validFiles)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to process files: %v", err))
		o.sendNotification(ctx, result, "error")
		return result, err
	}

	result.ChunksCreated = chunks
	result.EmbeddingsGenerated = len(embeddings)

	// Step 5: Upsert to vector database
	if len(embeddings) > 0 {
		if err := o.upsertVectors(ctx, embeddings, projectID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to upsert vectors: %v", err))
			o.sendNotification(ctx, result, "error")
			return result, err
		}
		result.VectorsUpserted = len(embeddings)
	}

	// Step 6: Update metadata
	for _, file := range validFiles {
		metadata := &models.SyncMetadata{
			ProjectID:      projectID,
			Repository:     file.Repository,
			FilePath:       file.FilePath,
			LastCommitSHA:  file.CommitSHA,
			LastSyncedAt:   time.Now(),
			EmbeddingCount: 0, // Would need to track per file
			Status:         "synced",
		}
		_ = o.saveMetadata(ctx, metadata)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = true

	logger.Info("Sync completed successfully: %d embeddings in %s", result.EmbeddingsGenerated, result.Duration)

	// Send success notification
	o.sendNotification(ctx, result, "success")

	return result, nil
}

// discoverRepositories gets repositories from GitHub service
func (o *Orchestrator) discoverRepositories(ctx context.Context) ([]*models.Repository, error) {
	url := fmt.Sprintf("%s/repositories?org=%s&keyword=%s",
		o.githubServiceURL, o.config.GitHub.Organization, o.config.GitHub.FilterKeyword)

	resp, err := o.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repos []*models.Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// getChangedFiles gets changed files for a repository
func (o *Orchestrator) getChangedFiles(ctx context.Context, repo *models.Repository, lastCommitSHA string) ([]*models.FileChange, error) {
	url := fmt.Sprintf("%s/changes?repo=%s&last_commit=%s", o.githubServiceURL, repo.FullName, lastCommitSHA)

	resp, err := o.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var files []*models.FileChange
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

// filterFiles filters files based on extensions and patterns
func (o *Orchestrator) filterFiles(files []*models.FileChange) []*models.FileChange {
	var validFiles []*models.FileChange

	for _, file := range files {
		// Check extensions
		valid := false
		for _, ext := range o.config.Processing.AllowedExtensions {
			if len(file.FilePath) >= len(ext) && file.FilePath[len(file.FilePath)-len(ext):] == ext {
				valid = true
				break
			}
		}

		if !valid {
			continue
		}

		// Check exclude patterns
		excluded := false
		for _, pattern := range o.config.Processing.ExcludePatterns {
			if contains(file.FilePath, pattern) {
				excluded = true
				break
			}
		}

		if !excluded {
			validFiles = append(validFiles, file)
		}
	}

	return validFiles
}

// processFiles processes files into embeddings
func (o *Orchestrator) processFiles(ctx context.Context, files []*models.FileChange) ([]*models.Embedding, int, error) {
	var allEmbeddings []*models.Embedding
	totalChunks := 0

	// Process in batches
	batchSize := o.config.Processing.MaxWorkers
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		embeddings, chunks, err := o.processBatch(ctx, batch)
		if err != nil {
			return nil, 0, err
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
		totalChunks += chunks
	}

	return allEmbeddings, totalChunks, nil
}

// processBatch processes a batch of files
func (o *Orchestrator) processBatch(ctx context.Context, files []*models.FileChange) ([]*models.Embedding, int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allEmbeddings []*models.Embedding
	totalChunks := 0

	for _, file := range files {
		wg.Add(1)
		go func(f *models.FileChange) {
			defer wg.Done()

			// Chunk document
			documents, err := o.chunkDocument(ctx, f)
			if err != nil {
				logger.Warning("Failed to chunk document %s: %v", f.FilePath, err)
				return
			}

			// Generate embeddings
			embeddings, err := o.generateEmbeddings(ctx, documents)
			if err != nil {
				logger.Warning("Failed to generate embeddings for %s: %v", f.FilePath, err)
				return
			}

			mu.Lock()
			allEmbeddings = append(allEmbeddings, embeddings...)
			totalChunks += len(documents)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return allEmbeddings, totalChunks, nil
}

// chunkDocument chunks a document
func (o *Orchestrator) chunkDocument(ctx context.Context, file *models.FileChange) ([]*models.Document, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"file_change": file,
	})

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/chunk", o.documentProcessorURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Documents []*models.Document `json:"documents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Documents, nil
}

// generateEmbeddings generates embeddings for documents
func (o *Orchestrator) generateEmbeddings(ctx context.Context, documents []*models.Document) ([]*models.Embedding, error) {
	if len(documents) == 0 {
		return []*models.Embedding{}, nil
	}

	// Extract texts
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	// Call embedding service
	reqBody, _ := json.Marshal(map[string]interface{}{
		"texts": texts,
	})

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/embed", o.embeddingServiceURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Create embeddings
	embeddings := make([]*models.Embedding, len(documents))
	for i, doc := range documents {
		embeddings[i] = &models.Embedding{
			ID:         doc.ID,
			Vector:     result.Embeddings[i],
			Metadata:   doc.Metadata,
			Repository: doc.Repository,
			FilePath:   doc.FilePath,
			Namespace:  o.config.GitHub.Organization,
		}
	}

	return embeddings, nil
}

// upsertVectors upserts vectors to Pinecone
func (o *Orchestrator) upsertVectors(ctx context.Context, embeddings []*models.Embedding, namespace string) error {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"embeddings": embeddings,
	})

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/upsert", o.vectorStorageURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed: %s", body)
	}

	return nil
}

// saveMetadata saves sync metadata
func (o *Orchestrator) saveMetadata(ctx context.Context, metadata *models.SyncMetadata) error {
	reqBody, _ := json.Marshal(metadata)

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/metadata", o.metadataServiceURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// getLastCommitSHA gets the last synced commit SHA
func (o *Orchestrator) getLastCommitSHA(ctx context.Context, projectID, repository string) (string, error) {
	url := fmt.Sprintf("%s/metadata?project_id=%s&repository=%s", o.metadataServiceURL, projectID, repository)

	resp, err := o.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	var metadata models.SyncMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", err
	}

	return metadata.LastCommitSHA, nil
}

// sendNotification sends a notification
func (o *Orchestrator) sendNotification(ctx context.Context, result *models.SyncResult, notifType string) {
	title := "RepoSync Update"
	message := fmt.Sprintf("Processed %d files, generated %d embeddings in %s",
		result.FilesProcessed, result.EmbeddingsGenerated, result.Duration)

	if notifType == "error" {
		title = "RepoSync Failed"
		if len(result.Errors) > 0 {
			message = result.Errors[0]
		}
	}

	payload := &models.NotificationPayload{
		Type:      notifType,
		Title:     title,
		Message:   message,
		Result:    result,
		Timestamp: time.Now(),
	}

	reqBody, _ := json.Marshal(payload)
	_, _ = o.httpClient.Post(
		fmt.Sprintf("%s/notify", o.notificationServiceURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
}

// HTTP Handlers
func (o *Orchestrator) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		projectID = "default"
	}

	incremental := r.URL.Query().Get("incremental") == "true"

	result, err := o.SyncProject(r.Context(), projectID, incremental)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (o *Orchestrator) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Helper functions
func getServiceURL(envVar, defaultURL string) string {
	if url := os.Getenv(envVar); url != "" {
		return url
	}
	return defaultURL
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "orchestrator"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Orchestrator Service on port %d", cfg.Services.OrchestratorPort)

	// Create orchestrator
	orchestrator := NewOrchestrator(cfg)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", orchestrator.handleHealth)
	mux.HandleFunc("/sync", orchestrator.handleSync)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.OrchestratorPort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down orchestrator...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Orchestrator Service listening on port %d", cfg.Services.OrchestratorPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
