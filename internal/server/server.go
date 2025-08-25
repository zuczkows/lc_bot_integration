package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/handlers"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
	"github.com/zuczkows/text-bot-integration/internal/store"
)

type BotApplication struct {
	config              config.Config
	webhookHandler      *handlers.WebhookHandler
	installationHandler *handlers.AppInstallationHandler
	livechatSDK         *sdk.LivechatSDKClient
	server              *http.Server
}

func NewBotApplication(cfg config.Config, sdk *sdk.LivechatSDKClient) *BotApplication {
	botStore := store.NewBotStore()
	return &BotApplication{
		config:              cfg,
		webhookHandler:      handlers.NewWebhookHandler(botStore, sdk),
		installationHandler: handlers.NewAppInstallationHandler(sdk, cfg, botStore),
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
	app.server = &http.Server{
		Addr:         app.config.Addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverError := make(chan error, 1)
	go func() {
		log.Printf("server has started at %s", app.config.Addr)
		if err := app.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		log.Printf("Server error: %v", err)
	case sig := <-stop:
		log.Printf("Received shutdown signal: %v", sig)
	}

	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error %w", err)
	}
	log.Println("Server exited properly")
	return nil
}

// Register webhooks once per application - not once every app install
func (app *BotApplication) SetupApp() error {
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
