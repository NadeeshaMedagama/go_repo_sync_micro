# RepoSync - Go Microservices Edition

[//]: # ([![CI/CD]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/ci-cd.yml/badge.svg?branch=main&#41;]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/ci-cd.yml&#41;)

[//]: # ([![Docker Build & Test]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/docker-build-test.yml/badge.svg?branch=main&#41;]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/docker-build-test.yml&#41;)

[//]: # ([![CodeQL]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/codeql-analysis.yml/badge.svg?branch=main&#41;]&#40;https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/actions/workflows/codeql-analysis.yml&#41;)

[//]: # ([![Go Report Card]&#40;https://goreportcard.com/badge/github.com/NadeeshaMedagama/Go_RepoSync_Micro&#41;]&#40;https://goreportcard.com/report/github.com/NadeeshaMedagama/Go_RepoSync_Micro&#41;)

[//]: # ([![License: MIT]&#40;https://img.shields.io/badge/License-MIT-yellow.svg&#41;]&#40;https://opensource.org/licenses/MIT&#41;)

A robust, enterprise-grade microservices system built in Go for automatically synchronizing GitHub repositories with Pinecone vector databases for RAG (Retrieval-Augmented Generation) applications. Built with **SOLID principles** and designed for scalability and maintainability.

---

## ⚡ **AUTO-SYNC: ONE COMMAND DOES IT ALL!** ⭐ NEW

```bash
docker-compose up
```

**This automatically:**
- ✅ Starts all 7 microservices
- ✅ Retrieves new changes from GitHub
- ✅ Generates embeddings with Azure OpenAI
- ✅ Stores vectors in Pinecone
- ✅ Sends Slack notification

**No manual triggers. No API calls. Just run and watch!**

📖 **[Complete Auto-Sync Guide →](docs/AUTO_SYNC_GUIDE.md)**

---

## 🌟 Features

- **🏗️ Microservices Architecture**: 7 independent, scalable services
- **🔄 Automated Daily Sync**: GitHub Actions scheduled at 8:00 AM UTC
- **🚀 Auto-Sync with Docker Compose**: Automatically syncs on startup! ⭐ NEW
- **⚡ Incremental Updates**: Process only changed files using commit hash tracking
- **🤖 Azure OpenAI Integration**: Generate embeddings using Azure OpenAI API
- **📊 Pinecone Vector Database**: Efficient vector storage with namespace support
- **🔔 Slack Notifications**: Real-time sync status updates
- **🐳 Docker Support**: Full containerization with Docker Compose
- **🔒 Security First**: CodeQL analysis, Dependabot updates, secrets management
- **📈 SOLID Principles**: Clean, maintainable, and extensible design
- **🚀 CI/CD Pipeline**: Automated testing, building, and deployment

---

## 📋 Architecture

### Microservices Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      GitHub Actions Scheduler                    │
│                    (Daily @ 8:00 AM UTC)                        │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │    Orchestrator      │
                    │    (Port 8080)       │
                    └──────────┬───────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
           ▼                   ▼                   ▼
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │   GitHub    │    │  Document   │    │  Embedding  │
    │  Discovery  │    │  Processor  │    │   Service   │
    │ (Port 8081) │    │ (Port 8082) │    │ (Port 8083) │
    └─────────────┘    └─────────────┘    └─────────────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
           ▼                   ▼                   ▼
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │   Vector    │    │  Metadata   │    │Notification │
    │   Storage   │    │   Service   │    │   Service   │
    │ (Port 8084) │    │ (Port 8086) │    │ (Port 8085) │
    └─────────────┘    └─────────────┘    └─────────────┘
           │                   │                   │
           ▼                   ▼                   ▼
      Pinecone            SQLite DB            Slack
```

### Services

1. **Orchestrator (8080)**: Coordinates all services and manages sync workflow
2. **GitHub Discovery (8081)**: Discovers repositories and detects file changes
3. **Document Processor (8082)**: Chunks documents with configurable size/overlap
4. **Embedding Service (8083)**: Generates embeddings via Azure OpenAI
5. **Vector Storage (8084)**: Manages Pinecone vector database operations
6. **Metadata Service (8086)**: Tracks sync state in SQLite database
7. **Notification Service (8085)**: Sends Slack notifications

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **Docker & Docker Compose** (for containerized deployment)
- **GitHub Personal Access Token**
- **Azure OpenAI API credentials**
- **Pinecone API key and index**
- **Slack Webhook URL** (optional)

### 1. Clone Repository

```bash
git clone https://github.com/NadeeshaMedagama/Go_RepoSync_Micro.git
cd Go_RepoSync_Micro
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your credentials
```

Required configuration:
```env
# Azure OpenAI
AZURE_OPENAI_API_KEY=your_key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT=text-embedding-ada-002

# GitHub
GH_TOKEN=your_github_token
GH_ORGANIZATION=your-org
GH_FILTER_KEYWORD=keyword

# Pinecone
PINECONE_API_KEY=your_pinecone_key
PINECONE_INDEX_NAME=reposync-index
PINECONE_DIMENSION=1536

# Slack (optional)
SLACK_WEBHOOK_URL=your_webhook_url
```

### 3. Run with Docker Compose (Recommended)

**Build first (one time):**
```bash
# Build all services (takes 2-3 minutes)
docker-compose build --parallel

# Or use the build script
./build-docker.sh parallel
```

**Then run:**
```bash
# Start all services - AUTO-SYNC HAPPENS AUTOMATICALLY! 🚀
docker-compose up

# Or run in background
docker-compose up -d

# View sync progress
docker-compose logs -f auto-sync

# Check service health
curl http://localhost:8080/health
```

**What happens automatically:**
1. ✅ All 7 services start up
2. ✅ Health checks pass
3. ✅ Auto-sync triggers automatically
4. ✅ Retrieves new changes from GitHub
5. ✅ Generates embeddings with Azure OpenAI
6. ✅ Stores vectors in Pinecone
7. ✅ Sends Slack notification (if configured)

**See [Auto-Sync Guide](docs/AUTO_SYNC_GUIDE.md) for complete documentation.**

**Docker Build Issues?** See [Docker Build Fix Guide](docs/DOCKER_BUILD_FIX.md)

### 4. Run Locally (Development)

```bash
# Install dependencies
go mod download

# Create directories
mkdir -p data logs

# Build all services
make build

# Start services (in separate terminals)
./bin/metadata
./bin/github-discovery
./bin/document-processor
./bin/embedding
./bin/vector-storage
./bin/notification
./bin/orchestrator

# Trigger sync
curl -X POST "http://localhost:8080/sync?incremental=true"
```

---

## 🤖 GitHub Actions Automation

### Scheduled Daily Sync

The system automatically syncs at **8:00 AM UTC daily** via GitHub Actions.

**Setup:**

1. **Add GitHub Secrets** (Settings → Secrets → Actions):
   ```
   AZURE_OPENAI_API_KEY
   AZURE_OPENAI_ENDPOINT
   AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT
   GH_TOKEN
   GH_ORGANIZATION
   GH_FILTER_KEYWORD
   PINECONE_API_KEY
   PINECONE_INDEX_NAME
   SLACK_WEBHOOK_URL
   ```

2. **Enable GitHub Actions** in repository settings

3. **Manual Trigger** (optional):
   - Go to Actions → Scheduled Sync → Run workflow
   - Configure incremental sync and project ID

### Workflow Files

- **`.github/workflows/scheduled-sync.yml`**: Daily automated sync
- **`.github/workflows/ci-cd.yml`**: Build, test, and Docker image creation
- **`.github/workflows/codeql-analysis.yml`**: Security scanning
- **`.github/workflows/release.yml`**: Automated releases

---

## 📚 Usage

### 🚀 Auto-Sync (Easiest Way!)

Simply run docker-compose and the system automatically syncs:

```bash
# Start and auto-sync
docker-compose up

# Configure sync behavior in .env:
SYNC_INCREMENTAL=true    # Only process changed files (fast)
SYNC_MODE=incremental    # incremental or full
PROJECT_ID=default       # Project identifier
```

**Output:**
```
auto-sync  | 🚀 Starting automatic sync...
auto-sync  | 📊 Configuration:
auto-sync  |    - Mode: incremental
auto-sync  |    - Project: default
auto-sync  | ✅ Sync completed successfully!
auto-sync  | 📈 Summary:
auto-sync  |    repositories_scanned: 25
auto-sync  |    files_processed: 42
auto-sync  |    embeddings_generated: 120
```

**See complete guide:** [Auto-Sync Documentation](docs/AUTO_SYNC_GUIDE.md)

### API Endpoints

#### Orchestrator (Port 8080)

**Trigger Sync**
```bash
POST /sync?project_id=default&incremental=true

# Response
{
  "project_id": "default",
  "start_time": "2025-02-03T08:00:00Z",
  "end_time": "2025-02-03T08:15:30Z",
  "duration": "15m30s",
  "repositories_scanned": 25,
  "files_discovered": 450,
  "files_changed": 12,
  "files_processed": 12,
  "chunks_created": 38,
  "embeddings_generated": 38,
  "vectors_upserted": 38,
  "success": true,
  "errors": [],
  "warnings": []
}
```

**Health Check**
```bash
GET /health
```

### Service Health Checks

```bash
# Check all services
curl http://localhost:8080/health  # Orchestrator
curl http://localhost:8081/health  # GitHub Discovery
curl http://localhost:8082/health  # Document Processor
curl http://localhost:8083/health  # Embedding Service
curl http://localhost:8084/health  # Vector Storage
curl http://localhost:8085/health  # Notification Service
curl http://localhost:8086/health  # Metadata Service
```

### Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `ALLOWED_FILE_EXTENSIONS` | `.md,.rst,.txt,.yaml,.yml,.json` | File types to process |
| `EXCLUDE_PATTERNS` | `node_modules,__pycache__,.git` | Directories to skip |
| `MAX_WORKERS` | `5` | Concurrent processing workers |
| `MAX_CHUNK_SIZE` | `1000` | Maximum chunk size (chars) |
| `CHUNK_OVERLAP` | `200` | Overlap between chunks |
| `EMBEDDING_BATCH_SIZE` | `100` | Batch size for embeddings |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | `60` | API rate limit |

---

## 🏗️ SOLID Principles Implementation

### Single Responsibility Principle
- Each service has one clear purpose
- Separation of concerns across microservices

### Open/Closed Principle
- Services are open for extension via interfaces
- Closed for modification through abstraction

### Liskov Substitution Principle
- All implementations follow interface contracts
- Services can be swapped without breaking system

### Interface Segregation Principle
- Small, focused interfaces in `pkg/interfaces/`
- Services depend only on what they use

### Dependency Inversion Principle
- High-level orchestrator depends on abstractions
- Services are injected via configuration

---

## 🐳 Docker Deployment

### Build Images

```bash
# Build all services
docker-compose build

# Build specific service
docker build -t reposync-orchestrator -f services/orchestrator/Dockerfile .
```

### Production Deployment

```bash
# Run in production mode
docker-compose -f docker-compose.yml up -d

# Scale services
docker-compose up -d --scale document-processor=3

# Update services
docker-compose pull
docker-compose up -d
```

---

## 🔒 Security

### Best Practices

- ✅ Never commit `.env` files
- ✅ Use GitHub Secrets for CI/CD
- ✅ Regular dependency updates via Dependabot
- ✅ CodeQL security scanning on every push
- ✅ Minimal Docker images (Alpine Linux)
- ✅ Read-only GitHub tokens when possible
- ✅ HTTPS/TLS for all external connections

### Security Scanning

```bash
# CodeQL runs automatically on push
# Manual security audit:
go list -json -m all | nancy sleuth
```

---

## 📊 Monitoring & Logs

### View Logs

```bash
# Docker Compose
docker-compose logs -f orchestrator
docker-compose logs --tail=100 embedding

# Local deployment
tail -f logs/orchestrator.log
tail -f logs/reposync.log
```

### Log Levels

Set via `LOG_LEVEL` environment variable:
- `DEBUG`: Detailed debugging information
- `INFO`: General information (default)
- `WARNING`: Warning messages
- `ERROR`: Error messages only

---

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific service tests
go test ./services/orchestrator/... -v

# Docker integration tests
make docker-test

# Docker build and test
make docker-test-build
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# View test results
docker-compose -f docker-compose.test.yml logs test-runner
```

See **[Docker Build & Test Guide](docs/DOCKER_BUILD_TEST.md)** for comprehensive testing documentation.

---

## 🛠️ Development

### Project Structure

```
reposync/
├── .github/
│   ├── workflows/          # GitHub Actions workflows
│   └── dependabot.yml      # Dependency updates
├── cmd/                    # Command-line tools
├── docs/                   # Documentation
├── pkg/
│   ├── config/            # Configuration management
│   ├── errors/            # Custom error types
│   ├── interfaces/        # SOLID interfaces
│   ├── logger/            # Logging utilities
│   └── models/            # Data models
├── services/
│   ├── orchestrator/      # Main coordinator
│   ├── metadata/          # Metadata tracking
│   ├── github-discovery/  # GitHub integration
│   ├── document-processor/# Document chunking
│   ├── embedding/         # Azure OpenAI embeddings
│   ├── vector-storage/    # Pinecone integration
│   └── notification/      # Slack notifications
├── docker-compose.yml     # Local development
├── .env.example          # Environment template
├── go.mod                # Go dependencies
├── Makefile              # Build automation
└── README.md             # This file
```

### Building

```bash
# Build all services
make build

# Build specific service
go build -o bin/orchestrator ./services/orchestrator

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/orchestrator-linux ./services/orchestrator
```

---

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Follow SOLID principles** and existing patterns
4. **Write tests** for new functionality
5. **Update documentation** as needed
6. **Commit changes**: `git commit -m 'feat: add amazing feature'`
7. **Push to branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Code Standards

- Follow Go best practices
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Maintain test coverage above 70%

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙋 Support

### Issues & Questions

- **Bug Reports**: [GitHub Issues](https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/NadeeshaMedagama/Go_RepoSync_Micro/discussions)
- **Documentation**: See `docs/` directory

### Troubleshooting

**Service won't start?**
```bash
# Check logs
docker-compose logs <service-name>

# Verify configuration
cat .env | grep -v "^#"

# Test connectivity
curl http://localhost:8080/health
```

**Sync failing?**
- Check GitHub token permissions
- Verify Azure OpenAI quota
- Ensure Pinecone index exists
- Review logs for specific errors

---

## 🗺️ Roadmap

- [ ] Kubernetes deployment manifests
- [ ] Prometheus metrics integration
- [ ] Grafana dashboards
- [ ] Multi-tenant support
- [ ] Web UI for management
- [ ] Support for additional vector databases
- [ ] Webhook support for real-time updates
- [ ] Advanced filtering and transformation rules

---

## 📚 Additional Documentation

- **[Architecture Guide](docs/ARCHITECTURE.md)**: Detailed system design
- **[API Documentation](docs/API.md)**: Complete API reference
- **[Deployment Guide](docs/DEPLOYMENT.md)**: Production deployment
- **[Configuration Guide](docs/CONFIGURATION.md)**: All config options
- **[Troubleshooting](docs/TROUBLESHOOTING.md)**: Common issues

---

**Built with ❤️ using Go, SOLID principles, and microservices architecture**

*For automated knowledge management and RAG applications*
