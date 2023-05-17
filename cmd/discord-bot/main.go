package main

import (
	"github.com/beamer64/buddieBot/pkg/bot"
	"github.com/beamer64/buddieBot/pkg/config"
	"math/rand"
	"time"
)

func main() {
	// deprecated
	//rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))

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
