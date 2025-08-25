package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
	"github.com/zuczkows/text-bot-integration/internal/store"
)

type WebhookHandler struct {
	botStore    *store.BotStore
	livechatSDK *sdk.LivechatSDKClient
}

func NewWebhookHandler(botStore *store.BotStore, liveChatSDK *sdk.LivechatSDKClient) *WebhookHandler {
	return &WebhookHandler{
		botStore:    botStore,
		livechatSDK: liveChatSDK,
	}
}

type Webhook struct {
	Action         string          `json:"action"`
	Payload        json.RawMessage `json:"payload"`
	OrganizationID string          `json:"organization_id"`
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
type Postback struct {
	ID       string `json:"id"`
	ThreadID string `json:"thread_id"`
	EventID  string `json:"event_id"`
	Type     string `json:"type"`
	Value    string `json:"value"`
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

func loadRichMessage() (json.RawMessage, error) {
	data, err := os.ReadFile("rich_message.json")
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (h *WebhookHandler) Reply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var webhook Webhook

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("Error decoding webhook JSON: %v", err)
		return
	}
	defer r.Body.Close()

	log.Printf("Received webhook: %s, payload %s", webhook.Action, string(webhook.Payload))

	switch {
	case webhook.Action == "incoming_event":
		if err := h.handleIncomingEvent(webhook); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case webhook.Action == "incoming_chat":
		if err := h.handleIncomingChat(webhook); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		log.Printf("Webhook not recognized: %v", webhook)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleIncomingEvent(webhook Webhook) error {
	var incomingEvent IncomingEvent
	if err := json.Unmarshal(webhook.Payload, &incomingEvent); err != nil {
		return fmt.Errorf("failed to unmarshall incoming event payload: %w", err)
	}
	token, exists := h.botStore.GetBotToken(webhook.OrganizationID)
	if !exists {
		return fmt.Errorf("bot token does not exists for organizationID: %s", webhook.OrganizationID)
	}
	var postback Postback
	if err := json.Unmarshal(incomingEvent.Event.Postback, &postback); err != nil {
		return fmt.Errorf("failed to unmarshal postback: %w", err)
	}
	log.Printf("Postback ID- %s", postback.ID)
	switch postback.ID {
	case "transfer_to_agent":
		return h.handleTransferToAgent(incomingEvent, token)
	case "continue_chat_with_bot":
		return h.handleContinueWithBot(incomingEvent, token)
	default:
		return fmt.Errorf("Postback ID not configured %s", postback.ID)
	}
	return nil
}

func (h *WebhookHandler) handleTransferToAgent(event IncomingEvent, token string) error {
	message := map[string]string{
		"type": "message",
		"text": "Understand. I will transfer chat to an agent now.",
	}
	if _, err := h.livechatSDK.SendEvent(event.ChatID, token, message); err != nil {
		return fmt.Errorf("failed to send transfer message: %w", err)
	}
	if err := h.livechatSDK.TransferChat(event.ChatID, token); err != nil {
		return fmt.Errorf("failed to transfer chat: %w", err)
	}
	return nil
}

func (h *WebhookHandler) handleContinueWithBot(event IncomingEvent, token string) error {
	message := map[string]string{
		"type": "message",
		"text": "Super, How can I help you",
	}
	if _, err := h.livechatSDK.SendEvent(event.ChatID, token, message); err != nil {
		return fmt.Errorf("failed to send continue message: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleIncomingChat(webhook Webhook) error {
	var incomingChat IncomingChat
	if err := json.Unmarshal(webhook.Payload, &incomingChat); err != nil {
		return fmt.Errorf("failed to parse incoming chat payload %w", err)
	}
	richMessage, err := loadRichMessage()
	if err != nil {
		return fmt.Errorf("failed to get rich message: %w", err)
	}
	token, exists := h.botStore.GetBotToken(webhook.OrganizationID)
	if !exists {
		return fmt.Errorf("bot token does not exist for organizationID: %s", webhook.OrganizationID)
	}
	response, err := h.livechatSDK.SendEvent(incomingChat.Chat.ID, token, richMessage)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	log.Printf("Rich message sent successfully, event ID: %s", response.EventID)
	return nil
}
