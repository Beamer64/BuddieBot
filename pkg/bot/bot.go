package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/commands"
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

	// fetch current commands
	fmt.Println("Fetching current commands")
	cmds, err := botSession.ApplicationCommands(botSession.State.User.ID, cfg.Configs.DiscordIDs.GuildID)
	if err != nil {
		return err
	}

	// unregister commands
	if len(cmds) > 0 {
		fmt.Println("Unregistering commands")
		for _, v := range cmds {
			err = botSession.ApplicationCommandDelete(botSession.State.User.ID, cfg.Configs.DiscordIDs.GuildID, v.ID)
			if err != nil {
				return err
			}
		}
	}

	// register new commands
	fmt.Println("Registering new commands")
	for _, v := range commands.Commands {
		_, err = botSession.ApplicationCommandCreate(botSession.State.User.ID, cfg.Configs.DiscordIDs.GuildID, v)
		if err != nil {
			return err
		}
	}

	fmt.Println("DiscordBot is running!")
	return nil
}

func registerEvents(s *discordgo.Session, cfg *config.ConfigStructs, u *discordgo.User) {
	s.AddHandler(events.NewReadyHandler().ReadyHandler)

	s.AddHandler(events.NewGuildJoinLeaveHandler().GuildJoinHandler)
	s.AddHandler(events.NewGuildJoinLeaveHandler().GuildLeaveHandler)

	s.AddHandler(events.NewMessageCreateHandler(cfg, u).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler().ReactHandlerAdd)

	s.AddHandler(events.NewCommandHandler(cfg).CommandHandler)
}
