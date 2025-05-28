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
	webhookURL := "YOUR_LARK_WEBHOOK_URL"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "Backend Deployment",
					"tag":     "plain_text",
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Environment: %s\nDeployer: %s\nService: %s",
							info.ENV, info.Deployer, info.ServiceName),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Changes:\n%s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "View Repository",
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
		RepoURL:     "YOUR_REPO_URL",
	}

	if err := sendToLark(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/git-webhook", handleGitHubWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//test2
//220
