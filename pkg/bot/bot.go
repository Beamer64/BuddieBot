package bot

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/beamer64/buddieBot/pkg/commands"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/events"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"time"
)

func Init(cfg *config.Configs) error {
	var botSession *discordgo.Session
	botENV := ""
	if helper.IsLaunchedByDebugger() {
		bs, err := discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
		if err != nil {
			return err
		}
		botSession = bs
		botENV = "BB Test is ready to go!"
	} else {
		bs, err := discordgo.New("Bot " + cfg.Configs.Keys.ProdBotToken)
		if err != nil {
			return err
		}
		botSession = bs
		botENV = "BuddieBot is ready to go!"
	}

	user, err := botSession.User("@me")
	if err != nil {
		return err
	}

	dynamodbSess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(cfg.Configs.Database.Region),
			Credentials: credentials.NewStaticCredentials(cfg.Configs.Database.AccessKey, cfg.Configs.Database.SecretKey, ""),
		},
	)
	if err != nil {
		return err
	}

	dbClient := dynamodb.New(dynamodbSess)

	registerEvents(botSession, cfg, user, dbClient)

	botSession.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
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

func registerEvents(s *discordgo.Session, cfg *config.Configs, u *discordgo.User, dbc *dynamodb.DynamoDB) {
	s.AddHandler(events.NewReadyHandler(cfg).ReadyHandler)

	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildCreateHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildDeleteHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildJoinHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildLeaveHandler)

	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildMemberUpdateHandler)

	s.AddHandler(events.NewMessageCreateHandler(cfg, u, dbc).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler(cfg, u).ReactHandlerAdd)

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
