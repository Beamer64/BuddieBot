package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/events"
	"github.com/bwmarrin/discordgo"
)

func Init(cfg *config.ConfigStructs) error {
	botSession, err := discordgo.New("Bot " + cfg.Configs.Keys.BotToken)
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

func registerEvents(s *discordgo.Session, cfg *config.ConfigStructs, u *discordgo.User) {
	s.AddHandler(events.NewGuildJoinLeaveHandler().GuildJoinHandler)
	s.AddHandler(events.NewGuildJoinLeaveHandler().GuildLeaveHandler)

	s.AddHandler(events.NewMessageCreateHandler(cfg, u).MessageCreateHandler)
	s.AddHandler(events.NewReadyHandler().ReadyHandler)

	s.AddHandler(events.NewReactionHandler().ReactHandlerAdd)
}
