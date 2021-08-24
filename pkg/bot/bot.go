package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	cfg   *config.Config
	botID string
}

func NewDiscordBot(cfg *config.Config) *DiscordBot {
	return &DiscordBot{
		cfg: cfg,
	}
}

func (d *DiscordBot) Start() error {
	goBot, err := discordgo.New("Bot " + d.cfg.ExternalServicesConfig.Token)
	if err != nil {
		return err
	}

	user, err := goBot.User("@me")
	if err != nil {
		return err
	}
	d.botID = user.ID

	goBot.AddHandler(d.messageHandler)
	goBot.AddHandler(d.guildCreateHandler)
	err = goBot.Open()
	if err != nil {
		return err
	}

	fmt.Println("DiscordBot is running!")
	return nil
}
