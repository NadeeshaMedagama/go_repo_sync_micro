# Docker Build and Test Guide

This guide explains how to build, test, and deploy Docker images for Go_RepoSync_Micro.

## Table of Contents

- [Local Testing](#local-testing)
- [GitHub Actions CI/CD](#github-actions-cicd)
- [Docker Build Process](#docker-build-process)
- [Multi-Platform Builds](#multi-platform-builds)
- [Security Scanning](#security-scanning)
- [Troubleshooting](#troubleshooting)

## Local Testing

### Quick Test

Run all services and tests locally:

```bash
# Build and test all services
make docker-test-build

# Or use docker-compose directly
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Individual Service Testing

Build and test a single service:

```bash
# Build specific service
docker build -t reposync-orchestrator -f services/orchestrator/Dockerfile .

# Run the service
docker run -p 8080:8080 --env-file .env reposync-orchestrator

# Test health endpoint
curl http://localhost:8080/health
```

### Integration Testing

Test all services together:

```bash
# Start all services
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be healthy
sleep 30

# Run health checks
for port in 8080 8081 8082 8083 8084 8085 8086; do
  echo "Testing port $port..."
  curl -f http://localhost:$port/health || echo "Failed: $port"
done

# View logs
docker-compose -f docker-compose.test.yml logs

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

## GitHub Actions CI/CD

### Workflow: `docker-build-test.yml`

This workflow automatically:

1. **Builds all service images** on push/PR
2. **Runs security scans** with Trivy
3. **Performs integration tests** with docker-compose
4. **Creates multi-platform builds** (amd64, arm64)
5. **Generates SBOM** (Software Bill of Materials)
6. **Publishes to GitHub Container Registry**

### Workflow Triggers

```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'services/**'
      - 'pkg/**'
      - 'go.mod'
      - 'docker-compose.yml'
  pull_request:
    branches: [ main, develop ]
  workflow_dispatch:  # Manual trigger
```

### Jobs Overview

#### 1. Build Services
- Builds Docker images for all 7 services
- Uses BuildKit caching for faster builds
- Parallel execution (matrix strategy)
- Pushes to GitHub Container Registry (GHCR)

#### 2. Integration Test
- Starts all services with docker-compose
- Waits for health checks
- Tests API endpoints
- Collects and uploads logs

#### 3. Multi-Platform Build
- Builds for amd64 and arm64
- Only runs on main branch
- Uses QEMU for cross-compilation

#### 4. Validate Compose
- Validates docker-compose.yml syntax
- Checks Dockerfiles for best practices

#### 5. Generate SBOM
- Creates Software Bill of Materials
- Lists all dependencies
- Useful for security audits

#### 6. Summary
- Generates build summary
- Reports success/failure

### Manual Trigger

1. Go to **Actions** tab
2. Select "Docker Build and Test"
3. Click "Run workflow"
4. Select branch and click "Run"

## Docker Build Process

### Build Arguments

Each Dockerfile supports build arguments:

```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -t reposync-orchestrator:v1.0.0 \
  -f services/orchestrator/Dockerfile \
  .
```

### Multi-Stage Builds

All services use multi-stage builds for optimization:

```dockerfile
# Stage 1: Builder
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o service ./services/service-name

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/service .
CMD ["./service"]
```

**Benefits:**
- Small final image size (~15-30 MB)
- No build tools in production image
- Improved security (minimal attack surface)

### Cache Optimization

GitHub Actions uses BuildKit cache:

```yaml
cache-from: type=gha,scope=service-name
cache-to: type=gha,mode=max,scope=service-name
```

**Result:** 
- First build: ~5-10 minutes
- Cached builds: ~1-2 minutes

## Multi-Platform Builds

### Supported Platforms

- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64-bit)

### Building Locally

```bash
# Setup buildx
docker buildx create --use --name multiarch

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t reposync-orchestrator:latest \
  -f services/orchestrator/Dockerfile \
  --push \
  .
```

### Platform-Specific Testing

```bash
# Test on current platform
docker run --platform linux/amd64 reposync-orchestrator:latest

# Test on ARM (if available)
docker run --platform linux/arm64 reposync-orchestrator:latest
```

## Security Scanning

### Trivy Vulnerability Scanning

Automatically scans all images for vulnerabilities:

```bash
# Scan locally
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy:latest image \
  reposync-orchestrator:latest
```

### Severity Levels

- **CRITICAL**: Immediate action required
- **HIGH**: Should be fixed soon
- **MEDIUM**: Fix when convenient
- **LOW**: Informational

### GitHub Security Alerts

Scan results are uploaded to GitHub Security tab:

1. Go to **Security** → **Code scanning alerts**
2. Filter by "trivy"
3. Review and remediate vulnerabilities

### Best Practices

✅ **DO:**
- Pin base image versions (e.g., `alpine:3.19`)
- Use minimal base images
- Update dependencies regularly
- Scan before production deployment
- Review and fix CRITICAL/HIGH vulnerabilities

❌ **DON'T:**
- Use `:latest` tag in production
- Include unnecessary packages
- Run containers as root
- Store secrets in images

## Image Registry

### GitHub Container Registry (GHCR)

Images are published to:
```
ghcr.io/nadeeshame/go-reposync-micro-SERVICE:TAG
```

**Examples:**
```
ghcr.io/nadeeshame/go-reposync-micro-orchestrator:latest
ghcr.io/nadeeshame/go-reposync-micro-orchestrator:main
ghcr.io/nadeeshame/go-reposync-micro-orchestrator:abc123
```

### Pulling Images

```bash
# Login to GHCR (if private)
echo $GH_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull image
docker pull ghcr.io/nadeeshame/go-reposync-micro-orchestrator:latest

# Run from registry
docker run -p 8080:8080 \
  --env-file .env \
  ghcr.io/nadeeshame/go-reposync-micro-orchestrator:latest
```

### Tagging Strategy

- `latest` - Latest build from main branch
- `main` - Latest from main branch
- `develop` - Latest from develop branch
- `pr-123` - Pull request builds
- `abc123` - Git commit SHA
- `v1.0.0` - Semantic version tags

## Troubleshooting

### Build Failures

**Error:** `failed to solve with frontend dockerfile.v0`

**Solution:**
```bash
# Clear BuildKit cache
docker buildx prune -a -f

# Rebuild
docker-compose build --no-cache
```

**Error:** `go.mod: no such file or directory`

**Solution:**
- Check Dockerfile context is repository root
- Ensure COPY paths are correct

### Health Check Failures

**Error:** Service not responding to health checks

**Solution:**
```bash
# Check service logs
docker-compose logs SERVICE_NAME

# Test manually
docker exec CONTAINER_NAME wget -O- http://localhost:PORT/health

# Increase health check timeout
healthcheck:
  timeout: 10s
  retries: 10
```

### Integration Test Failures

**Error:** `test-runner exited with code 1`

**Solution:**
```bash
# Run tests with verbose output
docker-compose -f docker-compose.test.yml up

# Check individual service health
docker-compose -f docker-compose.test.yml ps

# View detailed logs
docker-compose -f docker-compose.test.yml logs --tail=100
```

### Memory Issues

**Error:** `signal: killed` or OOM errors

**Solution:**
```bash
# Increase Docker memory limit
# Docker Desktop → Settings → Resources → Memory

# Or add resource limits
services:
  orchestrator:
    deploy:
      resources:
        limits:
          memory: 1G
```

### Network Issues

**Error:** Services can't communicate

**Solution:**
```bash
# Check network
docker network ls
docker network inspect go_reposync_micro_reposync-test-network

# Use service names (not localhost)
GITHUB_SERVICE_URL=http://github-discovery:8081  # Correct
GITHUB_SERVICE_URL=http://localhost:8081         # Wrong in Docker
```

## Performance Optimization

### Build Speed

```bash
# Use BuildKit
export DOCKER_BUILDKIT=1

# Enable layer caching
docker-compose build --parallel

# Use .dockerignore
echo "data/" >> .dockerignore
echo "logs/" >> .dockerignore
echo ".git/" >> .dockerignore
```

### Runtime Performance

```bash
# Use proper resource limits
services:
  embedding:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 512M
```

### Image Size

Current image sizes:
- Orchestrator: ~25 MB
- Metadata: ~35 MB (includes SQLite)
- Other services: ~20-25 MB each

**Optimization tips:**
- Use alpine base images
- Multi-stage builds
- Remove unnecessary files
- Use `.dockerignore`

## Monitoring

### View Build Logs

```bash
# GitHub Actions
# Go to Actions → Select workflow run → View logs

# Local
docker-compose logs -f SERVICE_NAME
```

### Image Inspection

```bash
# View image details
docker inspect reposync-orchestrator:latest

# View layers
docker history reposync-orchestrator:latest

# View size
docker images | grep reposync
```

### Container Stats

```bash
# Real-time stats
docker stats

# Specific service
docker stats reposync-orchestrator
```

## CI/CD Best Practices

### Pre-commit Checks

```bash
# Validate before committing
make lint
make test
docker-compose config
```

### Branch Protection

Recommended branch protection rules:
- ✅ Require status checks (CI/CD passing)
- ✅ Require review before merge
- ✅ Require up-to-date branches
- ✅ Include administrators

### Deployment Strategy

1. **Development**: Push to `develop` branch
2. **Testing**: Create PR to `main`
3. **Staging**: Merge to `main` (auto-deploy to staging)
4. **Production**: Create release tag (manual approval)

## Support

For issues with Docker builds:

1. Check workflow logs in Actions tab
2. Test locally with `make docker-test`
3. Review [Troubleshooting](#troubleshooting) section
4. Open issue on GitHub

---

**Last Updated**: February 3, 2026
