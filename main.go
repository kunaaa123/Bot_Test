package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"`
}

type GitCommitInfo struct {
	Message     string   `json:"message"`
	Features    []string `json:"features"`
	Environment string   `json:"environment"`
	ServiceName string   `json:"service_name"`
	Deployer    string   `json:"deployer"`
	RepoURL     string   `json:"repo_url"`
}

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Commits []struct {
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

func getEnvironmentColor(env string) string {
	switch strings.ToUpper(env) {
	case "PROD", "PRODUCTION":
		return "red"
	case "STAGING", "STAGE":
		return "orange"
	case "DEV", "DEVELOPMENT":
		return "blue"
	case "TEST":
		return "green"
	default:
		return "grey"
	}
}

func getEnvironmentEmoji(env string) string {
	switch strings.ToUpper(env) {
	case "PROD", "PRODUCTION":
		return "🚀"
	case "STAGING", "STAGE":
		return "🔄"
	case "DEV", "DEVELOPMENT":
		return "⚡"
	case "TEST":
		return "🧪"
	default:
		return "📦"
	}
}

func sendToLark(info DeploymentInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	envColor := getEnvironmentColor(info.ENV)
	envEmoji := getEnvironmentEmoji(info.ENV)

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("%s Backend Deployment Success", envEmoji),
					"tag":     "plain_text",
				},
				"template": envColor,
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**🌍 Environment:** `%s`\n**👤 Deployer:** `%s`\n**⚙️ Service:** `%s`",
							info.ENV, info.Deployer, info.ServiceName),
						"tag": "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**📝 Latest Changes:**\n```\n%s\n```", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**⏰ Deployment Time:** %s\n**✅ Status:** Successfully Deployed",
							time.Now().Format("2006-01-02 15:04:05 MST")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": "🎉 Deployment completed successfully! All systems are operational.",
						},
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔗 View Repository",
								"tag":     "plain_text",
							},
							"type": "primary",
							"url":  info.RepoURL,
						},
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "📚 Documentation",
								"tag":     "plain_text",
							},
							"type": "default",
							"url":  "https://github.com/kunaaa123/Bot_Test/wiki",
						},
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "📊 Monitoring",
								"tag":     "plain_text",
							},
							"type": "default",
							"url":  "#",
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error sending to Lark: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("Lark Response: %s", string(body))
	return nil
}

func sendGitDeploymentToLark(commit GitCommitInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	envColor := getEnvironmentColor(commit.Environment)
	envEmoji := getEnvironmentEmoji(commit.Environment)

	// Count commits
	commitCount := strings.Count(commit.Message, "- ")
	if commitCount == 0 && commit.Message != "" {
		commitCount = 1
	}

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("🚀 New Deployment - %s", strings.ToUpper(commit.Environment)),
					"tag":     "plain_text",
				},
				"template": envColor,
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**%s Environment:** `%s`\n**👨‍💻 Developer:** `%s`\n**📦 Service:** `%s`\n**📊 Commits:** `%d`",
							envEmoji, commit.Environment, commit.Deployer, commit.ServiceName, commitCount),
						"tag": "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**📋 Commit Messages:**\n```\n%s```", strings.TrimSpace(commit.Message)),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**⏰ Deployed at:** %s", time.Now().Format("2006-01-02 15:04:05 MST")),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("🎯 Auto-deployment triggered by Git push to %s environment", commit.Environment),
						},
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔗 View Repository",
								"tag":     "plain_text",
							},
							"type": "primary",
							"url":  commit.RepoURL,
						},
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔍 View Changes",
								"tag":     "plain_text",
							},
							"type": "default",
							"url":  fmt.Sprintf("%s/commits", commit.RepoURL),
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("Sending beautiful payload to Lark: %s", string(payloadBytes))

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("Lark Response: %s", string(body))
	return nil
}

func sendErrorNotification(errorMsg string, service string) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "❌ Deployment Failed",
					"tag":     "plain_text",
				},
				"template": "red",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**⚠️ Service:** `%s`\n**🕐 Failed at:** %s",
							service, time.Now().Format("2006-01-02 15:04:05 MST")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**🐛 Error Details:**\n```\n%s\n```", errorMsg),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": "🚨 Immediate attention required! Please check the deployment logs.",
						},
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔧 View Logs",
								"tag":     "plain_text",
							},
							"type": "danger",
							"url":  "#",
						},
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "📞 Contact Support",
								"tag":     "plain_text",
							},
							"type": "default",
							"url":  "#",
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎯 Received webhook from GitHub")

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		return
	}

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		log.Printf("❌ Error decoding webhook: %v", err)
		sendErrorNotification(err.Error(), "webhook-handler")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	commitMessages := ""
	for _, commit := range pushEvent.Commits {
		commitMessages += "• " + commit.Message + "\n"
	}

	info := GitCommitInfo{
		Message:     commitMessages,
		Environment: "DEV",
		ServiceName: pushEvent.Repository.Name,
		Deployer:    pushEvent.Commits[0].Author.Name,
		RepoURL:     "https://github.com/kunaaa123/Bot_Test",
	}

	if err := sendGitDeploymentToLark(info); err != nil {
		log.Printf("❌ Error sending to Lark: %v", err)
		sendErrorNotification(err.Error(), info.ServiceName)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Webhook processed successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("🎉 Webhook processed successfully"))
}

func main() {
	http.HandleFunc("/deployment-info", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "DEV",
			Deployer:    "rutchanai",
			ServiceName: "tgth-backend-main",
			CommitMsg:   "feat: enhanced beautiful notification system with emojis and colors",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		if err := sendToLark(info); err != nil {
			log.Printf("❌ Failed to send to Lark: %v", err)
			sendErrorNotification(err.Error(), info.ServiceName)
		} else {
			log.Printf("✅ Deployment notification sent successfully")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	http.HandleFunc("/test-notification", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info := DeploymentInfo{
			ENV:         "TEST",
			Deployer:    "test-user",
			ServiceName: "notification-service",
			CommitMsg:   "test: beautiful notification system with enhanced UI/UX",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		if err := sendToLark(info); err != nil {
			log.Printf("❌ Error sending to Lark: %v", err)
			sendErrorNotification(err.Error(), info.ServiceName)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "🎉 Beautiful notification sent successfully!",
		})
	})

	http.HandleFunc("/custom-notification", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var info DeploymentInfo
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := sendToLark(info); err != nil {
			log.Printf("❌ Error sending to Lark: %v", err)
			sendErrorNotification(err.Error(), info.ServiceName)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "✨ Custom beautiful notification sent successfully!",
		})
	})

	// New endpoint for testing error notifications
	http.HandleFunc("/test-error", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sendErrorNotification("Database connection timeout after 30 seconds", "test-service")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "🚨 Error notification sent successfully!",
		})
	})

	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Beautiful Notification Server running on port %s\n", port)
	fmt.Printf("📋 Available endpoints:\n")
	fmt.Printf("   • POST /deployment-info - Send deployment notification\n")
	fmt.Printf("   • POST /test-notification - Send test notification\n")
	fmt.Printf("   • POST /custom-notification - Send custom notification\n")
	fmt.Printf("   • POST /test-error - Send error notification\n")
	fmt.Printf("   • POST /git-webhook - GitHub webhook handler\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
