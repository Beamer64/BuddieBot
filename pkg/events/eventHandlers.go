package events

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/commands/prefix"
	"github.com/Beamer64/BuddieBot/pkg/commands/slash"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// recoverPanic should be deferred at the top of each event handler so a
// single bad event can't crash the whole bot. Logs the panic + stack to the
// error channel and console.
func recoverPanic(s *discordgo.Session, cfg *config.Configs, guildID string) {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic in event handler: %v\n%s", r, debug.Stack())
		helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, err, guildID)
	}
}

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
	defer recoverPanic(s, c.cfg, i.GuildID)
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := slash.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, c.cfg)
		}
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID

		for key, h := range slash.ComponentHandlers {
			if strings.HasPrefix(customID, key) { // or strings.Contains(customID, key)
				h(s, i, c.cfg)
				return
			}
		}

		/*if h, ok := slash.ComponentHandlers[i.MessageComponentData().CustomID]; ok {
			h(s, i, c.cfg)
		}*/
	}
}

// ReadyHandler session is created
func (h *ReadyHandler) ReadyHandler(s *discordgo.Session, e *discordgo.Ready) {
	defer recoverPanic(s, h.cfg, "")
	err := s.UpdateGameStatus(0, "try /tuuck")
	if err != nil {
		helper.LogErrorsToErrorChannel(s, h.cfg.DiscordIDs.ErrorLogChannelID, err, "")
		return
	}

	// FYI can get all connected Guild list here
	log.Println(fmt.Sprintf("Invited to %d Servers!", len(e.Guilds)))
	log.Printf("Logged in as %s\n", e.User.String())
}

// ReactHandlerAdd when reactions are added to messages
func (r *ReactionHandler) ReactHandlerAdd(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {
	defer recoverPanic(s, r.cfg, mr.GuildID)
	if mr.UserID == r.botID {
		return
	}

	channel, err := s.Channel(mr.ChannelID)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, r.cfg.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
		return
	}

	msg, err := s.ChannelMessage(channel.ID, mr.MessageID)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, r.cfg.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
		return
	}

	// find the poll msg
	if msg.Content == helper.PollMessageContent {
		for _, v := range msg.Reactions {
			// remove extra reactions
			if v.Emoji.Name == mr.Emoji.Name && v.Count < 2 {
				err = s.MessageReactionRemove(channel.ID, msg.ID, mr.MessageReaction.Emoji.Name, mr.UserID)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, r.cfg.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
				}
			}
		}
	} else if mr.MessageReaction.Emoji.Name == helper.LmgtfyEmojiName {
		if err := r.sendLmgtfy(s, msg); err != nil {
			helper.LogErrorsToErrorChannel(s, r.cfg.DiscordIDs.ErrorLogChannelID, err, mr.GuildID)
		}
	}
}

// MessageCreateHandler handles all messages sent to the discord server
func (d *MessageCreateHandler) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer recoverPanic(s, d.cfg, m.GuildID)
	if m.Author.ID == d.botID {
		return
	}

	prefix.ParsePrefixCmds(s, m, d.cfg)
}

// GuildJoinHandler when someone joins our server
func (g *GuildHandler) GuildJoinHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	defer recoverPanic(s, g.cfg, m.GuildID)
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, g.cfg.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
		return
	}

	log.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", m.Member.User.String(), guild.Name)
}

// GuildLeaveHandler when someone leaves our server
func (g *GuildHandler) GuildLeaveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	defer recoverPanic(s, g.cfg, m.GuildID)
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, g.cfg.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
		return
	}

	log.Printf("%s left the server %s\n Seacrest OUT..", m.Member.User.String(), guild.Name)
}

// GuildCreateHandler bot joins new guild
func (g *GuildHandler) GuildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	defer recoverPanic(s, g.cfg, e.ID)
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

	_, err := s.ChannelMessageSendEmbed(g.cfg.DiscordIDs.EventNotifChannelID, embed)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, g.cfg.DiscordIDs.ErrorLogChannelID, err, e.ID)
		return
	}
}

// GuildDeleteHandler when bot leaves a server
func (g *GuildHandler) GuildDeleteHandler(s *discordgo.Session, e *discordgo.GuildDelete) {
	defer recoverPanic(s, g.cfg, e.ID)
	if helper.IsLaunchedByDebugger() {
		return
	}

	guildID := ""
	guildName := "Unknown"
	unavailable := false
	if e.Guild != nil {
		guildID = e.ID
		if e.Name != "" {
			guildName = e.Name
		}
		unavailable = e.Unavailable
	}
	if guildName == "Unknown" && e.BeforeDelete != nil && e.BeforeDelete.Name != "" {
		guildName = e.BeforeDelete.Name
	}

	embed := &discordgo.MessageEmbed{
		Title: "SERVER LEAVE",
		Color: 1564907,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "ServerID", Value: guildID, Inline: true},
			{Name: "Server Name", Value: guildName, Inline: true},
			{Name: "Unavailable", Value: fmt.Sprintf("%v", unavailable), Inline: true},
		},
	}

	_, err := s.ChannelMessageSendEmbed(g.cfg.DiscordIDs.EventNotifChannelID, embed)
	if err != nil {
		helper.LogErrorsToErrorChannel(s, g.cfg.DiscordIDs.ErrorLogChannelID, err, guildID)
		return
	}
}
