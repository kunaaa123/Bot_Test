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
	"path/filepath"
)

const (
	LARK_WEBHOOK     = "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5"
	APP_ID           = "cli_a8b2c70af7389029"             // แทนที่ด้วย app_id จริง
	APP_SECRET       = "QUbHQALAU0xrxWid9QU8Hb50wpY1wtwv" // แทนที่ด้วย app_secret จริง
	IMAGE_UPLOAD_URL = "https://open.larksuite.com/open-apis/im/v1/images"
	TOKEN_URL        = "https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal"
)

// โครงสร้างสำหรับ GitHub Push Event
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

// ฟังก์ชันรับ tenant_access_token
func getTenantAccessToken() (string, error) {
	payload := map[string]string{
		"app_id":     APP_ID,
		"app_secret": APP_SECRET,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(TOKEN_URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code              int    `json:"code"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("failed to get token, code: %d", result.Code)
	}
	return result.TenantAccessToken, nil
}

// ฟังก์ชันอัปโหลดรูปภาพและรับ image_key
func uploadImageToLark(filePath, token string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequest("POST", IMAGE_UPLOAD_URL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("upload failed, code: %d", result.Code)
	}
	return result.Data.ImageKey, nil
}

// ฟังก์ชันส่งข้อความไปยัง Lark พร้อม image_key
func sendToLark(message, repo, author, imageKey string) error {
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("Repository: %s\nAuthor: %s\nMessage: %s",
							repo, author, message),
						"tag": "plain_text",
					},
				},
				{
					"tag":     "img",
					"img_key": imageKey,
					"alt": map[string]interface{}{
						"tag":     "plain_text",
						"content": "Image from GitHub webhook",
					},
				},
			},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "GitHub Webhook Notification",
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send to Lark, status: %d", resp.StatusCode)
	}
	return nil
}

// ฟังก์ชันจัดการ GitHub Webhook
func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// ตั้งค่า header
	w.Header().Set("Content-Type", "application/json")

	// ตรวจสอบ Content-Type ของ request
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// ตรวจสอบว่าเป็น POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(pushEvent.Commits) > 0 {
		// รับ tenant_access_token
		token, err := getTenantAccessToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// อัปโหลดรูปภาพ (แทนที่ด้วย path รูปภาพจริง)
		imagePath := "./Screenshot 2025-05-28 171410.png"
		imageKey, err := uploadImageToLark(imagePath, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ส่งข้อความไปยัง Lark
		lastCommit := pushEvent.Commits[0]
		err = sendToLark(
			lastCommit.Message,
			pushEvent.Repository.Name,
			lastCommit.Author.Name,
			imageKey,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ส่ง response กลับในรูปแบบ JSON
		response := map[string]string{
			"status":  "success",
			"message": "Webhook processed successfully",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]string{
			"status":  "success",
			"message": "No commits found in the push event",
		}
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	http.HandleFunc("/webhook", handleGitHubWebhook)

	port := ":8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

//
