package app

import (
	"fmt"
	"larkbot/cmd/internal/domain"
	"strings"
)

type WebhookUsecase struct {
	Lark domain.LarkService
}

func NewWebhookUsecase(lark domain.LarkService) *WebhookUsecase {
	return &WebhookUsecase{Lark: lark}
}

func getBranchEnvironment(branch string) string {
	switch branch {
	case "main", "master":
		return "PRODUCTION"
	case "staging", "qa":
		return "STAGING"
	case "develop":
		return "DEVELOPMENT"
	default:
		if strings.HasPrefix(branch, "feature/") {
			return "DEVELOPMENT"
		}
		if strings.HasPrefix(branch, "hotfix/") {
			return "HOTFIX"
		}
		return "DEVELOPMENT"
	}
}

func (u *WebhookUsecase) HandleGitHubPush(event domain.GitHubPushEvent) error {
	if len(event.Commits) == 0 {
		return nil
	}

	token, err := u.Lark.GetTenantAccessToken()
	if err != nil {
		return err
	}

	imageKey, err := u.Lark.UploadImage("c:/Users/Singha/Desktop/Larkbot/github_logo.png", token) // แก้ไขให้ใช้ path ไม่ต้องการให้รูป fix
	if err != nil {
		return err
	}

	lastCommit := event.Commits[0]

	// แปลง ref (refs/heads/main) เป็นชื่อ branch (main)
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")
	env := getBranchEnvironment(branch)

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": "Backend Deployment", // แก้ไขให้แสดงแค่ Backend Deployment
				},
				"template": "blue",
			},
			"elements": []map[string]interface{}{
				{
					"tag":     "img",
					"img_key": imageKey,
					"mode":    "fit_horizontal",
					"preview": true,
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**🌿 Branch**\n`%s`", branch),
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**🤖 Deployer**\n%s", lastCommit.Author.Name),
							},
						},
					},
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**🌐 ENV**\n`%s`", env),
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("**📦 Service**\n`%s`", event.Repository.Name),
							},
						},
					},
				},
				{
					"tag": "hr",
				},
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**Commit Messages** 🤔\n• %s", lastCommit.Message),
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
							"url":  event.Repository.HTMLURL, // ใช้ URL ที่ GitHub
							"type": "default",
						},
					},
				},
			},
		},
	}

	return u.Lark.SendWebhookMessage(payload)
}
