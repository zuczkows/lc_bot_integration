package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/zuczkows/text-bot-integration/internal/store"
)

type WebhookHandler struct {
	botStore *store.BotStore
}

func NewWebhookHandler(botStore *store.BotStore) *WebhookHandler {
	return &WebhookHandler{
		botStore: botStore,
	}
}

type Webhook struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type Chat struct {
	ID string `json:"id"`
}

type IncomingChat struct {
	Chat Chat `json:"chat"`
}

// copy paste from sdk
type eventSpecific struct {
	Text              json.RawMessage `json:"text"`
	TextVars          json.RawMessage `json:"text_vars"`
	Fields            json.RawMessage `json:"fields"`
	FormType          json.RawMessage `json:"form_type"`
	ContentType       json.RawMessage `json:"content_type"`
	Name              json.RawMessage `json:"name"`
	URL               json.RawMessage `json:"url"`
	ThumbnailURL      json.RawMessage `json:"thumbnail_url"`
	Thumbnail2xURL    json.RawMessage `json:"thumbnail2x_url"`
	Width             json.RawMessage `json:"width"`
	Height            json.RawMessage `json:"height"`
	Size              json.RawMessage `json:"size"`
	TemplateID        json.RawMessage `json:"template_id"`
	Elements          json.RawMessage `json:"elements"`
	Postback          json.RawMessage `json:"postback"`
	AlternativeText   json.RawMessage `json:"alternative_text"`
	SystemMessageType json.RawMessage `json:"system_message_type"`
	Source            json.RawMessage `json:"source"`
	Subtype           json.RawMessage `json:"subtype"`
	Details           json.RawMessage `json:"details"`
	Version           json.RawMessage `json:"version"`
}

type Event struct {
	ID         string    `json:"id,omitempty"`
	CustomID   string    `json:"custom_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	AuthorID   string    `json:"author_id"`
	Visibility string    `json:"visibility,omitempty"`
	Type       string    `json:"type,omitempty"`
	eventSpecific
}

type IncomingEvent struct {
	ChatID   string `json:"chat_id"`
	ThreadID string `json:"thread_id"`
	Event    Event  `json:"event"`
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

	switch {
	case webhook.Action == "incoming_event":
		log.Printf("Incoming envet action")
	case webhook.Action == "incoming_chat":
		log.Printf("Incoming chat action")
	default:
		log.Printf("Webhook not recognized: %v", webhook)
	}

}
