# Docker Build Troubleshooting Guide

## Quick Fix Summary

### ✅ Issue Resolved!

**Problem**: Docker build was failing with error:
```
failed to solve: process "/bin/sh -c CGO_ENABLED=0 GOOS=linux go build..." did not complete successfully: exit code: 1
```

**Root Cause**: Dockerfiles were copying `pkg/` and `services/` directories individually instead of the entire source tree, causing module import issues.

**Solution**: Updated all Dockerfiles to use `COPY . .` to copy the entire source tree after downloading dependencies.

---

## How to Build

### Option 1: Using Docker Compose (Recommended)

```bash
# Build all services
docker-compose build

# Build in parallel (faster)
docker-compose build --parallel

# Build specific service
docker-compose build orchestrator
```

### Option 2: Using Build Script

```bash
# Sequential build (safer)
./build-docker.sh sequential

# Parallel build (faster)
./build-docker.sh parallel

# Test build (one service)
./build-docker.sh test
```

### Option 3: Using Makefile

```bash
# Build all services
make docker-build

# Build and run
make docker-up
```

### Option 4: Manual Build

```bash
# Build individual service
docker build -t reposync-orchestrator -f services/orchestrator/Dockerfile .

# Build all manually
for service in metadata github-discovery document-processor embedding vector-storage notification orchestrator; do
  docker build -t reposync-$service -f services/$service/Dockerfile .
done
```

---

## Dockerfile Structure (Fixed)

All Dockerfiles now follow this pattern:

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy ALL source code (FIXED!)
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SERVICE_NAME ./services/SERVICE_NAME

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/SERVICE_NAME .
RUN mkdir -p /logs
EXPOSE PORT
CMD ["./SERVICE_NAME"]
```

**Key Change**: `COPY . .` instead of selective copying ensures all dependencies are available.

---

## Troubleshooting Common Issues

### Issue 1: "no such file or directory" during build

**Symptoms**:
```
cannot find package "github.com/nadeeshame/Go_RepoSync_Micro/pkg/..."
```

**Solution**:
```bash
# Ensure you're building from project root
cd /path/to/Go_RepoSync_Micro

# Verify go.mod exists
ls -la go.mod

# Run go mod tidy
go mod tidy

# Rebuild
docker-compose build --no-cache
```

### Issue 2: Build cache issues

**Symptoms**: Old code persists after changes

**Solution**:
```bash
# Clear build cache
docker builder prune -a -f

# Rebuild without cache
docker-compose build --no-cache
```

### Issue 3: Module download failures

**Symptoms**:
```
go: github.com/...: Get "https://...": dial tcp: lookup ... no such host
```

**Solution**:
```bash
# Check internet connectivity
ping google.com

# Set Go proxy
export GOPROXY=https://proxy.golang.org,direct

# Rebuild
docker-compose build
```

### Issue 4: Out of disk space

**Symptoms**:
```
no space left on device
```

**Solution**:
```bash
# Clean up Docker
docker system prune -a -f

# Remove unused images
docker image prune -a -f

# Remove unused volumes
docker volume prune -f
```

### Issue 5: Permission denied

**Symptoms**:
```
permission denied while trying to connect to Docker daemon
```

**Solution**:
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Logout and login again, or:
newgrp docker

# Rebuild
docker-compose build
```

---

## Build Times

Typical build times (with empty cache):

| Service | Build Time | Image Size |
|---------|-----------|------------|
| metadata | ~25s | ~35 MB |
| github-discovery | ~20s | ~25 MB |
| document-processor | ~20s | ~25 MB |
| embedding | ~20s | ~25 MB |
| vector-storage | ~20s | ~25 MB |
| notification | ~20s | ~25 MB |
| orchestrator | ~22s | ~25 MB |
| **Total** | **~2-3 min** | **~185 MB** |

With cache: ~30-60 seconds total

---

## Optimization Tips

### 1. Use BuildKit Cache

```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Build with cache
docker-compose build
```

### 2. Optimize .dockerignore

The `.dockerignore` file excludes unnecessary files:
```
.git/
docs/
*.md
data/
logs/
bin/
```

### 3. Parallel Builds

```bash
# Build all services in parallel
docker-compose build --parallel

# Or use the build script
./build-docker.sh parallel
```

### 4. Layer Caching

Dockerfiles are structured to maximize cache hits:
1. `COPY go.mod go.sum` (rarely changes)
2. `RUN go mod download` (cached if go.mod unchanged)
3. `COPY . .` (changes frequently)
4. `RUN go build` (only runs if source changed)

---

## Verification

After building, verify images:

```bash
# List images
docker images | grep reposync

# Expected output:
# go_reposync_micro-orchestrator       latest    ...    25MB
# go_reposync_micro-metadata           latest    ...    35MB
# go_reposync_micro-github-discovery   latest    ...    25MB
# ...

# Test one service
docker run --rm test-notification ./notification --version 2>&1 | head -5
```

---

## Running After Build

```bash
# Start all services with auto-sync
docker-compose up

# Or in background
docker-compose up -d

# View logs
docker-compose logs -f

# Check health
curl http://localhost:8080/health
```

---

## Clean Rebuild

If you need to completely rebuild from scratch:

```bash
# Stop and remove everything
docker-compose down -v --rmi all

# Clean Docker system
docker system prune -a -f

# Remove local binaries
rm -rf bin/

# Rebuild
docker-compose build --no-cache

# Start fresh
docker-compose up
```

---

## CI/CD Integration

The GitHub Actions workflow handles builds automatically:

```yaml
# .github/workflows/docker-build-test.yml
- name: Build and test
  run: docker-compose build --parallel
```

---

## Support

If build issues persist:

1. **Check logs**: Full build output shows the exact error
2. **Verify go.mod**: Ensure `module github.com/nadeeshame/Go_RepoSync_Micro`
3. **Test local build**: `go build -o bin/orchestrator ./services/orchestrator`
4. **Check Docker version**: `docker --version` (should be 20.10+)
5. **Review Dockerfile**: Ensure `COPY . .` is present

---

## Summary

✅ **All Dockerfiles fixed**
✅ **Build script created** (`build-docker.sh`)
✅ **Tested and working**
✅ **Optimized for caching**
✅ **Ready to use**

**Simple command to build everything:**
```bash
docker-compose build --parallel
```

**Or use the convenience script:**
```bash
./build-docker.sh parallel
```

**Then run:**
```bash
docker-compose up
```

---

**Last Updated**: February 3, 2026
