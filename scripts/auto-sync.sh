#!/bin/bash

# One-command auto-sync script
# Usage: ./auto-sync.sh [incremental|full]

set -e

MODE=${1:-incremental}

echo "🚀 Go_RepoSync_Micro - Auto Sync"
echo "================================"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    echo "📝 Please copy .env.example to .env and configure it:"
    echo "   cp .env.example .env"
    exit 1
fi

# Set sync mode
if [ "$MODE" = "full" ]; then
    export SYNC_INCREMENTAL=false
    export SYNC_MODE=full
    echo "📊 Mode: Full Sync (all files)"
else
    export SYNC_INCREMENTAL=true
    export SYNC_MODE=incremental
    echo "📊 Mode: Incremental Sync (changed files only)"
fi

echo "🔧 Project: ${PROJECT_ID:-default}"
echo ""
echo "⏳ Starting services and triggering sync..."
echo ""

# Start docker-compose
docker-compose up

echo ""
echo "✅ Done! Check logs above for results."
