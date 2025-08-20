package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type LivechatSDKClient struct {
	httpClient http.Client
	header     http.Header
	url        string
	logger     *slog.Logger
	clientID   string
}

func NewLivechatSDKClient(httpClient http.Client, header http.Header, url string, clientID string, logger *slog.Logger) *LivechatSDKClient {
	return &LivechatSDKClient{
		httpClient: httpClient,
		header:     header,
		url:        url,
		clientID:   clientID,
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

type deleteBotRequest struct {
	BotID         string `json:"id"`
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

type listWebhooksRequest struct {
	OwnerClientID string `json:"owner_client_id,omitempty"`
}

func (c *LivechatSDKClient) MakeRequest(path string, method string, body []byte) ([]byte, error) {
	c.logger.Info("Making API request",
		slog.String("method", method),
		slog.String("url", c.url+path),
		slog.String("request_body", string(body)),
	)

	req, err := http.NewRequest(http.MethodPost, c.url+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}
	req.Header = c.header

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
		OwnerClientID: c.clientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}
	responseBody, err := c.MakeRequest("/configuration/action/create_bot", "POST", body)
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
			SecretKey: "xsdhai232",
			URL:       url,
			Type:      "bot",
		},
		OwnerClientID: c.clientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}
	responseBody, err := c.MakeRequest("/configuration/action/register_webhook", "POST", body)
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

	_, err = c.MakeRequest("/agent/action/set_routing_status", "POST", body)
	if err != nil {
		return fmt.Errorf("failed to set routing status: %w", err)
	}
	return nil
}
