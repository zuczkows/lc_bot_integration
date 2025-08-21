package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/livechat/sdk"
	"github.com/zuczkows/text-bot-integration/internal/server"
	"github.com/zuczkows/text-bot-integration/internal/utils"
)

func main() {
	configPath := "internal/config/config.json"

	cfg := config.ParseConfig(configPath)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", utils.GenerateBasicAuthToken(cfg.AccountID, cfg.PersonalToken))
	liveChatSDK := sdk.NewLivechatSDKClient(
		*http.DefaultClient,
		headers,
		cfg,
		logger,
	)
	botApplication := server.NewBotApplication(cfg, liveChatSDK)
	if err := botApplication.SetupApp(); err != nil {
		log.Fatalf("failed to setup application %v", err)
	}
	mux := botApplication.Mount()
	botApplication.Run(mux)

}
