# Auto-Sync with Docker Compose

This guide explains how to use the automatic sync feature when running the project with Docker Compose.

## Overview

When you run `docker-compose up`, the system will:

1. ✅ Start all 7 microservices
2. ✅ Wait for all services to be healthy
3. ✅ Automatically trigger a sync operation
4. ✅ Retrieve new changes from GitHub repositories
5. ✅ Process documents and generate embeddings
6. ✅ Store embeddings in Pinecone index
7. ✅ Send Slack notification (if configured)

## Quick Start

### 1. Configure Environment

Edit your `.env` file with your credentials:

```bash
# Required: Azure OpenAI
AZURE_OPENAI_API_KEY=your_key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT=text-embedding-ada-002

# Required: GitHub
GH_TOKEN=your_github_token
GH_ORGANIZATION=your-org
GH_FILTER_KEYWORD=keyword

# Required: Pinecone
PINECONE_API_KEY=your_pinecone_key
PINECONE_INDEX_NAME=your-index-name

# Optional: Slack notifications
SLACK_WEBHOOK_URL=your_webhook_url

# Auto-sync configuration (optional - defaults shown)
SYNC_INCREMENTAL=true
SYNC_MODE=incremental
PROJECT_ID=default
```

### 2. Run Docker Compose

```bash
# Start all services and trigger automatic sync
docker-compose up

# Or run in detached mode
docker-compose up -d

# View logs in real-time
docker-compose logs -f auto-sync
```

### 3. Watch the Sync

You'll see output like this:

```
auto-sync  | ⏳ Waiting 10 seconds for all services to stabilize...
auto-sync  | 🚀 Starting automatic sync...
auto-sync  | 📊 Configuration:
auto-sync  |    - Mode: incremental
auto-sync  |    - Project: default
auto-sync  | 
auto-sync  | 📋 Sync Response:
auto-sync  | {"project_id":"default","repositories_scanned":25,...}
auto-sync  | 
auto-sync  | ✅ Sync completed successfully!
auto-sync  | 📈 Summary:
auto-sync  |    repositories_scanned: 25
auto-sync  |    files_processed: 42
auto-sync  |    embeddings_generated: 120
```

## Configuration Options

### Sync Modes

Control the sync behavior with environment variables:

#### Incremental Sync (Default - Recommended)
```bash
SYNC_INCREMENTAL=true
SYNC_MODE=incremental
```
- Processes only changed files since last sync
- Fast (typically 2-10 minutes)
- Lower API costs
- **Recommended for daily use**

#### Full Sync
```bash
SYNC_INCREMENTAL=false
SYNC_MODE=full
```
- Processes all files in repositories
- Slower (can take 30-90 minutes)
- Higher API costs
- Use for initial setup or complete refresh

### Project ID

Manage different projects:

```bash
# Default project
PROJECT_ID=default

# Custom project
PROJECT_ID=my-project
```

## Usage Examples

### Example 1: Daily Incremental Sync

**Scenario**: Automatically sync new changes every day

**Configuration** (`.env`):
```env
SYNC_INCREMENTAL=true
SYNC_MODE=incremental
PROJECT_ID=default
```

**Command**:
```bash
docker-compose up
```

**Expected behavior**: Processes only new/changed files, completes in 2-10 minutes

### Example 2: Full Initial Setup

**Scenario**: First time setup or complete rebuild

**Configuration** (`.env`):
```env
SYNC_INCREMENTAL=false
SYNC_MODE=full
PROJECT_ID=default
```

**Command**:
```bash
docker-compose up
```

**Expected behavior**: Processes all files, takes 30-90 minutes depending on repository size

### Example 3: Multiple Projects

**Scenario**: Sync different organizations/repositories

**Project 1** (`.env`):
```env
PROJECT_ID=project-a
GH_ORGANIZATION=org-a
GH_FILTER_KEYWORD=docs
PINECONE_INDEX_NAME=project-a-index
```

**Project 2** (separate `.env` or environment override):
```bash
PROJECT_ID=project-b \
GH_ORGANIZATION=org-b \
GH_FILTER_KEYWORD=api \
PINECONE_INDEX_NAME=project-b-index \
docker-compose up
```

### Example 4: Manual Trigger After Services Start

**Scenario**: Start services but don't auto-sync initially

**Option 1: Skip auto-sync service**
```bash
docker-compose up orchestrator metadata github-discovery document-processor embedding vector-storage notification
```

**Option 2: Manually trigger sync later**
```bash
# Start services
docker-compose up -d

# Trigger sync manually
curl -X POST "http://localhost:8080/sync?incremental=true"

# Or trigger auto-sync service
docker-compose up auto-sync
```

## Viewing Results

### Check Sync Status

```bash
# View auto-sync logs
docker-compose logs auto-sync

# View orchestrator logs
docker-compose logs orchestrator

# View all logs
docker-compose logs
```

### Check Pinecone Index

Verify embeddings were created:

```bash
# Check index stats
curl -X GET "https://controller.YOUR_ENV.pinecone.io/databases/YOUR_INDEX/describe" \
  -H "Api-Key: YOUR_API_KEY"
```

Or use Pinecone Console:
1. Go to https://app.pinecone.io
2. Select your index
3. View vector count and statistics

### Check Slack Notifications

If configured, you'll receive a Slack message with sync results:

```
✅ RepoSync Update

Duration: 5m30s
Repositories: 25
Files Processed: 42/450
Embeddings Generated: 120
```

## Troubleshooting

### Sync Fails Immediately

**Issue**: auto-sync exits with error

**Check**:
```bash
# View auto-sync logs
docker-compose logs auto-sync

# Check orchestrator health
curl http://localhost:8080/health

# Check all service health
docker-compose ps
```

**Common causes**:
- Services not healthy yet (wait longer)
- Invalid credentials in .env
- Network connectivity issues

**Solution**:
```bash
# Restart with fresh logs
docker-compose down -v
docker-compose up

# Or increase wait time in docker-compose.yml
# Change: sleep 10 → sleep 30
```

### Services Won't Start

**Issue**: Services fail health checks

**Check**:
```bash
# View individual service logs
docker-compose logs metadata
docker-compose logs embedding
docker-compose logs vector-storage

# Check environment variables
docker-compose config
```

**Common causes**:
- Missing or invalid credentials
- Port conflicts
- Insufficient resources

**Solution**:
```bash
# Verify .env file
cat .env | grep -v "^#"

# Check ports are available
netstat -tlnp | grep -E "8080|8081|8082|8083|8084|8085|8086"

# Increase Docker resources (Docker Desktop → Settings → Resources)
```

### Sync Takes Too Long

**Issue**: Sync runs for hours

**Possible causes**:
- Full sync on large repositories
- Rate limiting from APIs
- Network latency

**Solutions**:
```bash
# Use incremental sync
SYNC_INCREMENTAL=true docker-compose up

# Reduce batch sizes in .env
EMBEDDING_BATCH_SIZE=50
MAX_WORKERS=3

# Filter repositories more specifically
GH_FILTER_KEYWORD=specific-keyword
```

### API Rate Limits

**Issue**: `429 Too Many Requests` errors

**GitHub API** (5000 requests/hour):
```bash
# Check rate limit
curl -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/rate_limit
```

**Azure OpenAI** (varies by tier):
```bash
# Reduce request rate in .env
RATE_LIMIT_REQUESTS_PER_MINUTE=30
EMBEDDING_BATCH_SIZE=50
```

**Solution**: Wait for rate limit reset or adjust limits in .env

### No Files Processed

**Issue**: Sync completes but processes 0 files

**Check**:
```bash
# View detailed logs
docker-compose logs orchestrator | grep -i "file"
docker-compose logs github-discovery

# Verify GitHub configuration
echo $GH_ORGANIZATION
echo $GH_FILTER_KEYWORD
```

**Possible causes**:
- No repositories match filter keyword
- No files changed since last sync (incremental mode)
- Excluded by file extension filters

**Solution**:
```bash
# List matching repositories
curl "http://localhost:8081/repositories?org=$GH_ORGANIZATION&keyword=$GH_FILTER_KEYWORD"

# Try full sync instead of incremental
SYNC_INCREMENTAL=false docker-compose up

# Check file extension filters in .env
ALLOWED_FILE_EXTENSIONS=.md,.rst,.txt,.yaml,.yml,.json
```

## Advanced Usage

### Run Sync on Schedule

Use cron or systemd timer to run sync daily:

```bash
# Cron example (daily at 8 AM)
0 8 * * * cd /path/to/Go_RepoSync_Micro && docker-compose up auto-sync >> /var/log/reposync.log 2>&1
```

### Multiple Syncs Per Day

```bash
# Morning sync (incremental)
docker-compose up auto-sync

# Evening sync (incremental)
docker-compose up auto-sync

# Weekly full sync (Sunday 2 AM)
0 2 * * 0 SYNC_INCREMENTAL=false docker-compose up auto-sync
```

### Background Execution

```bash
# Start services in background
docker-compose up -d

# Wait for services to be ready
sleep 30

# Trigger sync and continue
docker-compose up -d auto-sync

# Check status later
docker-compose logs auto-sync
```

### Custom Filters

Process only specific file types or paths:

**In .env**:
```env
# Only markdown files
ALLOWED_FILE_EXTENSIONS=.md

# Exclude more patterns
EXCLUDE_PATTERNS=node_modules,__pycache__,.git,dist,build,test,examples

# Specific repositories
GH_FILTER_KEYWORD=docs-api
```

## Performance Optimization

### Fast Incremental Updates

```env
SYNC_INCREMENTAL=true
MAX_WORKERS=5
EMBEDDING_BATCH_SIZE=100
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

**Typical time**: 2-10 minutes for 10-50 changed files

### Resource Configuration

**Light Load** (small repos, few files):
```env
MAX_WORKERS=3
EMBEDDING_BATCH_SIZE=50
```

**Heavy Load** (large repos, many files):
```env
MAX_WORKERS=10
EMBEDDING_BATCH_SIZE=100
RATE_LIMIT_REQUESTS_PER_MINUTE=100
```

**Note**: Ensure Docker has sufficient resources (RAM: 4GB+, CPU: 2+ cores)

## Monitoring

### Real-time Logs

```bash
# Follow all logs
docker-compose logs -f

# Specific service
docker-compose logs -f orchestrator
docker-compose logs -f auto-sync

# Search logs
docker-compose logs | grep "error"
docker-compose logs | grep "success"
```

### Log Files

Logs are also written to `./logs/` directory:

```bash
# View orchestrator log
tail -f logs/orchestrator.log

# View all logs
tail -f logs/*.log

# Search for errors
grep -i error logs/*.log
```

### Health Monitoring

```bash
# Check all services
make health

# Or manually
curl http://localhost:8080/health  # Orchestrator
curl http://localhost:8081/health  # GitHub
curl http://localhost:8083/health  # Embedding
curl http://localhost:8084/health  # Vector Storage
```

## Cleanup

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clears data)
docker-compose down -v

# Stop and remove images
docker-compose down --rmi all
```

### Clear Data

```bash
# Remove database and logs
rm -rf data/ logs/

# Recreate directories
mkdir -p data logs
```

## Best Practices

✅ **DO:**
- Use incremental sync for daily updates
- Monitor logs for errors
- Configure Slack notifications
- Test with small repositories first
- Keep credentials secure in .env
- Run full sync weekly for data integrity

❌ **DON'T:**
- Run full sync multiple times per day
- Commit .env file to git
- Ignore error logs
- Use root credentials
- Skip health checks

## Summary

**To automatically sync GitHub changes to Pinecone:**

```bash
# 1. Configure .env with credentials
cp .env.example .env
# Edit .env with your keys

# 2. Run docker-compose
docker-compose up

# 3. Watch the sync happen automatically!
# Services start → Health checks pass → Sync triggers → Embeddings created
```

That's it! The system will automatically:
- ✅ Discover repositories
- ✅ Detect changed files
- ✅ Generate embeddings
- ✅ Store in Pinecone
- ✅ Notify via Slack

---

**For more information:**
- [README.md](../README.md) - Project overview
- [GETTING_STARTED.md](GETTING_STARTED.md) - Initial setup
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design
- [DOCKER_BUILD_TEST.md](DOCKER_BUILD_TEST.md) - Testing guide
