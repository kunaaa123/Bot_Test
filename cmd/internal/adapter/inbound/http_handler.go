package inbound

import (
	"encoding/json"
	"larkbot/cmd/internal/app"
	"larkbot/cmd/internal/domain"
	"net/http"
)

type WebhookHandler struct {
	Usecase *app.WebhookUsecase
}

func NewWebhookHandler(usecase *app.WebhookUsecase) *WebhookHandler {
	return &WebhookHandler{Usecase: usecase}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var event domain.GitHubPushEvent
	json.NewDecoder(r.Body).Decode(&event)
	err := h.Usecase.HandleGitHubPush(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
