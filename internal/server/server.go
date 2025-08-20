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
	webhooks := []string{"incoming_chat", "incoming_event"}
	for _, action := range webhooks {
		registerWebhookResponse, err := app.livechatSDK.RegisterWebhook(action, app.config.WebhookUrl)
		if err != nil {
			return fmt.Errorf("failed to register webhook %s: %w", action, err)
		}
		log.Printf("Webhook incoming_event created succesfully - ID %s", registerWebhookResponse.ID)
	}
	return nil
}
