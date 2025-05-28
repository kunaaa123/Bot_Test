package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"` // เพิ่ม field นี้
}

type GitHubPushEvent struct {
	Repository struct {
		Name string `json:"name"`
		URL  string `json:"html_url"` // เพิ่ม field นี้
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
					"content": "⚡ Backend Deployment",
					"tag":     "plain_text",
				},
				"template": "indigo",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("🚀 **ENV**\n%s", info.ENV),
								"tag":     "lark_md",
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("👨‍💻 **Deployer**\n%s", info.Deployer),
								"tag":     "lark_md",
							},
						},
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("🔧 **Service**\n%s", info.ServiceName),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("📝 **Commit Messages**\n%s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
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
							"url":  info.RepoURL, // ใช้ URL จริงจาก GitHub
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
		RepoURL:     pushEvent.Repository.URL, // เพิ่ม URL
	}

	if err := sendToLark(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
