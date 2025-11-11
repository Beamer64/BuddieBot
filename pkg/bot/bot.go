package bot

import (
	"fmt"
	"log"

	"github.com/Beamer64/BuddieBot/pkg/commands/slash"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/events"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"

	"time"
)

func Init(cfg *config.Configs) error {
	token := ""
	botENV := ""
	if helper.IsLaunchedByDebugger() {
		token = cfg.Configs.Keys.TestBotToken
		botENV = "BB Test is ready for commands!"
	} else {
		token = cfg.Configs.Keys.ProdBotToken
		botENV = "BuddieBot is ready for commands!"
	}

	botSession, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %v", err)
	}

	user, err := botSession.User("@me")
	if err != nil {
		return fmt.Errorf("failed to grab Discord session User: %v", err)
	}

	registerEvents(botSession, cfg, user)

	botSession.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	if err = botSession.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %v", err)
	}

	err = registerCommands(botSession)
	if err != nil {
		return fmt.Errorf("failed to register commands: %v", err)
	}

	log.Println(botENV)
	return nil
}

func registerEvents(s *discordgo.Session, cfg *config.Configs, u *discordgo.User) {
	// Session
	s.AddHandler(events.NewReadyHandler(cfg).ReadyHandler)

	// Guild
	s.AddHandler(events.NewGuildHandler(cfg).GuildCreateHandler)
	s.AddHandler(events.NewGuildHandler(cfg).GuildDeleteHandler)
	s.AddHandler(events.NewGuildHandler(cfg).GuildJoinHandler)
	s.AddHandler(events.NewGuildHandler(cfg).GuildLeaveHandler)

	// Members
	s.AddHandler(events.NewGuildHandler(cfg).GuildMemberUpdateHandler)

	// Messages
	s.AddHandler(events.NewMessageCreateHandler(cfg, u).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler(cfg, u).ReactHandlerAdd)

	// Commands
	s.AddHandler(events.NewCommandHandler(cfg).CommandHandler)
}

func registerCommands(s *discordgo.Session) error {
	log.Println("Updating commands")

	// added sleep timer to allow time for
	// ApplicationCommandBulkOverwrite after creating bot session
	time.Sleep(3 * time.Second)
	commandsRegistered, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", slash.Commands)
	if err != nil {
		return err
	}

	subCmds := 0
	for _, v := range commandsRegistered {
		subCmds += len(v.Options)
	}

	log.Printf("%d Commands Registered\n", len(commandsRegistered))
	log.Printf("%d Sub-Commands Registered\n", subCmds)
	return nil
}
