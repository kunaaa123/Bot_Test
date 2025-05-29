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
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"` // เพิ่มฟิลด์นี้เพื่อรับ avatar
		} `json:"author"`
	} `json:"commits"`
	Sender struct {
		AvatarURL string `json:"avatar_url"` // avatar ของผู้ push
	} `json:"sender"`
}

func sendToLark(message, repo, author, avatarURL string) error {
	// สร้าง elements slice
	elements := []map[string]interface{}{
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
	}

	// เพิ่มรูปภาพถ้ามี URL
	if avatarURL != "" {
		imageElement := map[string]interface{}{
			"tag":     "img",
			"img_key": avatarURL, // ใช้ URL โดยตรง
			"alt": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("Avatar ของ %s", author),
			},
		}
		elements = append(elements, imageElement)
	}

	// เพิ่มรูปภาพสถานะ (ตัวอย่าง)
	statusImageURL := "https://img.icons8.com/color/48/git.png" // ตัวอย่าง Git icon
	statusElement := map[string]interface{}{
		"tag":     "img",
		"img_key": statusImageURL,
		"alt": map[string]interface{}{
			"tag":     "plain_text",
			"content": "Git Status Icon",
		},
	}
	elements = append(elements, statusElement)

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
			"elements": elements,
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
	log.Printf("Lark Response: %s", respBody.String())

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lark API returned non-200 status code: %d, body: %s",
			resp.StatusCode, respBody.String())
	}
	return nil
}

// ฟังก์ชันสำหรับส่งข้อความพร้อมรูปภาพแบบกำหนดเอง
func sendToLarkWithCustomImage(message, repo, author, imageURL, statusIcon string) error {
	elements := []map[string]interface{}{
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
	}

	// เพิ่มรูปภาพหลัก
	if imageURL != "" {
		imageElement := map[string]interface{}{
			"tag":     "img",
			"img_key": imageURL,
			"alt": map[string]interface{}{
				"tag":     "plain_text",
				"content": "Custom Image",
			},
		}
		elements = append(elements, imageElement)
	}

	// เพิ่มไอคอนสถานะ
	if statusIcon != "" {
		statusElement := map[string]interface{}{
			"tag":     "img",
			"img_key": statusIcon,
			"alt": map[string]interface{}{
				"tag":     "plain_text",
				"content": "Status Icon",
			},
		}
		elements = append(elements, statusElement)
	}

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "🔔 แจ้งเตือน Commit ใหม่",
					"tag":     "plain_text",
				},
				"template": "green", // เปลี่ยนเป็นสีเขียว
			},
			"elements": elements,
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
	log.Printf("Lark Response: %s", respBody.String())

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

		// ใช้ avatar จาก sender หรือ commit author
		avatarURL := pushEvent.Sender.AvatarURL
		if avatarURL == "" && len(pushEvent.Commits) > 0 {
			avatarURL = pushEvent.Commits[0].Author.AvatarURL
		}

		// ตัวอย่างการใช้ฟังก์ชันแบบกำหนดเองพร้อมรูปภาพ
		customImageURL := "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
		statusIcon := "https://img.icons8.com/color/48/checkmark.png"

		err := sendToLarkWithCustomImage(
			pushEvent.Commits[0].Message,
			pushEvent.Repository.Name,
			pushEvent.Commits[0].Author.Name,
			customImageURL, // รูปภาพหลัก
			statusIcon,     // ไอคอนสถานะ
		)

		// หรือใช้แบบเดิมพร้อม avatar
		// err := sendToLark(
		// 	pushEvent.Commits[0].Message,
		// 	pushEvent.Repository.Name,
		// 	pushEvent.Commits[0].Author.Name,
		// 	avatarURL,
		// )

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
