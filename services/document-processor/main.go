package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
)

// DocumentProcessor implements interfaces.DocumentProcessor
type DocumentProcessor struct {
	maxChunkSize int
	chunkOverlap int
}

// NewDocumentProcessor creates a new document processor
func NewDocumentProcessor(maxChunkSize, chunkOverlap int) *DocumentProcessor {
	return &DocumentProcessor{
		maxChunkSize: maxChunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// ChunkDocument splits a document into smaller chunks
func (p *DocumentProcessor) ChunkDocument(ctx context.Context, fileChange *models.FileChange, maxSize, overlap int) ([]*models.Document, error) {
	content := p.CleanContent(fileChange.Content)

	if len(content) == 0 {
		return []*models.Document{}, nil
	}

	var chunks []string

	// Simple sentence-aware chunking
	if len(content) <= maxSize {
		chunks = []string{content}
	} else {
		chunks = p.splitIntoChunks(content, maxSize, overlap)
	}

	// Create documents
	documents := make([]*models.Document, len(chunks))
	for i, chunk := range chunks {
		docID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d", fileChange.Repository, fileChange.FilePath, i))))

		documents[i] = &models.Document{
			ID:           docID,
			Repository:   fileChange.Repository,
			FilePath:     fileChange.FilePath,
			Content:      chunk,
			ChunkIndex:   i,
			TotalChunks:  len(chunks),
			CommitSHA:    fileChange.CommitSHA,
			LastModified: fileChange.LastModified,
			Metadata: map[string]string{
				"repository":   fileChange.Repository,
				"file_path":    fileChange.FilePath,
				"commit_sha":   fileChange.CommitSHA,
				"chunk_index":  fmt.Sprintf("%d", i),
				"total_chunks": fmt.Sprintf("%d", len(chunks)),
				"file_ext":     filepath.Ext(fileChange.FilePath),
			},
		}
	}

	logger.Debug("Split %s into %d chunks", fileChange.FilePath, len(documents))
	return documents, nil
}

// splitIntoChunks splits text into chunks with overlap
func (p *DocumentProcessor) splitIntoChunks(text string, maxSize, overlap int) []string {
	var chunks []string
	start := 0
	textLen := len(text)

	for start < textLen {
		end := start + maxSize
		if end > textLen {
			end = textLen
		}

		// Try to break at sentence boundary
		if end < textLen {
			// Look for sentence endings
			lastPeriod := strings.LastIndexAny(text[start:end], ".!?\n")
			if lastPeriod > maxSize/2 { // Only break if we're past halfway
				end = start + lastPeriod + 1
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// Move start position with overlap
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// ValidateDocument checks if document should be processed
func (p *DocumentProcessor) ValidateDocument(fileChange *models.FileChange, allowedExtensions []string, excludePatterns []string) bool {
	// Check file extension
	ext := filepath.Ext(fileChange.FilePath)
	if len(allowedExtensions) > 0 {
		found := false
		for _, allowed := range allowedExtensions {
			if ext == allowed {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range excludePatterns {
		if strings.Contains(fileChange.FilePath, pattern) {
			return false
		}
	}

	// Check if file is deleted
	if fileChange.ChangeType == "deleted" || fileChange.ChangeType == "removed" {
		return false
	}

	return true
}

// CleanContent cleans and normalizes document content
func (p *DocumentProcessor) CleanContent(content string) string {
	// Remove excessive whitespace
	lines := strings.Split(content, "\n")
	var cleaned []string

	for _, line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Remove control characters except newlines and tabs
		line = strings.Map(func(r rune) rune {
			if r == '\t' || unicode.IsPrint(r) {
				return r
			}
			return -1
		}, line)

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// HTTP Handlers
type ChunkRequest struct {
	FileChange   *models.FileChange `json:"file_change"`
	MaxChunkSize int                `json:"max_chunk_size,omitempty"`
	ChunkOverlap int                `json:"chunk_overlap,omitempty"`
}

type ChunkResponse struct {
	Documents []*models.Document `json:"documents"`
	Count     int                `json:"count"`
}

func (p *DocumentProcessor) handleChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChunkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	maxSize := req.MaxChunkSize
	if maxSize == 0 {
		maxSize = p.maxChunkSize
	}

	overlap := req.ChunkOverlap
	if overlap == 0 {
		overlap = p.chunkOverlap
	}

	documents, err := p.ChunkDocument(r.Context(), req.FileChange, maxSize, overlap)
	if err != nil {
		logger.Error("Failed to chunk document: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ChunkResponse{
		Documents: documents,
		Count:     len(documents),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (p *DocumentProcessor) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "healthy",
		"max_chunk_size": fmt.Sprintf("%d", p.maxChunkSize),
		"chunk_overlap":  fmt.Sprintf("%d", p.chunkOverlap),
	})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "document-processor"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Document Processor Service on port %d", cfg.Services.DocumentProcessorPort)

	// Create document processor
	service := NewDocumentProcessor(cfg.Processing.MaxChunkSize, cfg.Processing.ChunkOverlap)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/chunk", service.handleChunk)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.DocumentProcessorPort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down document processor...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Document Processor Service listening on port %d", cfg.Services.DocumentProcessorPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
