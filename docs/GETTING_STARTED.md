# Getting Started with RepoSync

Welcome to RepoSync! This guide will help you get the system up and running in minutes.

## Prerequisites Checklist

Before starting, ensure you have:

- [ ] **Go 1.21 or higher** - [Download](https://golang.org/dl/)
- [ ] **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- [ ] **Git** - [Install Git](https://git-scm.com/downloads)
- [ ] **Azure OpenAI account** with API key
- [ ] **GitHub Personal Access Token**
- [ ] **Pinecone account** with API key and index
- [ ] **Slack webhook** (optional)

## Step 1: Clone Repository

```bash
git clone https://github.com/nadeeshame/Go_RepoSync_Micro.git
cd Go_RepoSync_Micro
```

## Step 2: Configure Environment

### Create .env file

```bash
cp .env.example .env
```

### Edit .env with your credentials

```env
# Azure OpenAI
AZURE_OPENAI_API_KEY=your_azure_key_here
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT=text-embedding-ada-002
AZURE_OPENAI_API_VERSION=2023-05-15

# GitHub
GH_TOKEN=ghp_your_token_here
GH_ORGANIZATION=your-organization
GH_FILTER_KEYWORD=docs

# Pinecone
PINECONE_API_KEY=your_pinecone_key_here
PINECONE_INDEX_NAME=reposync-index
PINECONE_DIMENSION=1536

# Slack (Optional)
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Get Credentials

#### Azure OpenAI
1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to your Azure OpenAI resource
3. Click "Keys and Endpoint"
4. Copy Key 1 and Endpoint

#### GitHub Token
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token (classic)
3. Select scopes: `repo`, `read:org`
4. Copy token immediately

#### Pinecone
1. Go to [Pinecone Console](https://app.pinecone.io)
2. Click "API Keys" → Copy your API key
3. Create index:
   ```bash
   curl -X POST "https://api.pinecone.io/indexes" \
     -H "Api-Key: YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"name": "reposync-index", "dimension": 1536, "metric": "cosine"}'
   ```

## Step 3: Choose Deployment Method

### Option A: Quick Start Script (Recommended)

```bash
./quick-start.sh
```

Follow the interactive prompts.

### Option B: Docker Compose (Manual)

```bash
# Download dependencies
go mod download

# Create directories
mkdir -p data logs

# Start all services - AUTO-SYNC HAPPENS AUTOMATICALLY! 🚀
docker-compose up

# What happens:
# 1. All services start and become healthy
# 2. Auto-sync service triggers automatically
# 3. Retrieves changes from GitHub
# 4. Generates embeddings
# 5. Stores in Pinecone
# 6. Sends notification

# View sync progress
docker-compose logs -f auto-sync

# Check health
curl http://localhost:8080/health
```

**Configuration options** (`.env`):
```env
SYNC_INCREMENTAL=true    # Fast incremental sync (only changed files)
SYNC_MODE=incremental    # Or "full" for complete resync
PROJECT_ID=default       # Your project ID
```

### Option C: Local Binaries (Development)

```bash
# Build all services
make build

# Start each service in separate terminals
./bin/metadata          # Terminal 1
./bin/github-discovery  # Terminal 2
./bin/document-processor # Terminal 3
./bin/embedding         # Terminal 4
./bin/vector-storage    # Terminal 5
./bin/notification      # Terminal 6
./bin/orchestrator      # Terminal 7
```

## Step 4: Verify Installation

### Check Service Health

```bash
# All services
make health

# Or individually
curl http://localhost:8080/health  # Orchestrator
curl http://localhost:8081/health  # GitHub Discovery
curl http://localhost:8082/health  # Document Processor
curl http://localhost:8083/health  # Embedding
curl http://localhost:8084/health  # Vector Storage
curl http://localhost:8085/health  # Notification
curl http://localhost:8086/health  # Metadata
```

Expected response:
```json
{"status":"healthy"}
```

### View Logs

```bash
# Docker Compose
docker-compose logs -f orchestrator

# Local binaries
tail -f logs/orchestrator.log
```

## Step 5: Trigger Your First Sync

### Incremental Sync (Recommended)

```bash
curl -X POST "http://localhost:8080/sync?incremental=true" | jq '.'
```

### Full Sync

```bash
curl -X POST "http://localhost:8080/sync?incremental=false" | jq '.'
```

### Expected Response

```json
{
  "project_id": "default",
  "start_time": "2025-02-03T08:00:00Z",
  "end_time": "2025-02-03T08:05:30Z",
  "duration": "5m30s",
  "repositories_scanned": 5,
  "files_discovered": 50,
  "files_changed": 10,
  "files_processed": 10,
  "chunks_created": 25,
  "embeddings_generated": 25,
  "vectors_upserted": 25,
  "success": true,
  "errors": [],
  "warnings": []
}
```

## Step 6: Set Up GitHub Actions (Optional)

For automated daily syncs:

### 1. Add GitHub Secrets

Go to your repository: **Settings → Secrets and variables → Actions**

Add these secrets:
- `AZURE_OPENAI_API_KEY`
- `AZURE_OPENAI_ENDPOINT`
- `AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT`
- `GH_TOKEN`
- `GH_ORGANIZATION`
- `GH_FILTER_KEYWORD`
- `PINECONE_API_KEY`
- `PINECONE_INDEX_NAME`
- `SLACK_WEBHOOK_URL` (optional)

See [GitHub Secrets Setup Guide](GITHUB_SECRETS_SETUP.md) for details.

### 2. Enable GitHub Actions

1. Go to repository **Settings → Actions → General**
2. Select "Allow all actions and reusable workflows"
3. Click "Save"

### 3. Verify Workflow

1. Go to **Actions** tab
2. Select "Scheduled Sync"
3. Click "Run workflow" → "Run workflow"
4. Wait for completion (2-10 minutes)
5. Check logs and artifacts

The workflow will now run **automatically every day at 8:00 AM UTC**.

## Common Commands

### Docker Compose

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart a service
docker-compose restart orchestrator

# Rebuild images
docker-compose build

# Scale a service
docker-compose up -d --scale document-processor=3
```

### Make Commands

```bash
# Build all services
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Check health
make health

# Trigger sync
make sync

# Full sync
make sync-full

# Clean build artifacts
make clean
```

### Manual API Calls

```bash
# Trigger sync with custom project
curl -X POST "http://localhost:8080/sync?project_id=my-project&incremental=true"

# Check orchestrator health
curl http://localhost:8080/health

# View specific service health
curl http://localhost:8083/health  # Embedding service
```

## Troubleshooting

### Services won't start

```bash
# Check logs
docker-compose logs orchestrator

# Verify .env file
cat .env | grep -v "^#"

# Restart services
docker-compose restart
```

### Sync fails

```bash
# Check orchestrator logs
docker-compose logs orchestrator

# Verify credentials
# - GitHub token has correct permissions
# - Azure OpenAI key is valid
# - Pinecone index exists

# Test individual services
curl http://localhost:8081/health  # GitHub
curl http://localhost:8083/health  # Embedding
curl http://localhost:8084/health  # Vector Storage
```

### Port already in use

```bash
# Change ports in .env
ORCHESTRATOR_PORT=8090
# (or other ports)

# Restart services
docker-compose down && docker-compose up -d
```

### Out of memory

```bash
# Reduce workers
MAX_WORKERS=3  # in .env

# Reduce batch size
EMBEDDING_BATCH_SIZE=50  # in .env
```

## Next Steps

### Customize Configuration

Edit `.env` to adjust:
- File extensions to process
- Chunk size and overlap
- Worker count
- Rate limits

See [Configuration Guide](docs/CONFIGURATION.md) for all options.

### Explore Features

- **Multi-project support**: Manage multiple knowledge bases
- **Incremental sync**: Fast updates (15x faster)
- **Slack notifications**: Get sync alerts
- **Health monitoring**: Track service status

### Learn Architecture

- Read [ARCHITECTURE.md](ARCHITECTURE.md) for system design
- Review [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for complete overview
- Explore individual service implementations

### Contribute

- Follow [CONTRIBUTING.md](CONTRIBUTING.md) guidelines
- Review SOLID principles in `pkg/interfaces/`
- Submit pull requests with tests
- Update documentation

## Support

- **Documentation**: See `docs/` directory
- **Issues**: [GitHub Issues](https://github.com/nadeeshame/Go_RepoSync_Micro/issues)
- **Discussions**: [GitHub Discussions](https://github.com/nadeeshame/Go_RepoSync_Micro/discussions)

## Quick Reference Card

```bash
# Start system
docker-compose up -d

# Trigger sync
make sync

# Check health
make health

# View logs
docker-compose logs -f orchestrator

# Stop system
docker-compose down
```

---

**Congratulations!** 🎉 Your RepoSync system is now ready to automatically sync your GitHub repositories to Pinecone for RAG applications!

For questions or help, check the documentation or open an issue on GitHub.
