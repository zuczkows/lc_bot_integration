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

func (c *LivechatSDKClient) CreateBot(name string) (*CreateBotResponse, error) {
	request := createBotRequest{
		Name:          name,
		OwnerClientID: c.clientID,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request")
	}

	c.logger.Info("Making API request",
		slog.String("method", "POST"),
		slog.String("url", c.url+"/configuration/action/create_bot"),
		slog.String("request_body", string(body)),
	)

	req, err := http.NewRequest(http.MethodPost, c.url+"/configuration/action/create_bot", bytes.NewBuffer(body))
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

	c.logger.Info("API response received",
		slog.Int("status_code", resp.StatusCode),
		slog.String("response_body", string(responseBody)),
	)

	var createBotResp CreateBotResponse
	if err := json.Unmarshal(responseBody, &createBotResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createBotResp, nil
}
