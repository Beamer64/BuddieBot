package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/commands"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type GuildJoinLeaveHandler struct{ cfg *config.ConfigStructs }
type ReactionHandler struct{ cfg *config.ConfigStructs }
type ReadyHandler struct{ cfg *config.ConfigStructs }
type GuildCreateHandler struct{ cfg *config.ConfigStructs }
type CommandHandler struct{ cfg *config.ConfigStructs }

func NewCommandHandler(cfg *config.ConfigStructs) *CommandHandler {
	return &CommandHandler{cfg: cfg}
}

func NewGuildCreateHandler(cfg *config.ConfigStructs) *GuildCreateHandler {
	return &GuildCreateHandler{cfg: cfg}
}

func NewGuildJoinLeaveHandler(cfg *config.ConfigStructs) *GuildJoinLeaveHandler {
	return &GuildJoinLeaveHandler{cfg: cfg}
}

func NewReactionHandler(cfg *config.ConfigStructs) *ReactionHandler {
	return &ReactionHandler{cfg: cfg}
}

func NewReadyHandler(cfg *config.ConfigStructs) *ReadyHandler {
	return &ReadyHandler{cfg: cfg}
}

// Events

func (g *GuildCreateHandler) GuildCreateHandler(s *discordgo.Session, e *discordgo.GuildCreate) {
	_, err := s.ChannelMessageSend(
		g.cfg.Configs.DiscordIDs.EventNotifChannelID,
		fmt.Sprintf("BuddieBot has joined ServerID: %s\nServerName: %s\nDescription: %s\nMemberCount: %v\nRegion: %s", e.ID, e.Name, e.Description, e.MemberCount, e.Region),
	)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSend(g.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
	}
}

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

func (h *ReadyHandler) ReadyHandler(s *discordgo.Session, e *discordgo.Ready) {
	err := s.UpdateGameStatus(0, "try /tuuck")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSend(h.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
	}

	// FYI can get all connected Guild list here
	fmt.Println(fmt.Sprintf("Invited to %d Servers!", len(e.Guilds)))
	fmt.Printf("Logged in as %s\n", e.User.String())
}

func (r *ReactionHandler) ReactHandlerAdd(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {
	if mr.MessageReaction.Emoji.Name == "lmgtfy" {
		msg, _ := s.ChannelMessage(mr.ChannelID, mr.MessageID)

		err := r.sendLmgtfy(s, msg)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			_, _ = s.ChannelMessageSend(r.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
		}
	}
}

func (d *GuildJoinLeaveHandler) GuildJoinHandler(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
	}

	fmt.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", e.Member.User.String(), guild.Name)
}

func (d *GuildJoinLeaveHandler) GuildLeaveHandler(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
	}

	fmt.Printf("%s left the server %s\n Seacrest OUT..", e.Member.User.String(), guild.Name)
}
