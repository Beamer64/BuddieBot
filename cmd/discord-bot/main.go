package main

import (
	"github.com/beamer64/discordBot/pkg/bot"
	"github.com/beamer64/discordBot/pkg/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		panic(err)
	}

	err = bot.Init(cfg)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
}
