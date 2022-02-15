package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/commands"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/godagpi/dagpi"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type GuildJoinLeaveHandler struct{ cfg *config.Configs }
type ReactionHandler struct{ cfg *config.Configs }
type ReadyHandler struct{ cfg *config.Configs }
type GuildCreateHandler struct{ cfg *config.Configs }
type CommandHandler struct{ cfg *config.Configs }

func NewCommandHandler(cfg *config.Configs) *CommandHandler {
	return &CommandHandler{cfg: cfg}
}

func NewGuildCreateHandler(cfg *config.Configs) *GuildCreateHandler {
	return &GuildCreateHandler{cfg: cfg}
}

func NewGuildJoinLeaveHandler(cfg *config.Configs) *GuildJoinLeaveHandler {
	return &GuildJoinLeaveHandler{cfg: cfg}
}

func NewReactionHandler(cfg *config.Configs) *ReactionHandler {
	return &ReactionHandler{cfg: cfg}
}

func NewReadyHandler(cfg *config.Configs) *ReadyHandler {
	return &ReadyHandler{cfg: cfg}
}

// Events

var firstRun = true

// GuildCreateHandler joins new guild
func (g *GuildCreateHandler) GuildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	if !IsLaunchedByDebugger() {
		if !firstRun {
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, ""))
			}
		}
		firstRun = false
	}
}

// CommandHandler new commands
func (c *CommandHandler) CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			client := dagpi.Client{Auth: c.cfg.Configs.Keys.DagpiAPIkey}
			h(s, i, c.cfg, client)
		}
	case discordgo.InteractionMessageComponent:
		if h, ok := commands.ComponentHandlers[i.MessageComponentData().CustomID]; ok {
			client := dagpi.Client{Auth: c.cfg.Configs.Keys.DagpiAPIkey}
			h(s, i, c.cfg, client)
		}
	}
}

// ReadyHandler session is created
func (h *ReadyHandler) ReadyHandler(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateGameStatus(0, "try /tuuck")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(h.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, ""))
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
			_, _ = s.ChannelMessageSendEmbed(r.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, mr.GuildID))
		}
	}
}

// GuildJoinHandler when someone joins our server
func (d *GuildJoinLeaveHandler) GuildJoinHandler(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, e.GuildID))
	}

	fmt.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", e.Member.User.String(), guild.Name)
}

// GuildLeaveHandler when someone leaves our server
func (d *GuildJoinLeaveHandler) GuildLeaveHandler(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, e.GuildID))
	}

	fmt.Printf("%s left the server %s\n Seacrest OUT..", e.Member.User.String(), guild.Name)
}
