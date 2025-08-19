package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/zuczkows/text-bot-integration/internal/server"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
	config := server.Config{
		Addr:          ":8080",
		PersonalToken: os.Getenv("PERSONAL_TOKEN"),
		AccountID:     os.Getenv("ACCOUNT_ID"),
		ClientID:      os.Getenv("CLIENT_ID"),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	botApplication := server.NewBotApplication(config, logger)
	mux := botApplication.Mount()
	botApplication.Run(mux)

}
