#!/bin/bash

# Build script for Go_RepoSync_Micro Docker images
# This script builds all service images sequentially or in parallel

set -e

COLOR_RESET='\033[0m'
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[0;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'

echo -e "${COLOR_BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     Go_RepoSync_Micro - Docker Build Script              ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${COLOR_RESET}"

SERVICES=("metadata" "github-discovery" "document-processor" "embedding" "vector-storage" "notification" "orchestrator")
BUILD_MODE=${1:-sequential}

if [ "$BUILD_MODE" = "parallel" ]; then
    echo -e "${COLOR_YELLOW}Building all services in parallel...${COLOR_RESET}"
    echo ""

    docker-compose build --parallel

    echo ""
    echo -e "${COLOR_GREEN}✅ All services built successfully!${COLOR_RESET}"

elif [ "$BUILD_MODE" = "sequential" ]; then
    echo -e "${COLOR_YELLOW}Building services sequentially...${COLOR_RESET}"
    echo ""

    for service in "${SERVICES[@]}"; do
        echo -e "${COLOR_BLUE}Building $service...${COLOR_RESET}"
        docker-compose build $service
        echo -e "${COLOR_GREEN}✓ $service built${COLOR_RESET}"
        echo ""
    done

    echo -e "${COLOR_GREEN}✅ All services built successfully!${COLOR_RESET}"

elif [ "$BUILD_MODE" = "test" ]; then
    echo -e "${COLOR_YELLOW}Building and testing one service (notification)...${COLOR_RESET}"
    echo ""

    docker build -t test-notification -f services/notification/Dockerfile .

    echo ""
    echo -e "${COLOR_GREEN}✅ Test build successful!${COLOR_RESET}"

else
    echo -e "${COLOR_RED}Invalid build mode: $BUILD_MODE${COLOR_RESET}"
    echo ""
    echo "Usage: $0 [sequential|parallel|test]"
    echo ""
    echo "  sequential - Build services one by one (default, safer)"
    echo "  parallel   - Build all services at once (faster)"
    echo "  test       - Build one service to test"
    exit 1
fi

echo ""
echo -e "${COLOR_BLUE}Next steps:${COLOR_RESET}"
echo "  1. Configure .env file with your credentials"
echo "  2. Run: docker-compose up"
echo "  3. Services will start and auto-sync will trigger"
echo ""
