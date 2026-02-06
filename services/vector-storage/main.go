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

	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/errors"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// VectorStorageService implements interfaces.VectorStore
type VectorStorageService struct {
	client    *pinecone.Client
	indexName string
	dimension int
}

// NewVectorStorageService creates a new vector storage service
func NewVectorStorageService(apiKey, indexName string, dimension int) (*VectorStorageService, error) {
	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	return &VectorStorageService{
		client:    client,
		indexName: indexName,
		dimension: dimension,
	}, nil
}

// UpsertVectors inserts or updates vectors
func (s *VectorStorageService) UpsertVectors(ctx context.Context, embeddings []*models.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	// Determine namespace
	namespace := ""
	if len(embeddings) > 0 && embeddings[0].Namespace != "" {
		namespace = embeddings[0].Namespace
	}

	// Convert to Pinecone vectors
	vectors := make([]*pinecone.Vector, len(embeddings))
	for i, emb := range embeddings {
		// Convert metadata to structpb.Struct
		metadataMap := make(map[string]interface{})
		for k, v := range emb.Metadata {
			metadataMap[k] = v
		}
		metadata, err := structpb.NewStruct(metadataMap)
		if err != nil {
			return errors.Internal("failed to convert metadata", err)
		}

		vectors[i] = &pinecone.Vector{
			Id:       emb.ID,
			Values:   emb.Vector,
			Metadata: metadata,
		}
	}

	// Get index connection
	idx, err := s.client.DescribeIndex(ctx, s.indexName)
	if err != nil {
		return errors.External("Pinecone", "failed to describe index", err)
	}

	idxConnection, err := s.client.Index(pinecone.NewIndexConnParams{Host: idx.Host, Namespace: namespace})
	if err != nil {
		return errors.External("Pinecone", "failed to connect to index", err)
	}

	// Upsert vectors (namespace is set on the connection)
	_, err = idxConnection.UpsertVectors(ctx, vectors)
	if err != nil {
		return errors.External("Pinecone", "failed to upsert vectors", err)
	}

	logger.Info("Upserted %d vectors to namespace '%s'", len(vectors), namespace)
	return nil
}

// DeleteVectors removes vectors by IDs
func (s *VectorStorageService) DeleteVectors(ctx context.Context, ids []string, namespace string) error {
	if len(ids) == 0 {
		return nil
	}

	idx, err := s.client.DescribeIndex(ctx, s.indexName)
	if err != nil {
		return errors.External("Pinecone", "failed to describe index", err)
	}

	idxConnection, err := s.client.Index(pinecone.NewIndexConnParams{Host: idx.Host, Namespace: namespace})
	if err != nil {
		return errors.External("Pinecone", "failed to connect to index", err)
	}

	err = idxConnection.DeleteVectorsById(ctx, ids)
	if err != nil {
		return errors.External("Pinecone", "failed to delete vectors", err)
	}

	logger.Info("Deleted %d vectors from namespace '%s'", len(ids), namespace)
	return nil
}

// QueryVectors searches for similar vectors
func (s *VectorStorageService) QueryVectors(ctx context.Context, vector []float32, topK int, namespace string) ([]*models.Embedding, error) {
	idx, err := s.client.DescribeIndex(ctx, s.indexName)
	if err != nil {
		return nil, errors.External("Pinecone", "failed to describe index", err)
	}

	idxConnection, err := s.client.Index(pinecone.NewIndexConnParams{Host: idx.Host, Namespace: namespace})
	if err != nil {
		return nil, errors.External("Pinecone", "failed to connect to index", err)
	}

	topK32 := uint32(topK)

	queryResp, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          vector,
		TopK:            topK32,
		IncludeMetadata: true,
		IncludeValues:   true,
	})

	if err != nil {
		return nil, errors.External("Pinecone", "failed to query vectors", err)
	}

	// Convert results
	results := make([]*models.Embedding, len(queryResp.Matches))
	for i, match := range queryResp.Matches {
		metadata := make(map[string]string)
		if match.Vector != nil && match.Vector.Metadata != nil {
			for k, v := range match.Vector.Metadata.AsMap() {
				if strVal, ok := v.(string); ok {
					metadata[k] = strVal
				} else {
					metadata[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		var id string
		var values []float32
		if match.Vector != nil {
			id = match.Vector.Id
			values = match.Vector.Values
		}

		results[i] = &models.Embedding{
			ID:        id,
			Vector:    values,
			Metadata:  metadata,
			Namespace: namespace,
		}
	}

	return results, nil
}

// DescribeIndex gets index statistics
func (s *VectorStorageService) DescribeIndex(ctx context.Context) (map[string]interface{}, error) {
	idx, err := s.client.DescribeIndex(ctx, s.indexName)
	if err != nil {
		return nil, errors.External("Pinecone", "failed to describe index", err)
	}

	stats := map[string]interface{}{
		"name":      idx.Name,
		"dimension": idx.Dimension,
		"metric":    idx.Metric,
		"host":      idx.Host,
		"status":    idx.Status.State,
	}

	return stats, nil
}

// Health checks the connection health
func (s *VectorStorageService) Health(ctx context.Context) error {
	_, err := s.client.DescribeIndex(ctx, s.indexName)
	return err
}

// HTTP Handlers
type UpsertRequest struct {
	Embeddings []*models.Embedding `json:"embeddings"`
}

type QueryRequest struct {
	Vector    []float32 `json:"vector"`
	TopK      int       `json:"top_k"`
	Namespace string    `json:"namespace"`
}

func (s *VectorStorageService) handleUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.UpsertVectors(r.Context(), req.Embeddings); err != nil {
		logger.Error("Failed to upsert vectors: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"upserted": len(req.Embeddings),
	})
}

func (s *VectorStorageService) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.Health(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	stats, err := s.DescribeIndex(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	stats["status"] = "healthy"
	json.NewEncoder(w).Encode(stats)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "vector-storage"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Vector Storage Service on port %d", cfg.Services.VectorStoragePort)

	// Create vector storage service
	service, err := NewVectorStorageService(
		cfg.Pinecone.APIKey,
		cfg.Pinecone.IndexName,
		cfg.Pinecone.Dimension,
	)
	if err != nil {
		logger.Fatal("Failed to create vector storage service: %v", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/upsert", service.handleUpsert)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.VectorStoragePort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down vector storage service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Vector Storage Service listening on port %d", cfg.Services.VectorStoragePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
