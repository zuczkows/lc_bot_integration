package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type WebhookHandler struct {
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

type Webhook struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func (h *WebhookHandler) Reply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	var webhook Webhook

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("Error decoding webhook JSON: %v", err)
		return
	}

	defer r.Body.Close()

	log.Printf("Received webhook: %s, payload %s", webhook.Action, string(webhook.Payload))
}
