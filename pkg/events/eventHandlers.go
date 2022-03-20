package events

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/beamer64/discordBot/pkg/commands"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/database"
	"github.com/beamer64/discordBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type ReactionHandler struct{ cfg *config.Configs }
type ReadyHandler struct{ cfg *config.Configs }
type CommandHandler struct{ cfg *config.Configs }
type GuildHandler struct {
	cfg      *config.Configs
	dbClient *dynamodb.DynamoDB
}

func NewCommandHandler(cfg *config.Configs) *CommandHandler {
	return &CommandHandler{cfg: cfg}
}

func NewGuildHandler(cfg *config.Configs, dbc *dynamodb.DynamoDB) *GuildHandler {
	return &GuildHandler{
		cfg:      cfg,
		dbClient: dbc,
	}
}

func NewReactionHandler(cfg *config.Configs) *ReactionHandler {
	return &ReactionHandler{cfg: cfg}
}

func NewReadyHandler(cfg *config.Configs) *ReadyHandler {
	return &ReadyHandler{cfg: cfg}
}

// Events

// CommandHandler new commands
func (c *CommandHandler) CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, c.cfg)
		}
	case discordgo.InteractionMessageComponent:
		if h, ok := commands.ComponentHandlers[i.MessageComponentData().CustomID]; ok {
			h(s, i, c.cfg)
		}
	}
}

// ReadyHandler session is created
func (h *ReadyHandler) ReadyHandler(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateGameStatus(0, "try /tuuck")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(h.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, ""))
	}

	// FYI can get all connected Guild list here
	fmt.Println(fmt.Sprintf("Invited to %d Servers!", len(e.Guilds)))
	fmt.Printf("Logged in as %s\n", e.User.String())
}

// ReactHandlerAdd when reactions are added to messages
func (r *ReactionHandler) ReactHandlerAdd(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {
	if mr.MessageReaction.Emoji.Name == "grey_question" {
		msg, _ := s.ChannelMessage(mr.ChannelID, mr.MessageID)

		err := r.sendLmgtfy(s, msg)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			_, _ = s.ChannelMessageSendEmbed(r.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, mr.GuildID))
		}
	}
}

// GuildMemberUpdateHandler Sent when a guild member is updated.
func (g *GuildHandler) GuildMemberUpdateHandler(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	/*embed := &discordgo.MessageEmbed{
		Title: "Hey, GuildMemberUpdateHandler is working now",
		Color: 1321,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User",
				Value:  e.User.Username,
				Inline: true,
			},
			{
				Name:   "ID",
				Value:  e.User.ID,
				Inline: true,
			},
		},
	}
	_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.EventNotifChannelID, embed)*/
}

// GuildJoinHandler when someone joins our server
func (g *GuildHandler) GuildJoinHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
	}

	fmt.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", m.Member.User.String(), guild.Name)

	err = database.InsertDBmemberData(g.dbClient, m.Member, g.cfg)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
	}
}

// GuildLeaveHandler when someone leaves our server
func (g *GuildHandler) GuildLeaveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
	}

	fmt.Printf("%s left the server %s\n Seacrest OUT..", m.Member.User.String(), guild.Name)

	err = database.DeleteDBmemberData(g.dbClient, m.Member, g.cfg)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
	}
}

// GuildCreateHandler bot joins new guild
func (g *GuildHandler) GuildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	err := database.InsertDBguildItem(g.dbClient, e.Guild, g.cfg)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, e.ID))
	}

	if !helper.IsLaunchedByDebugger() {
		desc := "None"
		if e.Description != "" {
			desc = e.Description
		}
		embed := &discordgo.MessageEmbed{
			Title: "NEW SERVER JOIN",
			Color: 1564907,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "ServerID",
					Value:  e.ID,
					Inline: true,
				},
				{
					Name:   "Server Name",
					Value:  e.Name,
					Inline: true,
				},
				{
					Name:   "Member Count",
					Value:  fmt.Sprintf("%v", e.MemberCount),
					Inline: true,
				},
				{
					Name:   "Description",
					Value:  desc,
					Inline: false,
				},
			},
		}

		_, err = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.EventNotifChannelID, embed)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, e.ID))
		}
	}
}

// GuildDeleteHandler when bot leaves a server
func (g *GuildHandler) GuildDeleteHandler(s *discordgo.Session, e *discordgo.GuildDelete) {
	err := database.DeleteDBguildData(g.dbClient, e.Guild, g.cfg)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, e.ID))
	}
}
