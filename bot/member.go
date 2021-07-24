package bot

import (
	"github.com/bwmarrin/discordgo"
	"sort"
)

func memberHasRole(member *discordgo.Member, role *discordgo.Role) bool {
	memberRoles := make([]string, len(member.Roles))

	copy(memberRoles, member.Roles)

	sort.Slice(memberRoles, func(i, j int) bool {
		return memberRoles[i] < memberRoles[j]
	})

	index := sort.Search(len(memberRoles), func(i int) bool {
		return memberRoles[i] >= role.ID
	})

	return index != len(memberRoles) && memberRoles[index] == role.ID
}
