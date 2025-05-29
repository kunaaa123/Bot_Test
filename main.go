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
					"content": "🔔 แจ้งเตือน Commit ใหม่",
					"tag":     "plain_text",
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag":  "img",
					"src":  "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png", // ตัวอย่าง URL รูปภาพ
					"mode": "fit_horizontal",
					"alt":  "GitHub Logo",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("### 📌 รายละเอียด Commit\n\n"+
							"**🏢 Repository:** %s\n"+
							"**👤 ผู้ทำการ Commit:** %s\n"+
							"**📝 ข้อความ:** %s",
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
		return fmt.Errorf("failed to send to Lark: %v", err)
	}
	defer resp.Body.Close()

	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)
	log.Printf("Lark Response: %s", respBody.String()) // เพิ่ม log นี้

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lark API returned non-200 status code: %d, body: %s",
			resp.StatusCode, respBody.String())
	}
	return nil
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		log.Printf("Failed to decode webhook payload: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(pushEvent.Commits) > 0 {
		log.Printf("Sending notification for commit by %s in repo %s",
			pushEvent.Commits[0].Author.Name,
			pushEvent.Repository.Name)

		err := sendToLark(
			pushEvent.Commits[0].Message,
			pushEvent.Repository.Name,
			pushEvent.Commits[0].Author.Name,
		)
		if err != nil {
			log.Printf("Failed to send to Lark: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Successfully sent notification to Lark")
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/git-webhook", handleGitHubWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

////
