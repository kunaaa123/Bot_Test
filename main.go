package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const LARK_WEBHOOK = "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"

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

func sendToLark(message, repo, author string) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "🚀 New Commit Notification",
					"tag":     "plain_text",
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag":     "img",
					"img_key": "https://sourcebae.com/blog/wp-content/uploads/2023/09/maxresdefault-44.jpg",
					"alt": map[string]interface{}{
						"tag":     "plain_text",
						"content": "Git Commit Image",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("📦 *Repository:* %s\n👨‍💻 *Author:* %s\n💬 *Message:* %s",
							repo, author, message),
						"tag": "lark_md",
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

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(pushEvent.Commits) > 0 {
		err := sendToLark(
			pushEvent.Commits[0].Message,
			pushEvent.Repository.Name,
			pushEvent.Commits[0].Author.Name,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/git-webhook", handleGitHubWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//
