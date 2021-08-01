package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func memberHasRole(member *discordgo.Member, roleName string) bool {
	//memberRoles := make([]string, len(member.Roles))

	/*for _, role := range member.Roles {
		if role == "@everyone" {
			continue
		}

		if strings.ToLower(role) == roleName {
			return true
		}
	}*/
	return true
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
