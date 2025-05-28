/*
*  ██╗  ██╗ █████╗  ██████╗██╗  ██╗███████╗██████╗
*  ██║  ██║██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗
*  ███████║███████║██║     █████╔╝ █████╗  ██████╔╝
*  ██╔══██║██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
*  ██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║
*  ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
*
*  DEPLOYMENT BOT v1.3.7 - [ACTIVE]
*  Last compile: 2023-11-30 23:59:59 UTC
*  Author: rutchanai
*  License: MIT
 */

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

// ██████╗ ███████╗██████╗ ██╗      ██████╗ ██╗   ██╗███╗   ███╗███████╗███╗   ██╗████████╗
// ██╔══██╗██╔════╝██╔══██╗██║     ██╔═══██╗██║   ██║████╗ ████║██╔════╝████╗  ██║╚══██╔══╝
// ██║  ██║█████╗  ██████╔╝██║     ██║   ██║██║   ██║██╔████╔██║█████╗  ██╔██╗ ██║   ██║
// ██║  ██║██╔══╝  ██╔═══╝ ██║     ██║   ██║██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║
// ██████╔╝███████╗██║     ███████╗╚██████╔╝╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║
// ╚═════╝ ╚══════╝╚═╝     ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝

type DeploymentInfo struct {
	ENV         string `json:"env"`         // Deployment environment (PROD/STAGE/DEV)
	Deployer    string `json:"deployer"`    // Who triggered the deploy
	ServiceName string `json:"serviceName"` // Service identifier
	CommitMsg   string `json:"commitMsg"`   // Git commit message
	RepoURL     string `json:"repoUrl"`     // GitHub repo URL
}

type GitCommitInfo struct {
	Message     string `json:"message"`      // Commit message
	Environment string `json:"environment"`  // Target environment
	ServiceName string `json:"service_name"` // Service name
	Deployer    string `json:"deployer"`     // Committer name
	RepoURL     string `json:"repo_url"`     // Repository URL
}

type GitHubPushEvent struct {
	Repository struct {
		Name string `json:"name"` // Repo name
	} `json:"repository"`
	Commits []struct {
		Message string `json:"message"` // Commit message
		Author  struct {
			Name string `json:"name"` // Author name
		} `json:"author"`
	} `json:"commits"`
}

// ███████╗██╗   ██╗███╗   ██╗ ██████╗████████╗██╗ ██████╗ ███╗   ██╗███████╗
// ██╔════╝██║   ██║████╗  ██║██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║██╔════╝
// █████╗  ██║   ██║██╔██╗ ██║██║        ██║   ██║██║   ██║██╔██╗ ██║███████╗
// ██╔══╝  ██║   ██║██║╚██╗██║██║        ██║   ██║██║   ██║██║╚██╗██║╚════██║
// ██║     ╚██████╔╝██║ ╚████║╚██████╗   ██║   ██║╚██████╔╝██║ ╚████║███████║
// ╚═╝      ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝

func getEnvironmentColor(env string) string {
	env = strings.ToUpper(env)
	switch env {
	case "PROD", "PRODUCTION":
		return "#FF0000" // Red for production
	case "STAGING", "STAGE":
		return "#FFA500" // Orange for staging
	case "DEV", "DEVELOPMENT":
		return "#1E90FF" // Dodger blue for dev
	case "TEST":
		return "#32CD32" // Lime green for test
	default:
		return "#808080" // Grey for unknown
	}
}

// ███████╗███████╗███╗   ██╗██████╗ ███████╗██████╗
// ██╔════╝██╔════╝████╗  ██║██╔══██╗██╔════╝██╔══██╗
// ███████╗█████╗  ██╔██╗ ██║██║  ██║█████╗  ██████╔╝
// ╚════██║██╔══╝  ██║╚██╗██║██║  ██║██╔══╝  ██╔══██╗
// ███████║███████╗██║ ╚████║██████╔╝███████╗██║  ██║
// ╚══════╝╚══════╝╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═╝

func sendToLark(info DeploymentInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("🚀 DEPLOYMENT: %s", strings.ToUpper(info.ENV)),
					"tag":     "plain_text",
				},
				"template": getEnvironmentColor(info.ENV),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Environment**: `%s`\n**Service**: `%s`\n**Deployer**: `%s`\n**Time**: `%s`",
							info.ENV, info.ServiceName, info.Deployer, time.Now().Format("2006-01-02 15:04:05")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Latest Changes**:\n```\n%s\n```", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag": "plain_text",
							"content": fmt.Sprintf("🔗 Repo: %s | 🕒 %s",
								info.RepoURL, time.Now().Format("15:04:05 MST")),
						},
					},
				},
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("❗ ERROR sending to Lark: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Unexpected status from Lark: %d", resp.StatusCode)
	}
	return nil
}

//  ██████╗ ██╗████████╗███████╗██╗     ██████╗
// ██╔════╝ ██║╚══██╔══╝██╔════╝██║     ██╔══██╗
// ██║  ███╗██║   ██║   █████╗  ██║     ██║  ██║
// ██║   ██║██║   ██║   ██╔══╝  ██║     ██║  ██║
// ╚██████╔╝██║   ██║   ███████╗███████╗██████╔╝
//  ╚═════╝ ╚═╝   ╚═╝   ╚══════╝╚══════╝╚═════╝

func sendGitDeploymentToLark(commit GitCommitInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("💾 GIT DEPLOY - %s", strings.ToUpper(commit.Environment)),
					"tag":     "plain_text",
				},
				"template": getEnvironmentColor(commit.Environment),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Service**: `%s`\n**Developer**: `%s`\n**Time**: `%s`",
							commit.ServiceName, commit.Deployer, time.Now().Format("2006-01-02 15:04:05")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Commits**:\n```\n%s\n```", strings.TrimSpace(commit.Message)),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag": "plain_text",
							"content": fmt.Sprintf("🔗 Repo: %s | 🚀 Triggered at %s",
								commit.RepoURL, time.Now().Format("15:04:05 MST")),
						},
					},
				},
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("❗ ERROR sending Git info to Lark: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

// ██╗  ██╗ ██████╗ ████████╗███████╗██╗
// ██║  ██║██╔═══██╗╚══██╔══╝██╔════╝██║
// ███████║██║   ██║   ██║   █████╗  ██║
// ██╔══██║██║   ██║   ██║   ██╔══╝  ██║
// ██║  ██║╚██████╔╝   ██║   ███████╗███████╗
// ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚══════╝╚══════╝

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		http.Error(w, "❌ Unsupported event type", http.StatusBadRequest)
		return
	}

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		http.Error(w, fmt.Sprintf("❌ JSON decode error: %v", err), http.StatusBadRequest)
		return
	}

	// Aggregate commit messages
	var commitMessages strings.Builder
	for _, commit := range pushEvent.Commits {
		commitMessages.WriteString("• " + commit.Message + "\n")
	}

	info := GitCommitInfo{
		Message:     commitMessages.String(),
		Environment: "DEV",
		ServiceName: pushEvent.Repository.Name,
		Deployer:    pushEvent.Commits[0].Author.Name,
		RepoURL:     "https://github.com/kunaaa123/Bot_Test",
	}

	if err := sendGitDeploymentToLark(info); err != nil {
		http.Error(w, fmt.Sprintf("❌ Failed to send notification: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Notification sent successfully"))
}

// ███╗   ███╗ █████╗ ██╗███╗   ██╗
// ████╗ ████║██╔══██╗██║████╗  ██║
// ██╔████╔██║███████║██║██╔██╗ ██║
// ██║╚██╔╝██║██╔══██║██║██║╚██╗██║
// ██║ ╚═╝ ██║██║  ██║██║██║ ╚████║
// ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝

func main() {
	// ASCII art header
	fmt.Println(`
	██████╗ ███████╗██████╗ ██╗      ██████╗ ██╗   ██╗███╗   ███╗███████╗███╗   ██╗████████╗
	██╔══██╗██╔════╝██╔══██╗██║     ██╔═══██╗██║   ██║████╗ ████║██╔════╝████╗  ██║╚══██╔══╝
	██║  ██║█████╗  ██████╔╝██║     ██║   ██║██║   ██║██╔████╔██║█████╗  ██╔██╗ ██║   ██║   
	██║  ██║██╔══╝  ██╔═══╝ ██║     ██║   ██║██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║   
	██████╔╝███████╗██║     ███████╗╚██████╔╝╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║   
	╚═════╝ ╚══════╝╚═╝     ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝   
	`)

	// Endpoint for manual deployment notifications
	http.HandleFunc("/deployment-info", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "DEV",
			Deployer:    "rutchanai",
			ServiceName: "tgth-backend-main",
			CommitMsg:   "Initial deployment\nAdded feature X\nFixed bug Y",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		if err := sendToLark(info); err != nil {
			http.Error(w, fmt.Sprintf("❌ Failed to send notification: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "🚀 Deployment notification sent!",
		})
	})

	// GitHub webhook endpoint
	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("⚡ Server initialized on port %s", port)
	log.Printf("📡 Endpoints:")
	log.Printf("   - POST /deployment-info")
	log.Printf("   - POST /git-webhook")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
