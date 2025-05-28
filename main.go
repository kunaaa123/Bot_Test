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

func sendToLark(info DeploymentInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "Deployment Notification",
					"tag":     "plain_text",
				},
				"template": getEnvironmentColor(info.ENV),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Environment: %s\nService: %s\nDeployer: %s\nTime: %s",
							info.ENV, info.ServiceName, info.Deployer, time.Now().Format("2006-01-02 15:04:05")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Changes: %s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func sendGitDeploymentToLark(commit GitCommitInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("Git Deploy - %s", commit.Environment),
					"tag":     "plain_text",
				},
				"template": getEnvironmentColor(commit.Environment),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Service: %s\nDeveloper: %s\nTime: %s",
							commit.ServiceName, commit.Deployer, time.Now().Format("2006-01-02 15:04:05")),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Commits:\n%s", strings.TrimSpace(commit.Message)),
						"tag":     "lark_md",
					},
				},
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
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
		commitMessages += commit.Message + "\n"
	}

	info := GitCommitInfo{
		Message:     commitMessages,
		Environment: "DEV",
		ServiceName: pushEvent.Repository.Name,
		Deployer:    pushEvent.Commits[0].Author.Name,
		RepoURL:     "https://github.com/kunaaa123/Bot_Test",
	}

	if err := sendGitDeploymentToLark(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/deployment-info", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "DEV",
			Deployer:    "rutchanai",
			ServiceName: "tgth-backend-main",
			CommitMsg:   "Latest deployment",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		sendToLark(info)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
