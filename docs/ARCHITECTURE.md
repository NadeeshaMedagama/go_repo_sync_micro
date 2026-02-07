# RepoSync Architecture Guide

## System Overview

RepoSync is a distributed microservices system designed to automatically synchronize GitHub repositories with Pinecone vector databases for RAG (Retrieval-Augmented Generation) applications. The system follows SOLID principles and implements a clean, maintainable architecture.

## Architecture Principles

### 1. Microservices Architecture

Each service is:
- **Independent**: Can be deployed and scaled separately
- **Focused**: Has a single, well-defined responsibility
- **Resilient**: Failures in one service don't cascade
- **Observable**: Provides health checks and logging

### 2. SOLID Principles

#### Single Responsibility Principle (SRP)
- Each service handles one aspect of the sync process
- Clear separation between data access, business logic, and presentation

#### Open/Closed Principle (OCP)
- Services are open for extension through interfaces
- Closed for modification via dependency injection

#### Liskov Substitution Principle (LSP)
- All service implementations adhere to defined interfaces
- Services can be swapped without affecting the system

#### Interface Segregation Principle (ISP)
- Small, focused interfaces in `pkg/interfaces/`
- Services depend only on interfaces they use

#### Dependency Inversion Principle (DIP)
- High-level modules depend on abstractions
- Low-level implementations injected via configuration

## Service Architecture

### 1. Orchestrator Service (Port 8080)

**Purpose**: Coordinates all services to execute sync operations

**Responsibilities**:
- Receives sync requests via HTTP API
- Orchestrates multi-step sync workflow
- Manages error handling and retries
- Aggregates results and generates reports

**Flow**:
```
HTTP Request → Discover Repos → Detect Changes → Process Files → 
Generate Embeddings → Store Vectors → Update Metadata → Notify
```

**Dependencies**:
- All other services (via HTTP)
- Configuration service

### 2. GitHub Discovery Service (Port 8081)

**Purpose**: Interfaces with GitHub API

**Responsibilities**:
- List repositories in organization
- Detect file changes using commit comparison
- Fetch file contents
- Track commit SHAs

**Implementation**:
- Uses `go-github` library
- Implements rate limiting
- Caches repository metadata

**Endpoints**:
- `GET /repositories?org=X&keyword=Y` - List repos
- `GET /changes?repo=X&last_commit=Y` - Get changes
- `GET /content?repo=X&path=Y` - Get file content

### 3. Document Processor Service (Port 8082)

**Purpose**: Chunks documents for embedding

**Responsibilities**:
- Validate file types and patterns
- Clean and normalize content
- Split into configurable chunks
- Maintain chunk metadata

**Algorithm**:
```go
1. Clean content (remove control chars, normalize whitespace)
2. Split into chunks (max size with overlap)
3. Break at sentence boundaries
4. Generate chunk IDs (MD5 hash)
5. Add metadata (repo, file path, chunk index)
```

**Configuration**:
- `MAX_CHUNK_SIZE`: Maximum characters per chunk (default: 1000)
- `CHUNK_OVERLAP`: Overlap between chunks (default: 200)

### 4. Embedding Service (Port 8083)

**Purpose**: Generate vector embeddings

**Responsibilities**:
- Call Azure OpenAI Embeddings API
- Batch processing for efficiency
- Handle rate limits and retries
- Return 1536-dimensional vectors

**Implementation**:
- Uses Azure SDK for Go
- Implements exponential backoff
- Supports batch sizes up to 100

**Performance**:
- Typical latency: 500-1000ms per batch
- Throughput: ~1000 embeddings/minute

### 5. Vector Storage Service (Port 8084)

**Purpose**: Manage Pinecone vector database

**Responsibilities**:
- Upsert vectors with metadata
- Delete obsolete vectors
- Query for similar vectors
- Manage namespaces

**Implementation**:
- Uses Pinecone Go SDK
- Supports batch operations
- Implements connection pooling

**Operations**:
- `POST /upsert` - Upsert vectors
- `DELETE /delete` - Delete vectors
- `POST /query` - Query similar vectors
- `GET /describe` - Index statistics

### 6. Metadata Service (Port 8086)

**Purpose**: Track sync state

**Responsibilities**:
- Store last commit SHAs
- Track sync timestamps
- Manage project configurations
- Provide sync history

**Storage**:
- SQLite database (`data/metadata.db`)
- Tables: `sync_metadata`, `projects`

**Schema**:
```sql
CREATE TABLE sync_metadata (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL,
    repository TEXT NOT NULL,
    file_path TEXT NOT NULL,
    last_commit_sha TEXT NOT NULL,
    last_synced_at DATETIME NOT NULL,
    embedding_count INTEGER,
    status TEXT,
    UNIQUE(project_id, repository, file_path)
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    organization TEXT NOT NULL,
    filter_keyword TEXT,
    namespace TEXT NOT NULL,
    enabled BOOLEAN,
    allowed_extensions TEXT,
    exclude_patterns TEXT,
    created_at DATETIME,
    updated_at DATETIME
);
```

### 7. Notification Service (Port 8085)

**Purpose**: Send notifications

**Responsibilities**:
- Format sync results
- Send Slack notifications
- Handle webhook failures

**Slack Message Format**:
```json
{
  "attachments": [{
    "color": "good",
    "title": "✅ RepoSync Update",
    "fields": [
      {"title": "Duration", "value": "15m30s", "short": true},
      {"title": "Repositories", "value": "25", "short": true},
      {"title": "Files Processed", "value": "12/450", "short": true},
      {"title": "Embeddings", "value": "38", "short": true}
    ]
  }]
}
```

## Data Flow

### Incremental Sync Flow

```
1. Orchestrator receives sync request
   ↓
2. Query metadata for last commit SHAs
   ↓
3. GitHub Discovery: List repositories
   ↓
4. For each repository:
   a. Get changed files since last commit
   b. Filter by extensions and patterns
   ↓
5. Document Processor: Chunk each file
   ↓
6. Embedding Service: Generate embeddings (batched)
   ↓
7. Vector Storage: Upsert to Pinecone
   ↓
8. Metadata Service: Update sync state
   ↓
9. Notification Service: Send Slack alert
```

### Full Sync Flow

Same as incremental but:
- Skip step 2 (no last commit SHA)
- Process all files in repositories

## Communication Patterns

### HTTP REST APIs

All inter-service communication uses HTTP REST:

**Advantages**:
- Simple and well-understood
- Easy to debug with curl
- Language-agnostic
- Works with existing infrastructure

**Considerations**:
- Synchronous (blocking)
- No built-in retry logic
- Requires explicit error handling

### Future: Message Queue (Optional)

For high-scale deployments, consider:
- **NATS** or **RabbitMQ** for async processing
- **Dead letter queues** for failed messages
- **Event sourcing** for audit trails

## Scalability

### Horizontal Scaling

Each service can scale independently:

```yaml
# Docker Compose example
services:
  document-processor:
    replicas: 3
  embedding:
    replicas: 2
```

### Vertical Scaling

Adjust resources per service:

```yaml
services:
  embedding:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### Bottlenecks

1. **Azure OpenAI API**: Rate limited (adjust `RATE_LIMIT_REQUESTS_PER_MINUTE`)
2. **Pinecone upserts**: Batch size limits (adjust `EMBEDDING_BATCH_SIZE`)
3. **GitHub API**: 5000 requests/hour (use conditional requests)

## Error Handling

### Retry Strategy

```go
// Exponential backoff with jitter
func retry(operation func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        if isRateLimitError(err) {
            waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
            time.Sleep(waitTime + jitter())
            continue
        }
        
        return err // Non-retriable error
    }
    return fmt.Errorf("max retries exceeded")
}
```

### Error Categories

1. **Transient Errors**: Retry with backoff
   - Network timeouts
   - Rate limits
   - Temporary service unavailability

2. **Permanent Errors**: Fail fast
   - Authentication failures
   - Invalid configuration
   - Malformed requests

3. **Partial Failures**: Continue processing
   - Single file processing errors
   - Non-critical API failures

## Monitoring & Observability

### Health Checks

Each service exposes `/health`:

```json
{
  "status": "healthy",
  "service": "orchestrator",
  "timestamp": "2025-02-03T08:00:00Z"
}
```

### Logging

Structured logging with levels:

```go
logger.Info("Processing file: %s", filePath)
logger.Warning("Rate limit approaching: %d requests remaining", remaining)
logger.Error("Failed to generate embedding: %v", err)
```

Log aggregation:
- Local: `logs/` directory
- Docker: stdout/stderr
- Production: Ship to ELK/Splunk

### Metrics (Future)

Prometheus-compatible metrics:

```
reposync_sync_duration_seconds
reposync_files_processed_total
reposync_embeddings_generated_total
reposync_errors_total{service="embedding",type="rate_limit"}
```

## Security

### Authentication

- **GitHub**: Personal access token (read-only)
- **Azure OpenAI**: API key
- **Pinecone**: API key
- **Services**: No auth (internal network)

### Network Security

```
┌─────────────────────────────┐
│   External Networks         │
│   (GitHub, Azure, Pinecone) │
└──────────────┬──────────────┘
               │ HTTPS/TLS
               ▼
┌─────────────────────────────┐
│   Orchestrator (Gateway)    │
└──────────────┬──────────────┘
               │ HTTP (internal)
               ▼
┌─────────────────────────────┐
│   Internal Services         │
│   (Private network)         │
└─────────────────────────────┘
```

### Best Practices

- ✅ Use environment variables for secrets
- ✅ Rotate credentials regularly
- ✅ Limit GitHub token scopes
- ✅ Use HTTPS for external APIs
- ✅ Scan Docker images (Trivy)
- ✅ Apply least privilege principle

## Deployment

### Local Development

```bash
docker-compose up -d
```

### GitHub Actions

```yaml
- Schedule: Daily at 8:00 AM UTC
- Build: All services
- Start: Sequential service startup
- Sync: Trigger orchestrator API
- Artifacts: Logs and results
```

### Kubernetes (Future)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reposync-orchestrator
spec:
  replicas: 2
  selector:
    matchLabels:
      app: orchestrator
  template:
    spec:
      containers:
      - name: orchestrator
        image: reposync/orchestrator:latest
        ports:
        - containerPort: 8080
```

## Performance

### Typical Sync Times

| Files | Chunks | Embeddings | Duration |
|-------|--------|------------|----------|
| 10    | 30     | 30         | 2-3 min  |
| 50    | 150    | 150        | 8-10 min |
| 100   | 300    | 300        | 15-20 min|
| 500   | 1500   | 1500       | 60-90 min|

### Optimization Tips

1. **Incremental sync**: Only process changed files
2. **Batch embeddings**: Use max batch size (100)
3. **Concurrent processing**: Adjust `MAX_WORKERS`
4. **Cache metadata**: Reduce database queries
5. **Rate limiting**: Stay within API quotas

## Future Enhancements

1. **Caching Layer**: Redis for metadata and embeddings
2. **Event Sourcing**: Audit trail of all changes
3. **Web UI**: Management interface
4. **Multi-tenancy**: Support multiple organizations
5. **Plugins**: Custom document processors
6. **Webhooks**: Real-time updates on Git push

---

**Last Updated**: February 3, 2026
