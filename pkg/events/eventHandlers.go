package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/commands"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type GuildJoinLeaveHandler struct{}
type ReactionHandler struct{}
type ReadyHandler struct{}
type CommandHandler struct {
	cfg *config.ConfigStructs
}

func NewCommandHandler(cfg *config.ConfigStructs) *CommandHandler {
	return &CommandHandler{
		cfg: cfg,
	}
}

func NewGuildJoinLeaveHandler() *GuildJoinLeaveHandler {
	return &GuildJoinLeaveHandler{}
}

func NewReactionHandler() *ReactionHandler {
	return &ReactionHandler{}
}

func NewReadyHandler() *ReadyHandler {
	return &ReadyHandler{}
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
		}
	}
}

func (d *GuildJoinLeaveHandler) GuildJoinHandler(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Println("Failed getting guild object: ", err)
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	fmt.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", e.Member.User.String(), guild.Name)
}

func (d *GuildJoinLeaveHandler) GuildLeaveHandler(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Println("Failed getting guild object: ", err)
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	fmt.Printf("%s left the server %s\n Seacrest OUT..", e.Member.User.String(), guild.Name)
}
