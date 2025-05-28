package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ###########################################
// ##               CONFIG                 ##
// ###########################################
const (
	LARK_WEBHOOK = "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"
	VERSION      = "v1.3.7"
)

// ###########################################
// ##             DATA STRUCTS             ##
// ###########################################
type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"`
}

type GitCommitInfo struct {
	Message     string `json:"message"`
	Environment string `json:"environment"`
	ServiceName string `json:"service_name"`
	Deployer    string `json:"deployer"`
	RepoURL     string `json:"repo_url"`
}

type GitHubPushEvent struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Commits []struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commits"`
}

// ###########################################
// ##              UTILITIES               ##
// ###########################################
func getEnvironmentColor(env string) string {
	env = strings.ToUpper(env)
	colors := map[string]string{
		"PROD":        "#FF2D2D", // Red alert for production
		"PRODUCTION":  "#FF2D2D",
		"STAGING":     "#FF8F1C", // Orange for staging
		"STAGE":       "#FF8F1C",
		"DEV":         "#1C6BFF", // Blue for dev
		"DEVELOPMENT": "#1C6BFF",
		"TEST":        "#00CC66", // Green for test
	}
	if color, exists := colors[env]; exists {
		return color
	}
	return "#666666" // Default gray
}

func generateHackerHeader(title, env string) map[string]interface{} {
	return map[string]interface{}{
		"title": map[string]interface{}{
			"content": fmt.Sprintf("⚡ %s ⚡", title),
			"tag":     "plain_text",
		},
		"template": getEnvironmentColor(env),
	}
}

func generateTerminalText(content string) map[string]interface{} {
	return map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("```\n%s\n```", strings.TrimSpace(content)),
			"tag":     "lark_md",
		},
	}
}

func generateSystemInfo(env, service, deployer string) map[string]interface{} {
	return map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf(
				"🌐 *ENV*: `%s`\n"+
					"🖥️ *SERVICE*: `%s`\n"+
					"👨‍💻 *DEPLOYER*: `%s`\n"+
					"🕒 *TIME*: `%s`",
				env, service, deployer, time.Now().Format("2006-01-02 15:04:05")),
			"tag": "lark_md",
		},
	}
}

// ###########################################
// ##           NOTIFICATION LOGIC          ##
// ###########################################
func sendHackerNotification(info DeploymentInfo) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": generateHackerHeader("DEPLOYMENT INITIATED", info.ENV),
			"elements": []map[string]interface{}{
				generateSystemInfo(info.ENV, info.ServiceName, info.Deployer),
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": "📌 *COMMIT MESSAGE*",
						"tag":     "lark_md",
					},
				},
				generateTerminalText(info.CommitMsg),
				{
					"tag": "hr",
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("🔗 Repo: %s | 🚀 DeployBot %s", info.RepoURL, VERSION),
						},
					},
				},
			},
		},
	}

	return sendPayload(payload)
}

func sendGitDeployNotification(commit GitCommitInfo) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": generateHackerHeader("GIT DEPLOY DETECTED", commit.Environment),
			"elements": []map[string]interface{}{
				generateSystemInfo(commit.Environment, commit.ServiceName, commit.Deployer),
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": "💾 *COMMIT LOG*",
						"tag":     "lark_md",
					},
				},
				generateTerminalText(commit.Message),
				{
					"tag": "hr",
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("🔗 Repo: %s | 🚀 DeployBot %s", commit.RepoURL, VERSION),
						},
					},
				},
			},
		},
	}

	return sendPayload(payload)
}

func sendPayload(payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("payload marshal error: %v", err)
	}

	resp, err := http.Post(LARK_WEBHOOK, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("webhook post error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}
	return nil
}

// ###########################################
// ##          WEBHOOK HANDLERS             ##
// ###########################################
func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-GitHub-Event") != "push" {
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		return
	}

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var commitMessages strings.Builder
	for _, commit := range pushEvent.Commits {
		commitMessages.WriteString("• ")
		commitMessages.WriteString(commit.Message)
		commitMessages.WriteString("\n")
	}

	info := GitCommitInfo{
		Message:     commitMessages.String(),
		Environment: "DEV",
		ServiceName: pushEvent.Repository.Name,
		Deployer:    pushEvent.Commits[0].Author.Name,
		RepoURL:     "https://github.com/kunaaa123/Bot_Test",
	}

	if err := sendGitDeployNotification(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Notification sent"))
}

// ###########################################
// ##               MAIN                    ##
// ###########################################
func main() {
	// Test endpoint
	http.HandleFunc("/deploy-test", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "PROD",
			Deployer:    "admin",
			ServiceName: "core-system",
			CommitMsg:   "🚀 Critical security update\n✅ Fixed auth vulnerability\n➕ Added rate limiting",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		if err := sendHackerNotification(info); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "notification_sent"})
	})

	// GitHub webhook endpoint
	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 DeployBot %s initializing...", VERSION)
	log.Printf("🔌 Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
