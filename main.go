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
)

const webhookURL = "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

// Core data structures
type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"`
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

// Helper functions
func getEnvironmentStyle(env string) (color string, emoji string) {
	env = strings.ToUpper(env)
	switch env {
	case "PROD", "PRODUCTION":
		return "red", "🚀"
	case "STAGING", "STAGE":
		return "orange", "🔄"
	case "DEV", "DEVELOPMENT":
		return "blue", "⚡"
	case "TEST":
		return "green", "🧪"
	default:
		return "grey", "📦"
	}
}

// Core notification functions
func sendLarkNotification(title string, color string, elements []map[string]interface{}) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title":    map[string]interface{}{"content": title, "tag": "plain_text"},
				"template": color,
			},
			"elements": elements,
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

	if _, err := io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	return nil
}

func handleDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info DeploymentInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	color, emoji := getEnvironmentStyle(info.ENV)
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**Environment:** `%s`\n**Deployer:** `%s`\n**Service:** `%s`",
					info.ENV, info.Deployer, info.ServiceName),
				"tag": "lark_md",
			},
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**Changes:**\n```\n%s\n```", info.CommitMsg),
				"tag":     "lark_md",
			},
		},
		{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag":  "button",
					"text": map[string]interface{}{"content": "View Repository", "tag": "plain_text"},
					"type": "primary",
					"url":  info.RepoURL,
				},
			},
		},
	}

	if err := sendLarkNotification(fmt.Sprintf("%s Deployment Update", emoji), color, elements); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

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

	commitMessages := ""
	for _, commit := range pushEvent.Commits {
		commitMessages += "• " + commit.Message + "\n"
	}

	color, emoji := getEnvironmentStyle("DEV")
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**Repository:** `%s`\n**Author:** `%s`\n**Commits:**\n```\n%s```",
					pushEvent.Repository.Name,
					pushEvent.Commits[0].Author.Name,
					commitMessages),
				"tag": "lark_md",
			},
		},
	}

	if err := sendLarkNotification(fmt.Sprintf("%s New Commits", emoji), color, elements); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/deploy", handleDeployment)
	http.HandleFunc("/webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
