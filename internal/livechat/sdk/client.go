package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/zuczkows/text-bot-integration/internal/config"
)

type LivechatSDKClient struct {
	httpClient http.Client
	header     http.Header
	config     config.Config
	logger     *slog.Logger
}

func NewLivechatSDKClient(httpClient http.Client, header http.Header, config config.Config, logger *slog.Logger) *LivechatSDKClient {
	return &LivechatSDKClient{
		httpClient: httpClient,
		header:     header,
		config:     config,
		logger:     logger,
	}
}

type Webhook struct {
	Action    string `json:"action"`
	SecretKey string `json:"secret_key"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

type setRoutingStatusRequest struct {
	Status  string `json:"status"`
	AgentID string `json:"agent_id"`
}

type registerWebhookRequest struct {
	*Webhook
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

type registerWebhookResponse struct {
	ID string `json:"id"`
}

type unregisterWebhookRequest struct {
	ID            string `json:"id"`
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

// Bot request could have more fields but for this app name and owner_client_id are enough
type createBotRequest struct {
	Name          string `json:"name"`
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

type CreateBotResponse struct {
	BotID  string `json:"id"`
	Secret string `json:"secret"`
}

type listWebhooksRequest struct {
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

type issueBotTokenRequest struct {
	BotID          string `json:"bot_id"`
	ClientID       string `json:"client_id"`
	Secret         string `json:"bot_secret"`
	OrganizationID string `json:"organization_id"`
}

type issueBotTokenResponse struct {
	Token string `json:"token"`
}

type sendEventRequest struct {
	ChatID string      `json:"chat_id"`
	Event  interface{} `json:"event"`
}

type sendEventResponse struct {
	EventID string `json:"event_id"`
}

// Only few fields - not whole response needed
type registeredWebhook struct {
	ID            string `json:"id"`
	Action        string `json:"action"`
	URL           string `json:"url"`
	Type          string `json:"type"`
	OwnerClientID string `json:"owner_client_id"`
}

type listWebhooksResponse []registeredWebhook

type transferChatRequest struct {
	ID string `json:"id"`
}

func (c *LivechatSDKClient) MakeRequest(path string, method string, body []byte, authToken ...string) ([]byte, error) {
	c.logger.Info("Making API request",
		slog.String("method", method),
		slog.String("url", c.config.ApiUrl+path),
		slog.String("request_body", string(body)),
	)

	req, err := http.NewRequest(http.MethodPost, c.config.ApiUrl+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}
	req.Header = c.header.Clone()
	if len(authToken) > 0 && authToken[0] != "" {
		req.Header.Set("Authorization", authToken[0])
		log.Printf("Making request as a bot with JWT: %s", authToken[0]) //debug log with plain token
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response")
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("API request failed",
			slog.Int("status_code", resp.StatusCode),
			slog.String("error_message", string(responseBody)))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}
	c.logger.Info("API response received",
		slog.Int("status_code", resp.StatusCode),
		slog.String("response_body", string(responseBody)),
	)
	return responseBody, nil
}

func (c *LivechatSDKClient) CreateBot(name string) (*CreateBotResponse, error) {
	request := createBotRequest{
		Name:          name,
		OwnerClientID: c.config.ClientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}
	responseBody, err := c.MakeRequest("/configuration/action/create_bot", http.MethodPost, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request")
	}

	var createBotResp CreateBotResponse
	if err := json.Unmarshal(responseBody, &createBotResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createBotResp, nil
}

func (c *LivechatSDKClient) RegisterWebhook(action, url string) (*registerWebhookResponse, error) {
	request := registerWebhookRequest{
		Webhook: &Webhook{
			Action:    action,
			SecretKey: c.config.SecretKey,
			URL:       url,
			Type:      "bot",
		},
		OwnerClientID: c.config.ClientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}
	responseBody, err := c.MakeRequest("/configuration/action/register_webhook", http.MethodPost, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request")
	}
	var registerWebhookResp registerWebhookResponse
	if err := json.Unmarshal(responseBody, &registerWebhookResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &registerWebhookResp, nil
}

func (c *LivechatSDKClient) SetRoutingStatus(status, agentID string) error {
	request := setRoutingStatusRequest{
		Status:  status,
		AgentID: agentID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request")
	}

	_, err = c.MakeRequest("/agent/action/set_routing_status", http.MethodPost, body)
	if err != nil {
		return fmt.Errorf("failed to set routing status: %w", err)
	}
	return nil
}

func (c *LivechatSDKClient) ListWebhooks() (*listWebhooksResponse, error) {
	request := listWebhooksRequest{
		OwnerClientID: c.config.ClientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}

	responseBody, err := c.MakeRequest("/configuration/action/list_webhooks", http.MethodPost, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", responseBody)
	}

	var webhookResponse listWebhooksResponse
	if err := json.Unmarshal(responseBody, &webhookResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
	}
	return &webhookResponse, nil
}

func (c *LivechatSDKClient) IssueBotToken(bot_id, secret, organization_id string) (*issueBotTokenResponse, error) {
	request := issueBotTokenRequest{
		BotID:          bot_id,
		ClientID:       c.config.ClientID,
		Secret:         secret,
		OrganizationID: organization_id,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}
	responseBody, err := c.MakeRequest("/configuration/action/issue_bot_token", http.MethodPost, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", responseBody)
	}
	var botTokenResponse issueBotTokenResponse
	if err := json.Unmarshal(responseBody, &botTokenResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
	}
	return &botTokenResponse, nil
}

func (c *LivechatSDKClient) SendEvent(chatID, token string, event interface{}) (*sendEventResponse, error) {
	request := sendEventRequest{
		ChatID: chatID,
		Event:  event,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	responseBody, err := c.MakeRequest("/agent/action/send_event", http.MethodPost, body, token)
	if err != nil {
		return nil, err
	}

	var sendEventResp sendEventResponse
	if err := json.Unmarshal(responseBody, &sendEventResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	return &sendEventResp, nil
}

func (c *LivechatSDKClient) TransferChat(chatID, token string) error {
	request := transferChatRequest{
		ID: chatID,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.MakeRequest("/agent/action/transfer_chat", http.MethodPost, body, token)
	if err != nil {
		return err
	}

	return nil
}
