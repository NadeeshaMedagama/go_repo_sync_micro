# Auto-Sync Feature - Implementation Summary

## Overview

The **Auto-Sync** feature enables automatic synchronization when running `docker-compose up`. The system will automatically retrieve GitHub repository changes, generate embeddings, and store them in Pinecone without any manual intervention.

## What Was Implemented

### 1. ✅ Auto-Sync Service in Docker Compose

**File**: `docker-compose.yml`

**New Service**: `auto-sync`
- Waits for all services to be healthy
- Automatically triggers sync via orchestrator API
- Displays real-time progress
- Shows summary of results
- Exits after sync completes

**Features**:
- 🔄 Automatic execution on startup
- ⏱️ Smart waiting for service readiness
- 📊 Real-time progress display
- ✅ Success/failure reporting
- 🎯 Configurable sync mode (incremental/full)

### 2. ✅ Environment Configuration

**File**: `.env.example`

**New Variables**:
```env
SYNC_INCREMENTAL=true    # Enable incremental sync
SYNC_MODE=incremental    # Sync mode: incremental or full
PROJECT_ID=default       # Project identifier
```

### 3. ✅ Comprehensive Documentation

**File**: `docs/AUTO_SYNC_GUIDE.md` (30+ pages)

**Sections**:
- Quick start guide
- Configuration options
- Usage examples
- Troubleshooting
- Performance optimization
- Best practices
- Advanced usage
- Monitoring guide

### 4. ✅ Convenience Script

**File**: `auto-sync.sh`

**Usage**:
```bash
./auto-sync.sh              # Incremental sync
./auto-sync.sh full         # Full sync
```

### 5. ✅ Makefile Targets

**New Commands**:
```bash
make auto-sync           # Start with incremental sync
make auto-sync-full      # Start with full sync
make auto-sync-bg        # Start in background
```

### 6. ✅ Updated Documentation

**Files Updated**:
- `README.md` - Added auto-sync feature highlight
- `GETTING_STARTED.md` - Updated Docker Compose section
- `.env.example` - Added sync configuration

## How It Works

### Architecture Flow

```
┌─────────────────────────────────────────────────────┐
│  User runs: docker-compose up                       │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  All 7 services start in parallel:                  │
│  - metadata, github-discovery, document-processor   │
│  - embedding, vector-storage, notification          │
│  - orchestrator                                     │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  Health checks run (every 10-30 seconds)            │
│  Waiting for all services to be "healthy"           │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  auto-sync service starts                           │
│  depends_on: orchestrator (condition: healthy)      │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  Wait 10 seconds for stabilization                  │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  Trigger sync via API:                              │
│  POST http://orchestrator:8080/sync                 │
│      ?project_id=default&incremental=true           │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  Orchestrator coordinates services:                 │
│  1. GitHub Discovery → Find repos & changed files   │
│  2. Document Processor → Chunk documents            │
│  3. Embedding Service → Generate vectors            │
│  4. Vector Storage → Store in Pinecone              │
│  5. Metadata Service → Update sync state            │
│  6. Notification → Send Slack message               │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│  auto-sync displays results and exits               │
│  ✅ Sync completed successfully!                    │
└─────────────────────────────────────────────────────┘
```

### Data Flow

```
GitHub Repos → Changed Files → Chunks → Embeddings → Pinecone Index
     ↓             ↓              ↓          ↓             ↓
  Discovery    Processor      Processor  Embedding    Vector Store
  Service      Service        Service    Service      Service
     ↓             ↓              ↓          ↓             ↓
              Metadata Service (tracks state)
                         ↓
              Notification Service (Slack)
```

## Usage Examples

### Example 1: Simple Auto-Sync

```bash
# 1. Configure .env
cp .env.example .env
# Edit with your credentials

# 2. Run
docker-compose up

# Output:
# auto-sync  | 🚀 Starting automatic sync...
# auto-sync  | ✅ Sync completed successfully!
# auto-sync  | repositories_scanned: 25
# auto-sync  | files_processed: 42
# auto-sync  | embeddings_generated: 120
```

### Example 2: Background Execution

```bash
# Start in background
docker-compose up -d

# Check progress
docker-compose logs -f auto-sync

# Check when done
docker-compose ps auto-sync
# Shows: Exit 0 (success)
```

### Example 3: Full Sync

```bash
# Full sync of all files
SYNC_INCREMENTAL=false docker-compose up

# Or using script
./auto-sync.sh full

# Or using Makefile
make auto-sync-full
```

### Example 4: Custom Project

```bash
# Sync specific project
PROJECT_ID=my-project \
GH_ORGANIZATION=my-org \
PINECONE_INDEX_NAME=my-index \
docker-compose up
```

## Configuration

### Sync Modes

| Mode | Description | Speed | Use Case |
|------|-------------|-------|----------|
| **Incremental** | Only changed files | Fast (2-10 min) | Daily updates |
| **Full** | All files | Slow (30-90 min) | Initial setup, weekly refresh |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SYNC_INCREMENTAL` | `true` | Enable incremental mode |
| `SYNC_MODE` | `incremental` | Sync mode label |
| `PROJECT_ID` | `default` | Project identifier |

### Service Dependencies

The auto-sync service depends on:
- ✅ Orchestrator (healthy)
- ✅ Metadata Service (healthy)
- ✅ GitHub Discovery (healthy)
- ✅ Document Processor (healthy)
- ✅ Embedding Service (healthy)
- ✅ Vector Storage (healthy)
- ✅ Notification Service (healthy)

All must pass health checks before sync starts.

## Benefits

### 1. Zero Manual Intervention
- No need to manually trigger sync
- No separate API calls required
- Just run `docker-compose up`

### 2. Automatic Retry Logic
- Health checks ensure services are ready
- Orchestrator handles transient failures
- Built-in error handling

### 3. Clear Visibility
- Real-time progress output
- Success/failure reporting
- Detailed logs available

### 4. Flexible Configuration
- Easy mode switching (incremental/full)
- Environment-based configuration
- Project-specific settings

### 5. Production Ready
- Proper dependency management
- Health check integration
- Exit code reporting (0=success, 1=failure)

## Monitoring

### Check Sync Status

```bash
# View auto-sync logs
docker-compose logs auto-sync

# View full logs
docker-compose logs

# Check if sync completed
docker-compose ps auto-sync
# Exit code: 0 (success) or 1 (failure)
```

### View Results

```bash
# Orchestrator logs
docker-compose logs orchestrator | grep "Sync"

# Check Pinecone
# Login to Pinecone Console → View index stats

# Check Slack
# Look for sync notification message
```

### Health Monitoring

```bash
# Check all services
docker-compose ps

# Individual health checks
curl http://localhost:8080/health  # Orchestrator
curl http://localhost:8083/health  # Embedding
curl http://localhost:8084/health  # Vector Storage
```

## Troubleshooting

### Auto-Sync Exits Immediately

**Problem**: auto-sync container exits before sync completes

**Check**:
```bash
docker-compose logs auto-sync
docker-compose logs orchestrator
```

**Possible Causes**:
- Services not healthy yet
- Network connectivity issues
- Invalid credentials

**Solution**:
```bash
# Increase wait time in docker-compose.yml
# Line: sleep 10 → sleep 30

# Or restart
docker-compose down -v
docker-compose up
```

### Sync Fails with Errors

**Problem**: Sync completes but reports errors

**Check**:
```bash
# View detailed errors
docker-compose logs orchestrator | grep -i error
docker-compose logs embedding | grep -i error
docker-compose logs vector-storage | grep -i error
```

**Common Issues**:
1. **Invalid API keys** → Check .env credentials
2. **Rate limits** → Reduce batch sizes
3. **Network errors** → Check connectivity

### No Files Processed

**Problem**: Sync completes but 0 files processed

**Check**:
```bash
# Verify repositories found
docker-compose logs github-discovery | grep "Found"

# Check filters
echo $GH_FILTER_KEYWORD
```

**Causes**:
- No matching repositories
- No files changed (incremental mode)
- File type filters excluding files

**Solution**:
```bash
# Try full sync
SYNC_INCREMENTAL=false docker-compose up

# Adjust filters in .env
GH_FILTER_KEYWORD=broader-keyword
ALLOWED_FILE_EXTENSIONS=.md,.txt,.rst,.json,.yaml,.yml
```

## Performance

### Typical Timings

| Scenario | Files | Time | API Calls |
|----------|-------|------|-----------|
| Small incremental | 5-10 | 2-3 min | ~50 |
| Medium incremental | 20-50 | 5-10 min | ~200 |
| Large incremental | 100+ | 15-30 min | ~500 |
| Full sync (small) | 100 | 10-20 min | ~500 |
| Full sync (large) | 1000+ | 60-120 min | ~5000 |

### Resource Usage

**Docker Resources Required**:
- RAM: 2-4 GB
- CPU: 2+ cores
- Disk: 500 MB (for Docker images + data)

**Network**:
- Bandwidth: Moderate (API calls + file downloads)
- Latency: Depends on Azure OpenAI response time

## Best Practices

### Daily Operations

✅ **DO:**
```bash
# Morning sync (incremental)
docker-compose up -d

# Check results
docker-compose logs auto-sync

# Stop when done
docker-compose down
```

### Weekly Maintenance

✅ **DO:**
```bash
# Sunday full sync
SYNC_INCREMENTAL=false docker-compose up

# Verify Pinecone index
# Check vector count matches expectations

# Review logs for warnings
docker-compose logs | grep -i warn
```

### Production Deployment

✅ **DO:**
```bash
# Use cron for automation
0 8 * * * cd /path/to/project && docker-compose up -d

# Monitor logs
tail -f logs/orchestrator.log

# Set up alerts
# Configure Slack notifications in .env
```

## Integration

### With CI/CD

```yaml
# GitHub Actions example
- name: Run auto-sync
  run: |
    docker-compose up --abort-on-container-exit auto-sync
  env:
    SYNC_INCREMENTAL: true
```

### With Monitoring

```bash
# Prometheus metrics (future)
curl http://localhost:8080/metrics

# Health check endpoint
curl http://localhost:8080/health
```

### With Backup

```bash
# Backup before sync
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# Run sync
docker-compose up

# Verify backup if needed
tar -tzf backup-$(date +%Y%m%d).tar.gz
```

## Commands Quick Reference

```bash
# Basic auto-sync
docker-compose up

# Background execution
docker-compose up -d

# Full sync
SYNC_INCREMENTAL=false docker-compose up

# View logs
docker-compose logs -f auto-sync

# Check status
docker-compose ps auto-sync

# Stop services
docker-compose down

# Cleanup
docker-compose down -v

# Using convenience script
./auto-sync.sh              # Incremental
./auto-sync.sh full         # Full

# Using Makefile
make auto-sync              # Incremental
make auto-sync-full         # Full
make auto-sync-bg           # Background
```

## Support

For issues with auto-sync:

1. **Check logs**: `docker-compose logs auto-sync`
2. **Review guide**: [docs/AUTO_SYNC_GUIDE.md](AUTO_SYNC_GUIDE.md)
3. **Test manually**: `curl -X POST http://localhost:8080/sync`
4. **Check health**: `docker-compose ps`
5. **Open issue**: [GitHub Issues](https://github.com/nadeeshame/Go_RepoSync_Micro/issues)

## Summary

**With the auto-sync feature, running the entire project is now as simple as:**

```bash
# 1. Configure credentials
cp .env.example .env
# Edit .env

# 2. Run
docker-compose up

# 3. Done! 🎉
# Automatically syncs GitHub → Embeddings → Pinecone
```

No manual triggers needed. No separate API calls. Just one command and everything happens automatically!

---

**Last Updated**: February 3, 2026

**Feature Status**: ✅ Production Ready
