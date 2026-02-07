package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/errors"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
)

// EmbeddingService implements interfaces.EmbeddingService
type EmbeddingService struct {
	client     *azopenai.Client
	deployment string
	dimension  int
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(endpoint, apiKey, deployment string) (*EmbeddingService, error) {
	keyCredential := azcore.NewKeyCredential(apiKey)
	client, err := azopenai.NewClientWithKeyCredential(endpoint, keyCredential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure OpenAI client: %w", err)
	}

	return &EmbeddingService{
		client:     client,
		deployment: deployment,
		dimension:  1536, // text-embedding-ada-002 dimension
	}, nil
}

// GenerateEmbedding creates a vector embedding for text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.GenerateBatchEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.Internal("no embeddings generated", nil)
	}
	return embeddings[0], nil
}

// GenerateBatchEmbeddings creates embeddings for multiple texts
func (s *EmbeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	resp, err := s.client.GetEmbeddings(ctx, azopenai.EmbeddingsOptions{
		Input:          texts,
		DeploymentName: &s.deployment,
	}, nil)

	if err != nil {
		return nil, errors.External("Azure OpenAI", "failed to generate embeddings", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, item := range resp.Data {
		embeddings[i] = item.Embedding
	}

	logger.Info("Generated %d embeddings", len(embeddings))
	return embeddings, nil
}

// GetDimension returns the dimension of embeddings
func (s *EmbeddingService) GetDimension() int {
	return s.dimension
}

// HTTP Handlers
type EmbeddingRequest struct {
	Texts []string `json:"texts"`
}

type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Count      int         `json:"count"`
}

func (s *EmbeddingService) handleEmbed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	embeddings, err := s.GenerateBatchEmbeddings(r.Context(), req.Texts)
	if err != nil {
		logger.Error("Failed to generate embeddings: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := EmbeddingResponse{
		Embeddings: embeddings,
		Count:      len(embeddings),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *EmbeddingService) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Test with a simple embedding
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := s.GenerateEmbedding(ctx, "test")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "dimension": fmt.Sprintf("%d", s.dimension)})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate embedding-specific requirements
	if err := cfg.ValidateForEmbedding(); err != nil {
		fmt.Printf("Failed to validate configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "embedding-service"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Embedding Service on port %d", cfg.Services.EmbeddingServicePort)

	// Create embedding service
	service, err := NewEmbeddingService(
		cfg.AzureOpenAI.Endpoint,
		cfg.AzureOpenAI.APIKey,
		cfg.AzureOpenAI.EmbeddingsDeployment,
	)
	if err != nil {
		logger.Fatal("Failed to create embedding service: %v", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/embed", service.handleEmbed)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.EmbeddingServicePort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down embedding service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Embedding Service listening on port %d", cfg.Services.EmbeddingServicePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
