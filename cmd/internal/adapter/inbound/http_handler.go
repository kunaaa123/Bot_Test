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
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Usecase.HandleGitHubPush(event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
