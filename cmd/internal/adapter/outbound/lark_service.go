package outbound

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

type Lark struct {
	Webhook string
	AppID   string
	Secret  string
}

func NewLark(webhook, appID, secret string) *Lark {
	return &Lark{Webhook: webhook, AppID: appID, Secret: secret}
}

func (l *Lark) GetTenantAccessToken() (string, error) {
	log.Printf("Getting tenant access token")
	payload := map[string]string{
		"app_id":     l.AppID,
		"app_secret": l.Secret,
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post("https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Token response: %s", string(body))

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing token response: %v", err)
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("failed to get token: %s", result.Msg)
	}

	return result.TenantAccessToken, nil
}

func (l *Lark) UploadImage(filePath string, token string) (string, error) {
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
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.ImageKey, err
}

func (l *Lark) SendWebhookMessage(payload any) error {
	payloadBytes, _ := json.Marshal(payload)
	log.Printf("Sending webhook message to Lark: %s", string(payloadBytes))

	resp, err := http.Post(l.Webhook, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error sending webhook: %v", err)
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Lark response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
