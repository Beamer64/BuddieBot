package events

import (
	"fmt"
	"github.com/beamer64/buddieBot/pkg/commands/prefix"
	"github.com/beamer64/buddieBot/pkg/commands/slash"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"log"
)

type MessageCreateHandler struct {
	cfg   *config.Configs
	botID string
}
type ReactionHandler struct {
	cfg   *config.Configs
	botID string
}

type ReadyHandler struct{ cfg *config.Configs }
type CommandHandler struct{ cfg *config.Configs }
type GuildHandler struct {
	cfg *config.Configs
}

func NewMessageCreateHandler(cfg *config.Configs, u *discordgo.User) *MessageCreateHandler {
	return &MessageCreateHandler{
		cfg:   cfg,
		botID: u.ID,
	}
}

func NewCommandHandler(cfg *config.Configs) *CommandHandler {
	return &CommandHandler{cfg: cfg}
}

func NewGuildHandler(cfg *config.Configs) *GuildHandler {
	return &GuildHandler{
		cfg: cfg,
	}
}

func NewReactionHandler(cfg *config.Configs, u *discordgo.User) *ReactionHandler {
	return &ReactionHandler{
		cfg:   cfg,
		botID: u.ID,
	}
}

func NewReadyHandler(cfg *config.Configs) *ReadyHandler {
	return &ReadyHandler{cfg: cfg}
}

// Events

// CommandHandler new commands
func (c *CommandHandler) CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := slash.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, c.cfg)
		}
	case discordgo.InteractionMessageComponent:
		if h, ok := slash.ComponentHandlers[i.MessageComponentData().CustomID]; ok {
			h(s, i, c.cfg)
		}
	}
}

// ReadyHandler session is created
func (h *ReadyHandler) ReadyHandler(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateGameStatus(0, "try /tuuck")
	if err != nil {
		helper.LogErrors(s, h.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, "")
		return
	}

	// FYI can get all connected Guild list here
	log.Println(fmt.Sprintf("Invited to %d Servers!", len(e.Guilds)))
	log.Printf("Logged in as %s\n", e.User.String())
}

// ReactHandlerAdd when reactions are added to messages
func (r *ReactionHandler) ReactHandlerAdd(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {
	if mr.UserID == r.botID {
		return
	}

	channel, err := s.Channel(mr.ChannelID)
	if err != nil {
		helper.LogErrors(s, r.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
		return
	}

	msg, err := s.ChannelMessage(channel.ID, mr.MessageID)
	if err != nil {
		helper.LogErrors(s, r.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
		return
	}

	// find the poll msg
	if msg.Content == "Poll Time!" {
		for _, v := range msg.Reactions {
			// remove extra reactions
			if v.Emoji.Name == mr.Emoji.Name && v.Count < 2 {
				err = s.MessageReactionRemove(channel.ID, msg.ID, mr.MessageReaction.Emoji.Name, mr.UserID)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(r.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, mr.GuildID))
				}
			}
		}
	} else if mr.MessageReaction.Emoji.Name == "grey_question" {
		msg, _ = s.ChannelMessage(mr.ChannelID, mr.MessageID)

		err = r.sendLmgtfy(s, msg)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			_, _ = s.ChannelMessageSendEmbed(r.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, mr.GuildID))
		}
	}
}

// MessageCreateHandler handles all messages sent to the discord server
func (d *MessageCreateHandler) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == d.botID {
		return
	}

	prefix.ParsePrefixCmds(s, m, d.cfg)
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
		helper.LogErrors(s, g.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
		return
	}

	log.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", m.Member.User.String(), guild.Name)
}

// GuildLeaveHandler when someone leaves our server
func (g *GuildHandler) GuildLeaveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		helper.LogErrors(s, g.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
		return
	}

	log.Printf("%s left the server %s\n Seacrest OUT..", m.Member.User.String(), guild.Name)
}

// GuildCreateHandler bot joins new guild
func (g *GuildHandler) GuildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	if helper.IsLaunchedByDebugger() {
		return
	}

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

	_, err := s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.EventNotifChannelID, embed)
	if err != nil {
		helper.LogErrors(s, g.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, e.ID)
		return
	}
}

// GuildDeleteHandler when bot leaves a server
func (g *GuildHandler) GuildDeleteHandler(s *discordgo.Session, e *discordgo.GuildDelete) {
	// TODO add this in
	/*err := Do sum
	if err != nil {
		helper.LogErrors(s, g.cfg.Configs.DiscordIDs.ErrorLogChannelID, err, e.ID)
		return
	}*/
}
