package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type DeploymentInfo struct {
	ENV         string `json:"env"`
	Deployer    string `json:"deployer"`
	ServiceName string `json:"serviceName"`
	CommitMsg   string `json:"commitMsg"`
}

type GitCommitInfo struct {
	Message     string   `json:"message"`
	Features    []string `json:"features"`
	Environment string   `json:"environment"`
	ServiceName string   `json:"service_name"`
	Deployer    string   `json:"deployer"`
}

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Commits []struct {
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

func sendToLark(info DeploymentInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f5" // ใส่ URL ของ Lark webhook ที่นี่

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "📊 Frontend Deployment",
					"tag":     "plain_text",
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**ENV:** %s\n**Deployer:** %s\n**Service:** %s",
							info.ENV, info.Deployer, info.ServiceName),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Commit Messages:**\n%s", info.CommitMsg),
						"tag":     "lark_md",
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error sending to Lark: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("Lark Response: %s", string(body))
	return nil
}

func sendGitDeploymentToLark(commit GitCommitInfo) error {
	webhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/66a2d4a9-a7dd-47d3-a15a-c11c6f97c7f"

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"content": "Frontend Deployment",
					"tag":     "plain_text",
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**🌟 ENV**\n%s\n\n**👨‍💻 Deployer**\n%s\n\n**🚀 Service**\n%s",
							commit.Environment, commit.Deployer, commit.ServiceName),
						"tag": "lark_md",
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"content": "**📝 Commit Messages**\n" + commit.Message,
						"tag":     "lark_md",
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// เพิ่มการอ่าน response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("Lark Response: %s", string(body))
	return nil
}

func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received webhook from GitHub")

	// ตรวจสอบ event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		return
	}

	// อ่านและ parse ข้อมูล
	var pushEvent GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&pushEvent); err != nil {
		log.Printf("Error decoding webhook: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// สร้างข้อมูลสำหรับส่งไป Lark
	commitMessages := ""
	for _, commit := range pushEvent.Commits {
		commitMessages += "- " + commit.Message + "\n"
	}

	info := GitCommitInfo{
		Message:     commitMessages,
		Environment: "DEV",
		ServiceName: pushEvent.Repository.Name,
		Deployer:    pushEvent.Commits[0].Author.Name,
	}

	// ส่งไป Lark
	if err := sendGitDeploymentToLark(info); err != nil {
		log.Printf("Error sending to Lark: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook processed successfully"))
}

func main() {
	http.HandleFunc("/deployment-info", func(w http.ResponseWriter, r *http.Request) {
		info := DeploymentInfo{
			ENV:         "DEV",
			Deployer:    "rutchanai",
			ServiceName: "tgth-backend-main",
			CommitMsg:   "feat: backend deployment structure",
		}

		// ส่งข้อมูลไปยัง Lark
		if err := sendToLark(info); err != nil {
			log.Printf("Failed to send to Lark: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// เพิ่ม endpoint สำหรับทดสอบ
	http.HandleFunc("/test-notification", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info := DeploymentInfo{
			ENV:         "TEST",
			Deployer:    "test-user",
			ServiceName: "test-service",
			CommitMsg:   "test: testing notification system",
		}

		if err := sendToLark(info); err != nil {
			log.Printf("Error sending to Lark: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "notification sent successfully",
		})
	})

	// เพิ่ม endpoint สำหรับทดสอบแบบกำหนดข้อมูลเอง
	http.HandleFunc("/custom-notification", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var info DeploymentInfo
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := sendToLark(info); err != nil {
			log.Printf("Error sending to Lark: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "custom notification sent successfully",
		})
	})

	http.HandleFunc("/git-webhook", handleGitHubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// หลังจากตั้งค่า authtoken แล้ว ลองรัน
// ngrok http 8080
