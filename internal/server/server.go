package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/handlers"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
)

type BotApplication struct {
	config              config.Config
	webhookHandler      *handlers.WebhookHandler
	installationHandler *handlers.AppInstallationHandler
	livechatSDK         *sdk.LivechatSDKClient
}

func NewBotApplication(cfg config.Config, sdk *sdk.LivechatSDKClient) *BotApplication {
	return &BotApplication{
		config:              cfg,
		webhookHandler:      handlers.NewWebhookHandler(),
		installationHandler: handlers.NewAppInstallationHandler(sdk, cfg),
		livechatSDK:         sdk,
	}
}

func (app *BotApplication) Mount() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("POST /webhook", app.webhookHandler.Reply)
	r.HandleFunc("POST /app_installation", app.installationHandler.AppInstallation)
	return r
}

func (app *BotApplication) Run(mux http.Handler) error {
	srv := http.Server{
		Addr:    app.config.Addr,
		Handler: mux,
	}
	log.Printf("server has started at %s", app.config.Addr)
	return srv.ListenAndServe()
}

// Register webhooks once per application - not once every app install
func (app *BotApplication) SetupAPP() error {
	registeredWebhooks, err := app.livechatSDK.ListWebhooks()
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %v", registeredWebhooks)
	}
	existingWebhooks := make(map[string]bool)
	for _, webhook := range *registeredWebhooks {
		if webhook.OwnerClientID == app.config.ClientID {
			existingWebhooks[webhook.Action] = true
			log.Printf("Found existing webhook: %s for client: %s", webhook.Action, webhook.ID)
		}
	}

	webhooks := []string{"incoming_chat", "incoming_event"}
	for _, action := range webhooks {
		if existingWebhooks[action] {
			log.Printf("Webhook %s already exists for client %s, skipping registration", action, app.config.ClientID)
			continue
		}

		registerWebhookResponse, err := app.livechatSDK.RegisterWebhook(action, app.config.WebhookUrl)
		if err != nil {
			return fmt.Errorf("failed to register webhook %s: %w", action, err)
		}
		log.Printf("Webhook %s created successfully - ID %s", action, registerWebhookResponse.ID)
	}
	return nil
}
