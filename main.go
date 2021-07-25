package main

import (
	"github.com/beamer64/discordBot/bot"
	"github.com/beamer64/discordBot/config"
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

	<-make(chan struct{})
	return
}
