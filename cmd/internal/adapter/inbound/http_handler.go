package inbound

import (
	"encoding/json"
	"larkbot/cmd/internal/app"
	"larkbot/cmd/internal/domain"
	"log"
	"net/http"
)

type WebhookHandler struct {
	Usecase *app.WebhookUsecase
}

func NewWebhookHandler(usecase *app.WebhookUsecase) *WebhookHandler {
	return &WebhookHandler{Usecase: usecase}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received webhook request from GitHub")

	var event domain.GitHubPushEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Repository: %s, Commits count: %d", event.Repository.Name, len(event.Commits))

	if err := h.Usecase.HandleGitHubPush(event); err != nil {
		log.Printf("Error handling webhook: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
