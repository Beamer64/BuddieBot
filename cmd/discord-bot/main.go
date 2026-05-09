package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Beamer64/BuddieBot/pkg/bot"
	"github.com/Beamer64/BuddieBot/pkg/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	if err := bot.Init(cfg); err != nil {
		bot.Shutdown()
		panic(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down…")
	bot.Shutdown()
}
