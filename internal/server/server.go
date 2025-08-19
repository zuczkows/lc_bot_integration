package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/handlers"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
)

type BotApplication struct {
	config              config.Config
	webhookHandler      *handlers.WebhookHandler
	installationHandler *handlers.AppInstallationHandler
	liveChatSDK         *sdk.LivechatSDKClient
}

func generateBasicAuthToken(accountID string, personalToken config.Secret) string {
	credentials := fmt.Sprintf("%s:%s", accountID, personalToken)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

func NewBotApplication(cfg config.Config, logger *slog.Logger) *BotApplication {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", generateBasicAuthToken(cfg.AccountID, cfg.PersonalToken))
	liveChatSDK := sdk.NewLivechatSDKClient(
		*http.DefaultClient,
		headers,
		cfg.ApiUrl,
		cfg.ClientID,
		logger,
	)
	return &BotApplication{
		config:              cfg,
		webhookHandler:      handlers.NewWebhookHandler(),
		installationHandler: handlers.NewAppInstallationHandler(liveChatSDK),
		liveChatSDK:         liveChatSDK,
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
