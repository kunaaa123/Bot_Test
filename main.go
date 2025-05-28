package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ===== Configuration =====
const (
	DefaultPort        = ":8080"
	LarkWebhookTimeout = 10 * time.Second
	MaxCommitMessages  = 5
)

// ===== Data Structures =====
type Config struct {
	LarkWebhookURL string
	Port           string
	Environment    string
}

type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"`
	Branch      string `json:"branch"`
	Timestamp   string `json:"timestamp"`
}

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		URL string `json:"url"`
	} `json:"commits"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
}

// ===== Lark Service =====
type LarkService struct {
	webhookURL string
	client     *http.Client
}

func NewLarkService(webhookURL string) *LarkService {
	return &LarkService{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: LarkWebhookTimeout,
		},
	}
}

func (ls *LarkService) SendDeploymentNotification(ctx context.Context, info DeploymentInfo) error {
	payload := ls.buildLarkPayload(info)

	payloadBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ls.webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHub-Webhook-Notifier/1.0")

	resp, err := ls.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("lark API returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Successfully sent notification to Lark for service: %s", info.ServiceName)
	return nil
}

func (ls *LarkService) buildLarkPayload(info DeploymentInfo) map[string]interface{} {
	// Determine environment color
	envColor := ls.getEnvironmentColor(info.ENV)

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("🚀 %s Deployment", strings.ToUpper(info.ENV)),
					"tag":     "plain_text",
				},
				"template": envColor,
			},
			"elements": []map[string]interface{}{
				// Service & Environment Info
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**🔧 Service**\n`%s`", info.ServiceName),
								"tag":     "lark_md",
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**🌐 Environment**\n`%s`", info.ENV),
								"tag":     "lark_md",
							},
						},
					},
				},
				// Deployer & Branch Info
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**👨‍💻 Deployer**\n%s", info.Deployer),
								"tag":     "lark_md",
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**🌿 Branch**\n`%s`", info.Branch),
								"tag":     "lark_md",
							},
						},
					},
				},
				// Separator
				{
					"tag": "hr",
				},
				// Commit Messages
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**📝 Recent Changes**\n%s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
				// Timestamp
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**⏰ Deployed at:** %s", info.Timestamp),
						"tag":     "lark_md",
					},
				},
				// Action Buttons
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔍 View Repository",
								"tag":     "plain_text",
							},
							"type": "primary",
							"url":  info.RepoURL,
						},
					},
				},
			},
		},
	}
}

func (ls *LarkService) getEnvironmentColor(env string) string {
	switch strings.ToUpper(env) {
	case "PROD", "PRODUCTION":
		return "red"
	case "STAGING", "STAGE":
		return "orange"
	case "DEV", "DEVELOPMENT":
		return "blue"
	case "TEST", "TESTING":
		return "purple"
	default:
		return "indigo"
	}
}

// ===== Webhook Handler =====
type WebhookHandler struct {
	larkService *LarkService
	config      *Config
}

func NewWebhookHandler(larkService *LarkService, config *Config) *WebhookHandler {
	return &WebhookHandler{
		larkService: larkService,
		config:      config,
	}
}

func (wh *WebhookHandler) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Method validation
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse GitHub webhook payload
	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		log.Printf("❌ Failed to decode webhook payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate payload
	if err := wh.validatePushEvent(pushEvent); err != nil {
		log.Printf("❌ Invalid push event: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Transform to deployment info
	deploymentInfo := wh.transformToDeploymentInfo(pushEvent)

	// Send notification
	ctx, cancel := context.WithTimeout(context.Background(), LarkWebhookTimeout)
	defer cancel()

	if err := wh.larkService.SendDeploymentNotification(ctx, deploymentInfo); err != nil {
		log.Printf("❌ Failed to send Lark notification: %v", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("🎉 Webhook processed successfully for %s", pushEvent.Repository.Name)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Notification sent successfully"}`))
}

func (wh *WebhookHandler) validatePushEvent(event GitHubPushEvent) error {
	if event.Repository.Name == "" {
		return fmt.Errorf("repository name is required")
	}
	if len(event.Commits) == 0 {
		return fmt.Errorf("no commits found in push event")
	}
	return nil
}

func (wh *WebhookHandler) transformToDeploymentInfo(event GitHubPushEvent) DeploymentInfo {
	// Extract branch name from ref
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Get deployer name (prefer author name, fallback to pusher)
	deployer := event.Pusher.Name
	if len(event.Commits) > 0 && event.Commits[0].Author.Name != "" {
		deployer = event.Commits[0].Author.Name
	}

	// Format commit messages
	commitMsg := wh.formatCommitMessages(event.Commits)

	return DeploymentInfo{
		ENV:         wh.determineEnvironment(branch),
		Deployer:    deployer,
		ServiceName: event.Repository.Name,
		CommitMsg:   commitMsg,
		RepoURL:     event.Repository.HTMLURL,
		Branch:      branch,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05 MST"),
	}
}

func (wh *WebhookHandler) formatCommitMessages(commits []GitHubPushEvent_Commit) string {
	if len(commits) == 0 {
		return "No commit messages available"
	}

	var messages []string
	limit := MaxCommitMessages
	if len(commits) < limit {
		limit = len(commits)
	}

	for i := 0; i < limit; i++ {
		commit := commits[i]
		shortID := commit.ID[:7] // First 7 characters of commit hash
		message := strings.TrimSpace(commit.Message)
		if len(message) > 60 {
			message = message[:57] + "..."
		}
		messages = append(messages, fmt.Sprintf("• `%s` %s", shortID, message))
	}

	if len(commits) > MaxCommitMessages {
		messages = append(messages, fmt.Sprintf("... and %d more commits", len(commits)-MaxCommitMessages))
	}

	return strings.Join(messages, "\n")
}

func (wh *WebhookHandler) determineEnvironment(branch string) string {
	branch = strings.ToLower(branch)
	switch {
	case strings.Contains(branch, "main") || strings.Contains(branch, "master"):
		return "PROD"
	case strings.Contains(branch, "staging") || strings.Contains(branch, "stage"):
		return "STAGING"
	case strings.Contains(branch, "develop") || strings.Contains(branch, "dev"):
		return "DEV"
	case strings.Contains(branch, "test"):
		return "TEST"
	default:
		return "DEV"
	}
}

// Helper type for commits to avoid repetition
type GitHubPushEvent_Commit = struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Author  struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"author"`
	URL string `json:"url"`
}

// ===== Configuration =====
func loadConfig() *Config {
	config := &Config{
		LarkWebhookURL: getEnv("LARK_WEBHOOK_URL", "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"),
		Port:           getEnv("PORT", DefaultPort),
		Environment:    getEnv("APP_ENV", "development"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// ===== Health Check =====
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "lark-notification-service",
	}
	json.NewEncoder(w).Encode(response)
}

// ===== Main Application =====
func main() {
	// Load configuration
	config := loadConfig()

	// Initialize services
	larkService := NewLarkService(config.LarkWebhookURL)
	webhookHandler := NewWebhookHandler(larkService, config)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/github", webhookHandler.HandleGitHubWebhook)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "endpoint not found"}`))
	})

	// Start server
	server := &http.Server{
		Addr:         config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Lark Notification Service starting on port %s", config.Port)
	log.Printf("📊 Environment: %s", config.Environment)
	log.Printf("🔗 Webhook endpoint: http://localhost%s/webhook/github", config.Port)
	log.Printf("❤️  Health check: http://localhost%s/health", config.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
