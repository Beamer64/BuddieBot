package main

import (
	"github.com/beamer64/discordBot/bot"
	"github.com/beamer64/discordBot/config"
)

func main() {
	defer config.Recovered()

	config_f, auth, err := config.ReadConfig("config/config.json", "config/auth.json")
	if err != nil {
		panic(err)
	}

	command, err := config.ReadCommands("config/command.json")
	if err != nil {
		panic(err)
	}

	err = bot.Start(config_f, auth, command)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
	return
}
