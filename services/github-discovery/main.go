package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/errors"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
	"golang.org/x/oauth2"
)

// GitHubService implements interfaces.RepositoryClient
type GitHubService struct {
	client *github.Client
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(token string) *GitHubService {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &GitHubService{client: client}
}

// ListRepositories finds all repositories matching the filter
func (s *GitHubService) ListRepositories(ctx context.Context, org, keyword string) ([]*models.Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*models.Repository
	for {
		repos, resp, err := s.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, errors.External("GitHub", "failed to list repositories", err)
		}

		for _, repo := range repos {
			if keyword == "" || strings.Contains(strings.ToLower(*repo.Name), strings.ToLower(keyword)) {
				allRepos = append(allRepos, &models.Repository{
					ID:            *repo.ID,
					Name:          *repo.Name,
					FullName:      *repo.FullName,
					Owner:         org,
					DefaultBranch: *repo.DefaultBranch,
					UpdatedAt:     repo.UpdatedAt.Time,
					Private:       *repo.Private,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	logger.Info("Found %d repositories matching keyword '%s'", len(allRepos), keyword)
	return allRepos, nil
}

// GetChangedFiles detects files that changed since last sync
func (s *GitHubService) GetChangedFiles(ctx context.Context, repo *models.Repository, lastCommitSHA string) ([]*models.FileChange, error) {
	var changes []*models.FileChange

	// Get latest commit
	latestCommit, _, err := s.client.Repositories.GetCommit(ctx, repo.Owner, repo.Name, repo.DefaultBranch, nil)
	if err != nil {
		return nil, errors.External("GitHub", "failed to get latest commit", err)
	}

	// If no last commit, fetch all files
	if lastCommitSHA == "" {
		return s.getAllFiles(ctx, repo)
	}

	// Compare commits
	comparison, _, err := s.client.Repositories.CompareCommits(ctx, repo.Owner, repo.Name, lastCommitSHA, *latestCommit.SHA, nil)
	if err != nil {
		return nil, errors.External("GitHub", "failed to compare commits", err)
	}

	// Process changed files
	for _, file := range comparison.Files {
		changeType := "modified"
		if file.Status != nil {
			changeType = *file.Status
		}

		// Skip deleted files - no content to fetch
		if changeType == "removed" || changeType == "deleted" {
			changes = append(changes, &models.FileChange{
				Repository:   repo.FullName,
				FilePath:     *file.Filename,
				CommitSHA:    *latestCommit.SHA,
				LastModified: latestCommit.Commit.Author.Date.Time,
				ChangeType:   changeType,
				Size:         int64(*file.Changes),
			})
			continue
		}

		// Fetch file content for added/modified files
		content, err := s.GetFileContent(ctx, repo.Owner, repo.Name, *file.Filename, repo.DefaultBranch)
		if err != nil {
			logger.Warning("Failed to get content for %s: %v", *file.Filename, err)
			continue
		}

		changes = append(changes, &models.FileChange{
			Repository:   repo.FullName,
			FilePath:     *file.Filename,
			Content:      string(content),
			CommitSHA:    *latestCommit.SHA,
			LastModified: latestCommit.Commit.Author.Date.Time,
			ChangeType:   changeType,
			Size:         int64(*file.Changes),
		})
	}

	logger.Info("Found %d changed files in %s", len(changes), repo.FullName)
	return changes, nil
}

// getAllFiles fetches all files from repository
func (s *GitHubService) getAllFiles(ctx context.Context, repo *models.Repository) ([]*models.FileChange, error) {
	var files []*models.FileChange

	tree, _, err := s.client.Git.GetTree(ctx, repo.Owner, repo.Name, repo.DefaultBranch, true)
	if err != nil {
		return nil, errors.External("GitHub", "failed to get repository tree", err)
	}

	latestSHA, err := s.GetLatestCommitSHA(ctx, repo.Owner, repo.Name, repo.DefaultBranch)
	if err != nil {
		return nil, err
	}

	for _, entry := range tree.Entries {
		if *entry.Type == "blob" {
			// Fetch file content
			content, err := s.GetFileContent(ctx, repo.Owner, repo.Name, *entry.Path, repo.DefaultBranch)
			if err != nil {
				logger.Warning("Failed to get content for %s: %v", *entry.Path, err)
				continue
			}

			files = append(files, &models.FileChange{
				Repository:   repo.FullName,
				FilePath:     *entry.Path,
				Content:      string(content),
				CommitSHA:    latestSHA,
				LastModified: time.Now(),
				ChangeType:   "added",
				Size:         int64(*entry.Size),
			})
		}
	}

	logger.Info("Found %d total files in %s", len(files), repo.FullName)
	return files, nil
}

// GetFileContent retrieves content of a specific file
func (s *GitHubService) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	fileContent, _, _, err := s.client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		return nil, errors.External("GitHub", "failed to get file content", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, errors.External("GitHub", "failed to decode file content", err)
	}

	return []byte(content), nil
}

// GetLatestCommitSHA gets the latest commit SHA for a repository
func (s *GitHubService) GetLatestCommitSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	commit, _, err := s.client.Repositories.GetCommit(ctx, owner, repo, branch, nil)
	if err != nil {
		return "", errors.External("GitHub", "failed to get latest commit SHA", err)
	}
	return *commit.SHA, nil
}

// HTTP Handlers
func (s *GitHubService) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Test GitHub API connection
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, _, err := s.client.Users.Get(ctx, "")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *GitHubService) handleRepositories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	org := r.URL.Query().Get("org")
	keyword := r.URL.Query().Get("keyword")

	if org == "" {
		http.Error(w, "org parameter is required", http.StatusBadRequest)
		return
	}

	repos, err := s.ListRepositories(r.Context(), org, keyword)
	if err != nil {
		logger.Error("Failed to list repositories: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(repos)
}

func (s *GitHubService) handleChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repoFullName := r.URL.Query().Get("repo")
	lastCommit := r.URL.Query().Get("last_commit")

	if repoFullName == "" {
		http.Error(w, "repo parameter is required", http.StatusBadRequest)
		return
	}

	// Parse repo full name (owner/name)
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		http.Error(w, "invalid repo format, expected owner/name", http.StatusBadRequest)
		return
	}

	// Get repository info
	ctx := r.Context()
	ghRepo, _, err := s.client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		logger.Error("Failed to get repository: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo := &models.Repository{
		ID:            *ghRepo.ID,
		Name:          *ghRepo.Name,
		FullName:      *ghRepo.FullName,
		Owner:         parts[0],
		DefaultBranch: *ghRepo.DefaultBranch,
		UpdatedAt:     ghRepo.UpdatedAt.Time,
		Private:       *ghRepo.Private,
	}

	changes, err := s.GetChangedFiles(ctx, repo, lastCommit)
	if err != nil {
		logger.Error("Failed to get changed files: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(changes)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate GitHub-specific requirements
	if err := cfg.ValidateForGitHub(); err != nil {
		fmt.Printf("Failed to validate configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "github-service"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting GitHub Discovery Service on port %d", cfg.Services.GitHubServicePort)

	// Create GitHub service
	service := NewGitHubService(cfg.GitHub.Token)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/repositories", service.handleRepositories)
	mux.HandleFunc("/changes", service.handleChanges)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.GitHubServicePort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down GitHub service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("GitHub Discovery Service listening on port %d", cfg.Services.GitHubServicePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
