package main

import (
	"log"

	"github.com/Beamer64/BuddieBot/pkg/bot"
	"github.com/Beamer64/BuddieBot/pkg/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	err = bot.Init(cfg)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
}
