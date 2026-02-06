# 🚀 Quick Command Reference - Go_RepoSync_Micro

## One-Command Setup & Run

```bash
# 1. Build images (one time)
docker-compose build --parallel

# 2. Start services with auto-sync
docker-compose up
```

**That's it!** Auto-sync happens automatically. 🎉

---

## Docker Commands

### Build
```bash
docker-compose build --parallel          # Build all (fast)
docker-compose build orchestrator        # Build one service
./build-docker.sh parallel               # Use build script
make docker-build                        # Use Makefile
```

### Run
```bash
docker-compose up                        # Start with logs
docker-compose up -d                     # Start in background
./auto-sync.sh                          # Start with auto-sync
make auto-sync                          # Using Makefile
```

### View Logs
```bash
docker-compose logs -f                   # All logs
docker-compose logs -f auto-sync         # Sync progress
docker-compose logs -f orchestrator      # Orchestrator only
docker-compose logs --tail=100           # Last 100 lines
```

### Stop & Clean
```bash
docker-compose down                      # Stop services
docker-compose down -v                   # Stop + remove volumes
docker-compose down --rmi all            # Stop + remove images
docker system prune -a -f                # Clean everything
```

### Status
```bash
docker-compose ps                        # List containers
docker images | grep reposync            # List images
curl http://localhost:8080/health        # Check health
```

---

## Sync Commands

### Auto-Sync (Automatic)
```bash
docker-compose up                        # Runs auto-sync
SYNC_INCREMENTAL=false docker-compose up # Full sync
```

### Manual Sync
```bash
curl -X POST "http://localhost:8080/sync?incremental=true"
make sync                                # Using Makefile
```

### Different Modes
```bash
./auto-sync.sh                          # Incremental (fast)
./auto-sync.sh full                     # Full sync (slow)
make auto-sync                          # Incremental
make auto-sync-full                     # Full
```

---

## Configuration

### Edit Environment
```bash
nano .env                               # Edit configuration
cp .env.example .env                    # Create from template
cat .env | grep -v "^#"                 # View settings
```

### Key Settings
```env
# GitHub
GH_ORGANIZATION=your-org
GH_FILTER_KEYWORD=keyword

# Auto-Sync
SYNC_INCREMENTAL=true                   # Fast incremental
PROJECT_ID=default
```

---

## Troubleshooting

### Build Issues
```bash
docker builder prune -a -f              # Clear build cache
docker-compose build --no-cache         # Rebuild from scratch
./build-docker.sh test                  # Test build
```

### Runtime Issues
```bash
docker-compose logs orchestrator | grep -i error    # Check errors
docker-compose restart orchestrator                 # Restart service
docker-compose down -v && docker-compose up        # Fresh start
```

### Health Checks
```bash
curl http://localhost:8080/health       # Orchestrator
curl http://localhost:8083/health       # Embedding
make health                             # All services
```

---

## Development

### Local Build
```bash
go mod tidy                             # Update dependencies
go build -o bin/orchestrator ./services/orchestrator
./bin/orchestrator                      # Run locally
```

### Test
```bash
go test ./...                           # Run tests
make test                               # Using Makefile
make docker-test                        # Integration tests
```

---

## Common Workflows

### Daily Update
```bash
docker-compose up                       # Auto-sync runs
# Wait for completion
docker-compose down                     # Stop
```

### Weekly Full Sync
```bash
SYNC_INCREMENTAL=false docker-compose up
```

### Background Execution
```bash
docker-compose up -d
docker-compose logs -f auto-sync        # Watch progress
docker-compose ps auto-sync             # Check status
```

### Fresh Install
```bash
git clone <repo>
cd Go_RepoSync_Micro
cp .env.example .env
# Edit .env
docker-compose build --parallel
docker-compose up
```

---

## Useful Aliases

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
alias rs-build='docker-compose build --parallel'
alias rs-up='docker-compose up'
alias rs-down='docker-compose down'
alias rs-logs='docker-compose logs -f'
alias rs-sync='docker-compose logs -f auto-sync'
alias rs-health='make health'
alias rs-clean='docker-compose down -v && docker system prune -a -f'
```

Then use:
```bash
rs-build                                # Build
rs-up                                   # Start
rs-logs                                 # View logs
rs-sync                                 # Watch sync
```

---

## Quick Checks

### Is it working?
```bash
docker-compose ps                       # All running?
curl localhost:8080/health              # Healthy?
docker-compose logs auto-sync | tail -5 # Sync done?
```

### How long did it take?
```bash
docker-compose logs auto-sync | grep "Duration"
```

### What was synced?
```bash
docker-compose logs auto-sync | grep "Summary" -A 10
```

### Any errors?
```bash
docker-compose logs | grep -i error
docker-compose logs orchestrator | grep -i "failed"
```

---

## Emergency Commands

### Service won't start
```bash
docker-compose restart SERVICE_NAME
docker-compose logs SERVICE_NAME
```

### Everything broken
```bash
docker-compose down -v --rmi all
docker system prune -a -f
rm -rf data/ logs/
mkdir -p data logs
docker-compose build --no-cache
docker-compose up
```

### Port conflict
```bash
# Edit .env and change ports
ORCHESTRATOR_PORT=9090
# Then restart
docker-compose down
docker-compose up
```

---

## Documentation Quick Links

- **[README.md](README.md)** - Full documentation
- **[AUTO_SYNC_GUIDE.md](docs/AUTO_SYNC_GUIDE.md)** - Auto-sync details
- **[DOCKER_BUILD_FIX.md](docs/DOCKER_BUILD_FIX.md)** - Build troubleshooting
- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Quick start guide

---

## Support

Having issues?

1. Check logs: `docker-compose logs`
2. Review docs: See links above
3. Clean rebuild: `docker system prune -a -f && docker-compose build --no-cache`
4. Open issue: [GitHub Issues](https://github.com/nadeeshame/Go_RepoSync_Micro/issues)

---

**Remember**: Just run `docker-compose up` and auto-sync does the rest! 🚀
