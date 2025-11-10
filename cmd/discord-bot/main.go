package main

import (
	"math/rand"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/bot"
	"github.com/Beamer64/BuddieBot/pkg/config"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

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
