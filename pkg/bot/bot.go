package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/events"
	"github.com/bwmarrin/discordgo"
)

func Init(cfg *config.Config) error {
	botSession, err := discordgo.New("Bot " + cfg.ExternalServicesConfig.Token)
	if err != nil {
		return err
	}

	user, err := botSession.User("@me")
	if err != nil {
		return err
	}

	registerEvents(botSession, cfg, user)

	if err = botSession.Open(); err != nil {
		return err
	}

	fmt.Println("DiscordBot is running!")
	return nil
}

func registerEvents(s *discordgo.Session, cfg *config.Config, u *discordgo.User) {
	joinLeaveHandler := events.NewJoinLeaveHandler()

	s.AddHandler(joinLeaveHandler.HandlerJoin)
	s.AddHandler(joinLeaveHandler.HandlerLeave)

	s.AddHandler(events.NewMessageHandler(cfg, u).Handler)
	s.AddHandler(events.NewReadyHandler().Handler)
}
