package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/commands"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/events"
	"github.com/bwmarrin/discordgo"
	"time"
)

func Init(cfg *config.ConfigStructs) error {
	var botSession *discordgo.Session
	botENV := ""
	if events.IsLaunchedByDebugger() {
		bs, err := discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
		if err != nil {
			return err
		}
		botSession = bs
		botENV = "BB Test is running!"
	} else {
		bs, err := discordgo.New("Bot " + cfg.Configs.Keys.ProdBotToken)
		if err != nil {
			return err
		}
		botSession = bs
		botENV = "BuddieBot is running!"
	}

	user, err := botSession.User("@me")
	if err != nil {
		return err
	}

	registerEvents(botSession, cfg, user)

	if err = botSession.Open(); err != nil {
		return err
	}

	err = registerCommands(botSession)
	if err != nil {
		return err
	}

	fmt.Println(botENV)
	return nil
}

func registerEvents(s *discordgo.Session, cfg *config.ConfigStructs, u *discordgo.User) {
	s.AddHandler(events.NewReadyHandler(cfg).ReadyHandler)

	s.AddHandler(events.NewGuildCreateHandler(cfg).GuildCreateHandler)
	s.AddHandler(events.NewGuildJoinLeaveHandler(cfg).GuildJoinHandler)
	s.AddHandler(events.NewGuildJoinLeaveHandler(cfg).GuildLeaveHandler)

	s.AddHandler(events.NewMessageCreateHandler(cfg, u).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler(cfg).ReactHandlerAdd)

	s.AddHandler(events.NewCommandHandler(cfg).CommandHandler)
}

func registerCommands(s *discordgo.Session) error {
	fmt.Println("Updating commands")

	time.Sleep(3 * time.Second)
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands.Commands)
	if err != nil {
		return err
	}

	cmds, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("%d Commands Registered", len(cmds)))
	return nil
}
