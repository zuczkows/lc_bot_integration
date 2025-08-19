package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zuczkows/text-bot-integration/internal/handlers"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
)

type Config struct {
	Addr          string
	PersonalToken string
	AccountID     string
	ClientID      string
}

type BotApplication struct {
	config              Config
	webhookHandler      *handlers.WebhookHandler
	installationHandler *handlers.AppInstallationHandler
	liveChatSDK         *sdk.LivechatSDKClient
}

func generateBasicAuthToken(accountID, personalToken string) string {
	credentials := fmt.Sprintf("%s:%s", accountID, personalToken)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

func NewBotApplication(config Config, logger *slog.Logger) *BotApplication {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", generateBasicAuthToken(config.AccountID, config.PersonalToken))
	liveChatSDK := sdk.NewLivechatSDKClient(
		http.Client{},
		headers,
		"https://api.labs.livechatinc.com/v3.6",
		config.ClientID,
		logger,
	)
	return &BotApplication{
		config:              config,
		webhookHandler:      handlers.NewWebhookHandler(),
		installationHandler: handlers.NewAppInstallationHandler(liveChatSDK),
		liveChatSDK:         liveChatSDK,
	}
}

func (app *BotApplication) Mount() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/webhook", app.webhookHandler.Reply).Methods("POST")
	r.HandleFunc("/app_installation", app.installationHandler.AppInstallation).Methods("POST")
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
