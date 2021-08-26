package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
)

func (d *DiscordBot) memberHasRole(session *discordgo.Session, message *discordgo.MessageCreate, roleName string) bool {
	guildID := message.GuildID
	roleName = strings.ToLower(roleName)

	for _, roleID := range message.Member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}

func GetMentionedMemberFromList(memberName string, members []*discordgo.Member) string {
	memberName = strings.ToLower(memberName)

	if strings.HasPrefix(memberName, "<") {
		for _, memb := range members {
			if memberName == "<@!"+memb.User.ID+">" {
				return memberName
			}
		}

	} else {
		for _, memb := range members {
			if strings.ToLower(memb.User.Username) == memberName {
				return "<@!" + memb.User.ID + ">"
			}
		}
	}

	return ""
}

func roleExists(g *discordgo.Guild, name string) (bool, *discordgo.Role) {
	name = strings.ToLower(name)

	for _, role := range g.Roles {
		if role.Name == "@everyone" {
			continue
		}

		if strings.ToLower(role.Name) == name {
			return true, role
		}
	}

	return false, nil
}
