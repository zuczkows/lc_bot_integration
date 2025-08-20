package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
)

type AppInstallationHandler struct {
	livechatSDK *sdk.LivechatSDKClient
	config      config.Config
}

func NewAppInstallationHandler(liveChatSDK *sdk.LivechatSDKClient, config config.Config) *AppInstallationHandler {
	return &AppInstallationHandler{
		livechatSDK: liveChatSDK,
		config:      config,
	}
}

func (h *AppInstallationHandler) AppInstallation(w http.ResponseWriter, r *http.Request) {
	var webhook AppInstallUninstallWebhook

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("error decoding webhook %v", err)
	}
	defer r.Body.Close()
	log.Printf("Received webhook - AppID: %s, Event: %s, LicenseID: %d",
		webhook.AppID, webhook.Event, webhook.LicenseID)

	w.WriteHeader(http.StatusAccepted)

	if webhook.Event == "application_installed" {
		botResponse, err := h.livechatSDK.CreateBot("Zuczkows-Bot-007")
		if err != nil {
			log.Printf("Failed to create bot: %v", err)
			return
		}
		log.Printf("Bot created succesfully - ID %s", botResponse.BotID)

		err = h.livechatSDK.SetRoutingStatus("accepting_chats", botResponse.BotID)
		if err != nil {
			log.Printf("Failed to set routing status for bot: %v", err)
			return
		}
		log.Printf("Bot - %s status changed to accepting chats", botResponse.BotID)
	}
}

type AppInstallUninstallWebhook struct {
	AppID     string `json:"applicationID"`
	Event     string `json:"event"`
	LicenseID int64  `json:"licenseID"`
}
