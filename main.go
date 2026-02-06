package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		fmt.Println("RepoSync v1.0.0 - Go Microservices Edition")
		fmt.Println("Built with SOLID principles for automated RAG knowledge base updates")
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `
RepoSync - Automated RAG Knowledge Base Updater (Go Microservices)

USAGE:
    reposync <command>

COMMANDS:
    help        Show this help message
    version     Show version information

MICROSERVICES ARCHITECTURE:
    This system runs as distributed microservices. Use Docker Compose or
    individual service binaries:

    Services:
        - orchestrator (port 8080)     Main coordinator
        - metadata (port 8086)         Metadata tracking
        - github-discovery (port 8081) GitHub API integration
        - document-processor (port 8082) Document chunking
        - embedding (port 8083)        Azure OpenAI embeddings
        - vector-storage (port 8084)   Pinecone integration
        - notification (port 8085)     Slack notifications

GETTING STARTED:
    1. Copy .env.example to .env and configure credentials
    2. Run with Docker Compose:
       docker-compose up
    3. Trigger sync:
       curl -X POST "http://localhost:8080/sync?incremental=true"

    OR build and run locally:
       go build -o bin/orchestrator ./services/orchestrator
       ./bin/orchestrator

GITHUB ACTIONS:
    Automated daily sync at 8:00 AM UTC via GitHub Actions workflow.
    See .github/workflows/scheduled-sync.yml

DOCUMENTATION:
    See README.md for complete documentation
    See docs/ for architecture and setup guides

For more information, visit: https://github.com/nadeeshame/Go_RepoSync_Micro
`
	fmt.Println(usage)
}
