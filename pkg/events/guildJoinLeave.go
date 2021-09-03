package events

import (
	"fmt"
	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
)

type JoinLeaveHandler struct{}

func NewJoinLeaveHandler() *JoinLeaveHandler {
	return &JoinLeaveHandler{}
}

func (d *JoinLeaveHandler) HandlerJoin(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Println("Failed getting guild object: ", err)
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	fmt.Printf("Hey! Look at this goofy goober! %s joined our %s server!\n", e.Member.User.String(), guild.Name)
}

func (d *JoinLeaveHandler) HandlerLeave(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	guild, err := s.Guild(e.GuildID)
	if err != nil {
		fmt.Println("Failed getting guild object: ", err)
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	fmt.Printf("%s left the server %s\n Seacrest OUT..", e.Member.User.String(), guild.Name)
}
