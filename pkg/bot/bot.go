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
	"log"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"time"
)

func Init(cfg *config.Configs) error {
	token := ""
	botENV := ""
	if helper.IsLaunchedByDebugger() {
		token = cfg.Configs.Keys.TestBotToken
		botENV = "BB Test is ready to go!"
	} else {
		token = cfg.Configs.Keys.ProdBotToken
		botENV = "BuddieBot is ready to go!"
	}

	botSession, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %v", err)
	}

	user, err := botSession.User("@me")
	if err != nil {
		return fmt.Errorf("failed to grab Discord session User: %v", err)
	}

	dbClient, err := createDynamoDBClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB client: %v", err)
	}

	registerEvents(botSession, cfg, user, dbClient)

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

func createDynamoDBClient(cfg *config.Configs) (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(cfg.Configs.Database.Region),
			Credentials: credentials.NewStaticCredentials(cfg.Configs.Database.AccessKey, cfg.Configs.Database.SecretKey, ""),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	dbClient := dynamodb.New(sess)
	return dbClient, nil
}

func registerEvents(s *discordgo.Session, cfg *config.Configs, u *discordgo.User, dbc *dynamodb.DynamoDB) {
	// Session
	s.AddHandler(events.NewReadyHandler(cfg).ReadyHandler)

	// Guild
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildCreateHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildDeleteHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildJoinHandler)
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildLeaveHandler)

	// Members
	s.AddHandler(events.NewGuildHandler(cfg, dbc).GuildMemberUpdateHandler)

	// Messages
	s.AddHandler(events.NewMessageCreateHandler(cfg, u, dbc).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler(cfg, u).ReactHandlerAdd)

	// Commands
	s.AddHandler(events.NewCommandHandler(cfg).CommandHandler)
}

func registerCommands(s *discordgo.Session) error {
	log.Println("Updating commands")

	// added sleep timer to allow time for
	// ApplicationCommandBulkOverwrite after creating bot session
	time.Sleep(3 * time.Second)
	commandsRegistered, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands.Commands)
	if err != nil {
		return err
	}

	log.Printf("%d commands registered\n", len(commandsRegistered))
	return nil
}
