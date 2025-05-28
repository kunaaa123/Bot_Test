package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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

func sendToLark(info DeploymentInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "🚀 Backend Deployment Status",
					"tag":     "plain_text",
				},
				"template": "indigo",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("🔹 **Environment:** %s\n👤 **Deployer:** %s\n🌟 **Service:** %s",
							info.ENV, info.Deployer, info.ServiceName),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": "📝 **Latest Changes:**\n" + info.CommitMsg,
						"tag":     "lark_md",
					},
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("🕒 Deployed at: %s", time.Now().Format("2006-01-02 15:04:05")),
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

	return nil
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	commitMsg := ""
	for _, commit := range pushEvent.Commits {
		commitMsg += "- " + commit.Message + "\n"
	}

	info := DeploymentInfo{
		ENV:         "DEV",
		Deployer:    pushEvent.Commits[0].Author.Name,
		ServiceName: pushEvent.Repository.Name,
		CommitMsg:   commitMsg,
	}

	if err := sendToLark(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
