package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	LARK_WEBHOOK = "https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff"
	APP_ID       = "cli_a8b2c70af7389029"
	APP_SECRET   = "QUbHQALAU0xrxWid9QU8Hb50wpY1wtwv"
)

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

func getTenantAccessToken() string {
	payload := map[string]string{
		"app_id":     APP_ID,
		"app_secret": APP_SECRET,
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, _ := http.Post("https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json", bytes.NewBuffer(payloadBytes))
	defer resp.Body.Close()

	var result struct {
		TenantAccessToken string `json:"tenant_access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.TenantAccessToken
}

func uploadImageToLark(filePath, token string) string {
	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("image_type", "message")
	part, _ := writer.CreateFormFile("image", filepath.Base(filePath))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://open.larksuite.com/open-apis/im/v1/images", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.ImageKey
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	var pushEvent GitHubPushEvent
	json.NewDecoder(r.Body).Decode(&pushEvent)

	if len(pushEvent.Commits) > 0 {
		token := getTenantAccessToken()
		imageKey := uploadImageToLark("./1_dDNpLKu_oTLzStsDTnkJ-g.png", token)
		lastCommit := pushEvent.Commits[0]

		payload := map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"header": map[string]interface{}{
					"title": map[string]interface{}{
						"tag":     "plain_text",
						"content": "Backend Deployment",
					},
					"template": "indigo",
				},
				"elements": []map[string]interface{}{
					{
						"tag":     "img",
						"img_key": imageKey,
						"mode":    "fit_horizontal",
					},
					{
						"fields": []map[string]interface{}{
							{
								"is_short": true,
								"text": map[string]interface{}{
									"tag":     "lark_md",
									"content": fmt.Sprintf("<center><div style=\"text-align: center; border: 1px solid #e0e0e0; padding: 8px; border-radius: 5px; background: rgba(0,0,0,0.05);\">**ENV**\nDEV</div></center>"),
								},
							},
							{
								"is_short": true,
								"text": map[string]interface{}{
									"tag":     "lark_md",
									"content": fmt.Sprintf("<center><div style=\"text-align: center; border: 1px solid #e0e0e0; padding: 8px; border-radius: 5px; background: rgba(0,0,0,0.05);\">**🤖 Deployer**\n%s</div></center>", lastCommit.Author.Name),
								},
							},
						},
					},
					{
						"tag": "div",
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": fmt.Sprintf("<center>**Service Name**\n%s</center>", pushEvent.Repository.Name),
						},
					},
					{
						"tag": "hr",
					},
					{
						"tag": "div",
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": fmt.Sprintf("<center>**Commit Messages** 🤔\n• %s</center>", lastCommit.Message),
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
								"url":  fmt.Sprintf("https://github.com/%s", pushEvent.Repository.Name),
								"type": "default",
							},
						},
					},
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		http.Post(LARK_WEBHOOK, "application/json", bytes.NewBuffer(payloadBytes))
	}
}

func main() {
	http.HandleFunc("/git-webhook", handleGitHubWebhook)
	http.ListenAndServe(":8080", nil)
}

//ee
//
