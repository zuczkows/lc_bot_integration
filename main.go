package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/zuczkows/text-bot-integration/internal/handlers"
)

type config struct {
	addr          string
	personalToken string
	accountID     string
}

type botApplication struct {
	config         config
	webhookHandler *handlers.WebhookHandler
}

func (app *botApplication) mount() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/webhook", app.webhookHandler.Reply).Methods("POST")
	return r
}

func (app *botApplication) run(mux http.Handler) error {
	srv := http.Server{
		Addr:    app.config.addr,
		Handler: mux,
	}
	log.Printf("server has started at %s", app.config.addr)
	return srv.ListenAndServe()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
	config := config{
		addr:          ":8080",
		personalToken: os.Getenv("PERSONAL_TOKEN"),
		accountID:     os.Getenv("ACCOUNT_ID"),
	}
	botApplication := botApplication{
		config:         config,
		webhookHandler: &handlers.WebhookHandler{},
	}
	mux := botApplication.mount()
	botApplication.run(mux)

}
