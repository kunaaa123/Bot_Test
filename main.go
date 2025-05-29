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
	LARK_WEBHOOK     = "https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff"
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

	// เพิ่ม logging
	log.Printf("Getting token for app_id: %s", APP_ID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(TOKEN_URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// เพิ่ม logging response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Token Response: %s", string(respBody))

	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

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
	log.Printf("Starting image upload: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// เพิ่มฟิลด์ image_type
	err = writer.WriteField("image_type", "message")
	if err != nil {
		return "", err
	}

	// สร้าง form file ด้วยชื่อฟิลด์ "image"
	part, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		log.Printf("Error creating form file: %v", err)
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Printf("Error copying file: %v", err)
		return "", err
	}
	writer.Close()

	req, err := http.NewRequest("POST", IMAGE_UPLOAD_URL, body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", err
	}

	// เพิ่ม headers ที่จำเป็น
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// เพิ่ม logging
	log.Printf("Request URL: %s", IMAGE_UPLOAD_URL)
	log.Printf("Authorization: Bearer %s", token)
	log.Printf("Content-Type: %s", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// เพิ่ม detailed logging
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Full Response Status: %d", resp.StatusCode)
	log.Printf("Full Response Headers: %v", resp.Header)
	log.Printf("Full Response Body: %s", string(respBody))

	// Reset response body for further reading
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	var result struct {
		Code int `json:"code"`
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
		Msg string `json:"msg"` // เพิ่มเพื่อดู error message
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", err
	}
	if result.Code != 0 {
		log.Printf("Upload failed with code %d: %s", result.Code, result.Msg)
		return "", fmt.Errorf("upload failed, code: %d, msg: %s", result.Code, result.Msg)
	}

	log.Printf("Upload successful, image key: %s", result.Data.ImageKey)
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
			log.Printf("Error getting token: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ตรวจสอบว่าไฟล์รูปภาพมีอยู่จริง
		imagePath := "./1_dDNpLKu_oTLzStsDTnkJ-g.png"
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			log.Printf("Image file not found: %v", err)
			http.Error(w, "Image file not found", http.StatusInternalServerError)
			return
		}

		// อัปโหลดรูปภาพ
		imageKey, err := uploadImageToLark(imagePath, token)
		if err != nil {
			log.Printf("Error uploading image: %v", err)
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
	// เปลี่ยนจาก /webhook เป็น /git-webhook
	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	port := ":8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

//ee
//
