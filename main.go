package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	LARK_WEBHOOK = "https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff"
	APP_ID       = "cli_a8b2c70af7389029"
	APP_SECRET   = "QUbHQALAU0xrxWid9QU8Hb50wpY1wtwv"
)

type TokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TenantAccessToken string `json:"tenant_access_token"`
	} `json:"data"`
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

func sendLarkNotification(pushEvent *GitHubPushEvent, imageKey string) error {
	lastCommit := pushEvent.Commits[0]
	card := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title":    map[string]string{"tag": "plain_text", "content": "Backend Deployment"},
				"template": "indigo",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "img", "img_key": imageKey,
					"mode": "fit_horizontal", "preview": true,
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag": "lark_md",
						"content": fmt.Sprintf("**Service:** %s\n**Deployer:** %s\n**Message:** %s",
							pushEvent.Repository.Name, lastCommit.Author.Name, lastCommit.Message),
					},
				},
			},
		},
	}

	payload, err := json.Marshal(card)
	if err != nil {
		return err
	}

	resp, err := http.Post(LARK_WEBHOOK, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func uploadImage(token string) (string, error) {
	file, err := os.Open("./github_logo.png")
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("image_type", "message")
	part, _ := writer.CreateFormFile("image", "github_logo.png")
	io.Copy(part, file)
	writer.Close()

	req, err := http.NewRequest("POST", "https://open.larksuite.com/open-apis/im/v1/images", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.ImageKey, nil
}

func getTenantAccessToken() string {
	data := map[string]string{
		"app_id":     APP_ID,
		"app_secret": APP_SECRET,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal token request: %v", err)
		return ""
	}

	resp, err := http.Post(
		"https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		return ""
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Printf("Failed to decode token response: %v", err)
		return ""
	}

	return tokenResp.Data.TenantAccessToken
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Log request details
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}
	log.Printf("Received webhook payload: %s", string(body))

	// Create new reader since body was read
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		log.Printf("Error decoding payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(pushEvent.Commits) == 0 {
		log.Println("No commits found in payload")
		w.WriteHeader(http.StatusOK)
		return
	}

	imageKey, err := uploadImage(getTenantAccessToken())
	if err != nil {
		log.Printf("Failed to upload image: %v", err)
		http.Error(w, "Failed to process webhook", http.StatusInternalServerError)
		return
	}

	if err := sendLarkNotification(&pushEvent, imageKey); err != nil {
		log.Printf("Failed to send notification: %v", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//
