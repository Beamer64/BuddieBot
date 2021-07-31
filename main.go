package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/beamer64/discordBot/bot"
	"github.com/beamer64/discordBot/config"
	"github.com/gorilla/mux"
)

func main() {
	defer config.Recovered()

	config_f, auth, comm, err := config.ReadConfig("config/config.json", "config/auth.json", "config/command.json")
	if err != nil {
		panic(err)
	}

	err = bot.Start(config_f, auth, comm)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("upchuck")
	}).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
