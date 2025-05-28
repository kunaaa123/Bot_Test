package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	LARK_WEBHOOK = "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"
)

type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
	RepoURL     string `json:"repoUrl"`
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

func sendCleanNotification(info DeploymentInfo) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": fmt.Sprintf("Frontend Deployment"),
					"tag":     "plain_text",
				},
				"template": getEnvironmentColor(info.ENV),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Env**\n%s\n\n**Service Name**\n%s\n\n**Deployer**\n%s",
							info.ENV, info.ServiceName, info.Deployer),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Commit Messages**\n%s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "View repo",
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

	resp, err := http.Post(LARK_WEBHOOK, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func main() {
	http.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "DEV",
			Deployer:    "rutchanai",
			ServiceName: "tgth-adminweb-banner",
			CommitMsg:   "• fix(redirect-type): ota web view custom tab",
			RepoURL:     "https://github.com/kunaaa123/Bot_Test",
		}

		if err := sendCleanNotification(info); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "notification_sent"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
