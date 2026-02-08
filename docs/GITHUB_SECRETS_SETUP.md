# GitHub Actions Secrets Setup Guide

This guide explains how to configure GitHub Secrets for automated daily syncs.

## Required Secrets

Navigate to your repository: **Settings → Secrets and variables → Actions → New repository secret**

### 1. Azure OpenAI Configuration

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `AZURE_OPENAI_API_KEY` | Azure OpenAI API key | `abc123...` |
| `AZURE_OPENAI_ENDPOINT` | Azure OpenAI endpoint URL | `https://your-resource.openai.azure.com/` |
| `AZURE_OPENAI_EMBEDDINGS_DEPLOYMENT` | Deployment name for embeddings | `text-embedding-ada-002` |
| `AZURE_OPENAI_API_VERSION` | API version | `2023-05-15` |

**How to get Azure OpenAI credentials:**
1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to your Azure OpenAI resource
3. Go to "Keys and Endpoint"
4. Copy the key and endpoint

### 2. GitHub Configuration

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GH_TOKEN` | Personal access token | `ghp_abc123...` |
| `GH_ORGANIZATION` | GitHub organization name | `your-org` |
| `GH_FILTER_KEYWORD` | Keyword to filter repositories | `docs` |

**How to create GitHub Personal Access Token:**
1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `read:org` (Read org and team membership)
4. Click "Generate token"
5. **Copy the token immediately** (you won't see it again!)

### 3. Pinecone Configuration

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `PINECONE_API_KEY` | Pinecone API key | `abc123-def456...` |
| `PINECONE_INDEX_NAME` | Pinecone index name | `reposync-index` |
| `PINECONE_DIMENSION` | Vector dimension (optional) | `1536` |
| `PINECONE_CLOUD` | Cloud provider (optional) | `aws` |
| `PINECONE_REGION` | Region (optional) | `us-east-1` |

**How to get Pinecone credentials:**
1. Go to [Pinecone Console](https://app.pinecone.io)
2. Navigate to "API Keys"
3. Copy your API key
4. Note your index name from "Indexes" page

**Create Pinecone Index** (if not exists):
```bash
curl -X POST "https://api.pinecone.io/indexes" \
  -H "Api-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "reposync-index",
    "dimension": 1536,
    "metric": "cosine"
  }'
```

### 4. Slack Configuration (Optional)

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `SLACK_WEBHOOK_URL` | Incoming webhook URL | `https://hooks.slack.com/services/...` |

**How to create Slack Webhook:**
1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app (e.g., "RepoSync") and select workspace
4. Go to "Incoming Webhooks" → Activate → "Add New Webhook to Workspace"
5. Select channel and click "Allow"
6. Copy the webhook URL

### 5. Docker Hub (Optional - for Docker image publishing)

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `DOCKER_USERNAME` | Docker Hub username | `your-username` |
| `DOCKER_PASSWORD` | Docker Hub password/token | `abc123...` |

## Verification

After adding secrets, verify they're set correctly:

1. Go to **Settings → Secrets and variables → Actions**
2. You should see all secrets listed (values are hidden)
3. No green checkmark = secret not set

## Testing GitHub Actions

### Manual Trigger

1. Go to **Actions** tab
2. Select "Scheduled Sync" workflow
3. Click "Run workflow"
4. Configure options:
   - Incremental sync: `true` (default)
   - Project ID: `default` (or your project ID)
5. Click "Run workflow"

### View Logs

1. Click on the running workflow
2. Click on the "sync" job
3. Expand steps to see detailed logs
4. Download artifacts (logs and sync results) at the bottom

### Schedule Verification

The workflow runs automatically at **8:00 AM UTC daily**.

To verify the schedule:
1. Check `.github/workflows/scheduled-sync.yml`
2. Look for:
   ```yaml
   schedule:
     - cron: '0 8 * * *'  # Daily at 8:00 AM UTC
   ```

## Troubleshooting

### Secret Not Found Error

```
Error: Secret AZURE_OPENAI_API_KEY not found
```

**Solution**: Ensure secret name matches exactly (case-sensitive)

### Invalid Credentials Error

```
Error: 401 Unauthorized
```

**Solution**: 
- Verify secret values are correct
- Check if tokens have expired
- Ensure sufficient permissions/scopes

### GitHub API Rate Limit

```
Error: API rate limit exceeded
```

**Solution**:
- Use GitHub token (authenticated requests get 5000/hour)
- Reduce `RATE_LIMIT_REQUESTS_PER_MINUTE`
- Wait for rate limit reset

### Workflow Not Running

**Check**:
1. Actions are enabled: Settings → Actions → Allow all actions
2. Workflow file syntax is correct: Check for YAML errors
3. No branch protection rules blocking workflows

## Security Best Practices

✅ **DO:**
- Rotate secrets regularly (every 90 days)
- Use read-only tokens when possible
- Limit token scopes to minimum required
- Use separate tokens for different purposes
- Monitor secret usage in audit logs

❌ **DON'T:**
- Commit secrets to code
- Share secrets in plaintext
- Use production secrets in test environments
- Reuse the same token across multiple services

## Example: Complete Setup

```bash
# Step 1: Create .env locally for testing
cp .env.example .env

# Step 2: Fill in .env
AZURE_OPENAI_API_KEY=sk-abc123...
AZURE_OPENAI_ENDPOINT=https://my-openai.openai.azure.com/
GH_TOKEN=ghp_xyz789...
GH_ORGANIZATION=my-org
PINECONE_API_KEY=abc-123-def-456
PINECONE_INDEX_NAME=reposync-index

# Step 3: Test locally
docker-compose up -d
curl -X POST "http://localhost:8080/sync?incremental=true"

# Step 4: Add same values as GitHub Secrets
# (via GitHub UI, not command line for security)

# Step 5: Trigger GitHub Action
# Manual trigger from Actions tab

# Step 6: Verify success
# Check workflow logs and artifacts
```

## Support

If you encounter issues:

1. **Check workflow logs**: Actions → Select run → View logs
2. **Verify secrets**: Settings → Secrets → Check all required secrets exist
3. **Test locally**: Run `docker-compose up` with same .env values
4. **Check service health**: View individual service logs

For more help, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
