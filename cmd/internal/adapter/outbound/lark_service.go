package outbound

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

type Lark struct {
	Webhook string
	AppID   string
	Secret  string
}

func NewLark(webhook, appID, secret string) *Lark {
	return &Lark{Webhook: webhook, AppID: appID, Secret: secret}
}

func (l *Lark) GetTenantAccessToken() (string, error) {
	payload := map[string]string{
		"app_id":     l.AppID,
		"app_secret": l.Secret,
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post("https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		TenantAccessToken string `json:"tenant_access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.TenantAccessToken, err
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
	resp, err := http.Post(l.Webhook, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	// อ่าน response body เพื่อตรวจสอบ error
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
