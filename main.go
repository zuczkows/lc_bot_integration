package main

import (
	"log/slog"
	"os"

	"github.com/zuczkows/text-bot-integration/internal/config"
	"github.com/zuczkows/text-bot-integration/internal/server"
)

func main() {
	configPath := "internal/config/config.json"

	cfg := config.ParseConfig(configPath)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	botApplication := server.NewBotApplication(cfg, logger)
	mux := botApplication.Mount()
	botApplication.Run(mux)

}
