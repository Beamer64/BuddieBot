package main

import (
	"github.com/beamer64/discordBot/pkg/bot"
	"github.com/beamer64/discordBot/pkg/config"
)

func main() {
	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		panic(err)
	}

	disbot := bot.NewDiscordBot(cfg)

	err = disbot.Start()
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
}
