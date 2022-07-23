package main

import (
	"github.com/beamer64/buddieBot/pkg/bot"
	"github.com/beamer64/buddieBot/pkg/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		panic(err)
	}

	err = bot.Init(cfg)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
}
