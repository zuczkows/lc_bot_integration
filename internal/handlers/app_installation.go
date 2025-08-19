package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
)

type AppInstallationHandler struct {
	LiveChatSDK *sdk.LivechatSDKClient
}

func NewAppInstallationHandler(liveChatSDK *sdk.LivechatSDKClient) *AppInstallationHandler {
	return &AppInstallationHandler{
		LiveChatSDK: liveChatSDK,
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
		botResponse, err := h.LiveChatSDK.CreateBot("Zuczkows-Bot-007")
		if err != nil {
			log.Printf("Failed to create bot: %v", err)
			return
		}

		log.Printf("Bot created succesfully - ID %s", botResponse.BotID)
	}
}

type AppInstallUninstallWebhook struct {
	AppID     string `json:"applicationID"`
	Event     string `json:"event"`
	LicenseID int64  `json:"licenseID"`
}
