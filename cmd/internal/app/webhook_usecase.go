package app

import (
	"fmt"
	"larkbot/cmd/internal/domain"
)

type WebhookUsecase struct {
	Lark domain.LarkService
}

func NewWebhookUsecase(lark domain.LarkService) *WebhookUsecase {
	return &WebhookUsecase{Lark: lark}
}

func (u *WebhookUsecase) HandleGitHubPush(event domain.GitHubPushEvent) error {
	if len(event.Commits) == 0 {
		return nil
	}

	token, err := u.Lark.GetTenantAccessToken()
	if err != nil {
		return err
	}

	imageKey, err := u.Lark.UploadImage("c:/Users/Singha/Desktop/Larkbot/github_logo.png", token)
	if err != nil {
		return err
	}

	lastCommit := event.Commits[0]

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "🚀 Deployment Notification",
				},
				"template": "blue", // เปลี่ยนสีให้โดดเด่น
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": "━━━━━━━━━━━━━━━━━━━━━━", // เส้นคั่นด้านบน
					},
				},
				{
					"tag":     "img",
					"img_key": imageKey,
					"mode":    "crop_center", // จัดรูปให้อยู่กลาง
					"preview": true,
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**🌟 Service: %s**", event.Repository.Name),
					},
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": "**📍 Environment**\n`DEV`",
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**👨‍💻 Deployer**\n`%s`", lastCommit.Author.Name),
							},
						},
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": "━━━━━━━━━━━━━━━━━━━━━━", // เส้นคั่นกลาง
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**📝 Latest Commit**\n> %s", lastCommit.Message),
					},
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": "━━━━━━━━━━━━━━━━━━━━━━", // เส้นคั่นล่าง
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"content": "🔍 View Repository",
								"tag":     "plain_text",
							},
							"url":  event.Repository.HTMLURL,
							"type": "primary", // ปรับสีปุ่มให้โดดเด่น
						},
					},
				},
			},
			"config": map[string]interface{}{
				"wide_screen_mode": true, // แสดงแบบเต็มจอ
			},
		},
	}

	return u.Lark.SendWebhookMessage(payload)
}
