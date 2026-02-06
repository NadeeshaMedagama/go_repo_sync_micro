#!/bin/bash

# RepoSync Quick Start Script
# This script helps you get RepoSync up and running quickly

set -e

COLOR_RESET='\033[0m'
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[0;33m'
COLOR_RED='\033[0;31m'
COLOR_BLUE='\033[0;34m'

echo -e "${COLOR_BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           RepoSync - Quick Start Script                  ║"
echo "║         Automated RAG Knowledge Base Updater              ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${COLOR_RESET}"

# Check prerequisites
echo -e "${COLOR_YELLOW}Checking prerequisites...${COLOR_RESET}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${COLOR_RED}✗ Go is not installed. Please install Go 1.21+${COLOR_RESET}"
    exit 1
fi
echo -e "${COLOR_GREEN}✓ Go $(go version | awk '{print $3}')${COLOR_RESET}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${COLOR_RED}✗ Docker is not installed${COLOR_RESET}"
    exit 1
fi
echo -e "${COLOR_GREEN}✓ Docker $(docker --version | awk '{print $3}' | sed 's/,//')${COLOR_RESET}"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${COLOR_RED}✗ Docker Compose is not installed${COLOR_RESET}"
    exit 1
fi
echo -e "${COLOR_GREEN}✓ Docker Compose $(docker-compose --version | awk '{print $4}')${COLOR_RESET}"

echo ""
echo -e "${COLOR_YELLOW}Setting up RepoSync...${COLOR_RESET}"

# Create directories
mkdir -p data logs
echo -e "${COLOR_GREEN}✓ Created data and logs directories${COLOR_RESET}"

# Create .env if doesn't exist
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${COLOR_GREEN}✓ Created .env from template${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}⚠ Please edit .env with your credentials before continuing${COLOR_RESET}"
    echo ""
    echo "Required credentials:"
    echo "  - AZURE_OPENAI_API_KEY"
    echo "  - AZURE_OPENAI_ENDPOINT"
    echo "  - GH_TOKEN"
    echo "  - GH_ORGANIZATION"
    echo "  - PINECONE_API_KEY"
    echo "  - PINECONE_INDEX_NAME"
    echo ""
    read -p "Press Enter after configuring .env..."
else
    echo -e "${COLOR_GREEN}✓ .env file exists${COLOR_RESET}"
fi

# Download dependencies
echo ""
echo -e "${COLOR_YELLOW}Downloading Go dependencies...${COLOR_RESET}"
go mod download
echo -e "${COLOR_GREEN}✓ Dependencies downloaded${COLOR_RESET}"

# Choose deployment method
echo ""
echo "Choose deployment method:"
echo "  1) Docker Compose (Recommended)"
echo "  2) Local binaries"
echo "  3) Skip deployment"
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        echo ""
        echo -e "${COLOR_YELLOW}Building Docker images...${COLOR_RESET}"
        docker-compose build
        echo -e "${COLOR_GREEN}✓ Docker images built${COLOR_RESET}"

        echo ""
        echo -e "${COLOR_YELLOW}Starting services...${COLOR_RESET}"
        docker-compose up -d
        echo -e "${COLOR_GREEN}✓ Services started${COLOR_RESET}"

        echo ""
        echo "Waiting for services to be healthy..."
        sleep 10

        # Check health
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${COLOR_GREEN}✓ Orchestrator is healthy${COLOR_RESET}"
        else
            echo -e "${COLOR_RED}✗ Orchestrator is not responding${COLOR_RESET}"
            echo "Check logs: docker-compose logs orchestrator"
            exit 1
        fi
        ;;
    2)
        echo ""
        echo -e "${COLOR_YELLOW}Building local binaries...${COLOR_RESET}"
        make build
        echo -e "${COLOR_GREEN}✓ Binaries built in bin/${COLOR_RESET}"

        echo ""
        echo -e "${COLOR_YELLOW}Note: You need to start services manually in separate terminals:${COLOR_RESET}"
        echo "  ./bin/metadata"
        echo "  ./bin/github-discovery"
        echo "  ./bin/document-processor"
        echo "  ./bin/embedding"
        echo "  ./bin/vector-storage"
        echo "  ./bin/notification"
        echo "  ./bin/orchestrator"
        ;;
    3)
        echo "Skipping deployment"
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo ""
echo -e "${COLOR_GREEN}╔═══════════════════════════════════════════════════════════╗${COLOR_RESET}"
echo -e "${COLOR_GREEN}║              Setup Complete! 🎉                           ║${COLOR_RESET}"
echo -e "${COLOR_GREEN}╚═══════════════════════════════════════════════════════════╝${COLOR_RESET}"
echo ""
echo "Next steps:"
echo ""
echo "1. Check service health:"
echo "   curl http://localhost:8080/health"
echo ""
echo "2. Trigger a sync:"
echo "   curl -X POST 'http://localhost:8080/sync?incremental=true'"
echo ""
echo "3. View logs:"
echo "   docker-compose logs -f orchestrator"
echo ""
echo "4. View all service logs:"
echo "   docker-compose logs -f"
echo ""
echo "5. Stop services:"
echo "   docker-compose down"
echo ""
echo -e "${COLOR_BLUE}For more information, see README.md${COLOR_RESET}"
echo ""
