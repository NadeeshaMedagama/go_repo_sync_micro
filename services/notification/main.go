package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/config"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/errors"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/logger"
	"github.com/nadeeshame/Go_RepoSync_Micro/pkg/models"
	"github.com/slack-go/slack"
)

// NotificationService implements interfaces.NotificationService
type NotificationService struct {
	webhookURL string
}

// NewNotificationService creates a new notification service
func NewNotificationService(webhookURL string) *NotificationService {
	return &NotificationService{
		webhookURL: webhookURL,
	}
}

// SendNotification sends a notification
func (s *NotificationService) SendNotification(ctx context.Context, payload *models.NotificationPayload) error {
	if s.webhookURL == "" {
		logger.Warning("Slack webhook URL not configured, skipping notification")
		return nil
	}
	return s.SendSlack(ctx, payload)
}

// SendSlack sends a Slack notification
func (s *NotificationService) SendSlack(ctx context.Context, payload *models.NotificationPayload) error {
	if s.webhookURL == "" {
		return nil
	}

	// Build Slack message
	msg := s.buildSlackMessage(payload)

	// Send webhook
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return errors.Internal("failed to marshal Slack message", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Network("failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Network("failed to send Slack notification", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.External("Slack", fmt.Sprintf("unexpected status code %d: %s", resp.StatusCode, body), nil)
	}

	logger.Info("Slack notification sent successfully")
	return nil
}

// buildSlackMessage builds a formatted Slack message
func (s *NotificationService) buildSlackMessage(payload *models.NotificationPayload) *slack.WebhookMessage {
	var color string
	var emoji string

	switch payload.Type {
	case "success":
		color = "good"
		emoji = ":white_check_mark:"
	case "error":
		color = "danger"
		emoji = ":x:"
	case "warning":
		color = "warning"
		emoji = ":warning:"
	default:
		color = "#439FE0"
		emoji = ":information_source:"
	}

	attachment := slack.Attachment{
		Color:      color,
		Title:      fmt.Sprintf("%s %s", emoji, payload.Title),
		Text:       payload.Message,
		Footer:     "RepoSync",
		FooterIcon: "https://github.com/favicon.ico",
		Ts:         json.Number(fmt.Sprintf("%d", payload.Timestamp.Unix())),
	}

	// Add result details if available
	if payload.Result != nil {
		result := payload.Result
		fields := []slack.AttachmentField{
			{
				Title: "Duration",
				Value: result.Duration.String(),
				Short: true,
			},
			{
				Title: "Repositories",
				Value: fmt.Sprintf("%d", result.RepositoriesScanned),
				Short: true,
			},
			{
				Title: "Files Processed",
				Value: fmt.Sprintf("%d / %d", result.FilesProcessed, result.FilesChanged),
				Short: true,
			},
			{
				Title: "Embeddings Generated",
				Value: fmt.Sprintf("%d", result.EmbeddingsGenerated),
				Short: true,
			},
		}

		if len(result.Errors) > 0 {
			fields = append(fields, slack.AttachmentField{
				Title: "Errors",
				Value: fmt.Sprintf("```%s```", result.Errors[0]),
				Short: false,
			})
		}

		attachment.Fields = fields
	}

	return &slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}
}

// HTTP Handlers
func (s *NotificationService) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.SendNotification(r.Context(), &payload); err != nil {
		logger.Error("Failed to send notification: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (s *NotificationService) handleHealth(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.FilePath, "notification-service"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Notification Service on port %d", cfg.Services.NotificationServicePort)

	// Create notification service
	service := NewNotificationService(cfg.Notifications.SlackWebhookURL)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/notify", service.handleNotify)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.NotificationServicePort),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down notification service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()

	// Start server
	logger.Info("Notification Service listening on port %d", cfg.Services.NotificationServicePort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: %v", err)
	}
}
