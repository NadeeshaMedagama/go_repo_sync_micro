package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Azure OpenAI
	AzureOpenAI AzureOpenAIConfig

	// GitHub
	GitHub GitHubConfig

	// Pinecone
	Pinecone PineconeConfig

	// Processing
	Processing ProcessingConfig

	// Database
	Database DatabaseConfig

	// Logging
	Logging LoggingConfig

	// Notifications
	Notifications NotificationsConfig

	// Scheduler
	Scheduler SchedulerConfig

	// Services
	Services ServicesConfig
}

type AzureOpenAIConfig struct {
	APIKey               string
	Endpoint             string
	EmbeddingsVersion    string
	EmbeddingsDeployment string
	APIVersion           string
	ChatDeployment       string
}

type GitHubConfig struct {
	Token         string
	Organization  string
	FilterKeyword string
}

type PineconeConfig struct {
	APIKey        string
	IndexName     string
	Dimension     int
	Cloud         string
	Region        string
	UseNamespaces bool
}

type ProcessingConfig struct {
	AllowedExtensions       []string
	ExcludePatterns         []string
	MaxWorkers              int
	RateLimitRequestsPerMin int
	EmbeddingBatchSize      int
	MaxChunkSize            int
	ChunkOverlap            int
}

type DatabaseConfig struct {
	MetadataDBPath string
}

type LoggingConfig struct {
	Level    string
	FilePath string
}

type NotificationsConfig struct {
	SlackWebhookURL string
}

type SchedulerConfig struct {
	Time     string
	Timezone string
}

type ServicesConfig struct {
	OrchestratorPort        int
	GitHubServicePort       int
	DocumentProcessorPort   int
	EmbeddingServicePort    int
	VectorStoragePort       int
	NotificationServicePort int
	MetadataServicePort     int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (optional)
	_ = godotenv.Load()

	config := &Config{
		AzureOpenAI: AzureOpenAIConfig{
			APIKey:               getEnv("AZURE_OPENAI_API_KEY", ""),
			Endpoint:             getEnv("AZURE_OPENAI_ENDPOINT", ""),
			EmbeddingsVersion:    getEnv("AZURE_OPENAI_EMBEDDINGS_VERSION", "2023-05-15"),
			EmbeddingsDeployment: getEnv("AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT", "text-embedding-ada-002"),
			APIVersion:           getEnv("AZURE_OPENAI_API_VERSION", "2023-05-15"),
			ChatDeployment:       getEnv("AZURE_OPENAI_CHAT_DEPLOYMENT", "gpt-35-turbo"),
		},
		GitHub: GitHubConfig{
			Token:         getEnv("GH_TOKEN", ""),
			Organization:  getEnv("GH_ORGANIZATION", ""),
			FilterKeyword: getEnv("GH_FILTER_KEYWORD", ""),
		},
		Pinecone: PineconeConfig{
			APIKey:        getEnv("PINECONE_API_KEY", ""),
			IndexName:     getEnv("PINECONE_INDEX_NAME", "reposync-index"),
			Dimension:     getEnvInt("PINECONE_DIMENSION", 1536),
			Cloud:         getEnv("PINECONE_CLOUD", "aws"),
			Region:        getEnv("PINECONE_REGION", "us-east-1"),
			UseNamespaces: getEnvBool("PINECONE_USE_NAMESPACES", true),
		},
		Processing: ProcessingConfig{
			AllowedExtensions:       parseCSV(getEnv("ALLOWED_FILE_EXTENSIONS", ".md,.rst,.txt,.yaml,.yml,.json")),
			ExcludePatterns:         parseCSV(getEnv("EXCLUDE_PATTERNS", "node_modules,__pycache__,.git,dist,build")),
			MaxWorkers:              getEnvInt("MAX_WORKERS", 5),
			RateLimitRequestsPerMin: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			EmbeddingBatchSize:      getEnvInt("EMBEDDING_BATCH_SIZE", 100),
			MaxChunkSize:            getEnvInt("MAX_CHUNK_SIZE", 1000),
			ChunkOverlap:            getEnvInt("CHUNK_OVERLAP", 200),
		},
		Database: DatabaseConfig{
			MetadataDBPath: getEnv("METADATA_DB_PATH", "./data/metadata.db"),
		},
		Logging: LoggingConfig{
			Level:    getEnv("LOG_LEVEL", "INFO"),
			FilePath: getEnv("LOG_FILE_PATH", "./logs/reposync.log"),
		},
		Notifications: NotificationsConfig{
			SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
		},
		Scheduler: SchedulerConfig{
			Time:     getEnv("SCHEDULE_TIME", "08:00"),
			Timezone: getEnv("SCHEDULE_TIMEZONE", "UTC"),
		},
		Services: ServicesConfig{
			OrchestratorPort:        getEnvInt("ORCHESTRATOR_PORT", 9090),
			GitHubServicePort:       getEnvInt("GITHUB_SERVICE_PORT", 9081),
			DocumentProcessorPort:   getEnvInt("DOCUMENT_PROCESSOR_PORT", 9082),
			EmbeddingServicePort:    getEnvInt("EMBEDDING_SERVICE_PORT", 9083),
			VectorStoragePort:       getEnvInt("VECTOR_STORAGE_PORT", 9084),
			NotificationServicePort: getEnvInt("NOTIFICATION_SERVICE_PORT", 9085),
			MetadataServicePort:     getEnvInt("METADATA_SERVICE_PORT", 9086),
		},
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.AzureOpenAI.APIKey == "" {
		return fmt.Errorf("AZURE_OPENAI_API_KEY is required")
	}
	if c.AzureOpenAI.Endpoint == "" {
		return fmt.Errorf("AZURE_OPENAI_ENDPOINT is required")
	}
	if c.GitHub.Token == "" {
		return fmt.Errorf("GH_TOKEN is required")
	}
	if c.GitHub.Organization == "" {
		return fmt.Errorf("GH_ORGANIZATION is required")
	}
	if c.Pinecone.APIKey == "" {
		return fmt.Errorf("PINECONE_API_KEY is required")
	}
	if c.Pinecone.IndexName == "" {
		return fmt.Errorf("PINECONE_INDEX_NAME is required")
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
