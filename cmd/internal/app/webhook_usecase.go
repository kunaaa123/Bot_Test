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
					"content": "Backend Deployment",
				},
				"template": "indigo",
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
								"content": "**ENV**\nDEV",
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
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": fmt.Sprintf("**Service Name**\n%s", event.Repository.Name),
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
							"url":  fmt.Sprintf("https://github.com/%s", event.Repository.Name),
							"type": "default",
						},
					},
				},
			},
		},
	}

	return u.Lark.SendWebhookMessage(payload)
}
