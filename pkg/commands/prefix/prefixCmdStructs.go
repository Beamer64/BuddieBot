package prefix

import (
	"github.com/bwmarrin/discordgo"
)

var releaseNotesEmbed = &discordgo.MessageEmbed{
	Title: "Release Notes!",
	URL:   "https://github.com/Beamer64/BuddieBot/blob/master/res/release.md",
	Description: "SUM MOAR BIG BOI CHANGES\n\nDetailed list can be found in the Title link above." +
		"\nCheck it out\n-----------------------------------------------------------------------------\n\n- Command changes:",
	Color: 11091696,
	Author: &discordgo.MessageEmbedAuthor{
		Name:    "",
		IconURL: "",
	},
	// Edit Fields fresh per release — `</foo:NNN>` command mentions only
	// resolve against currently-registered command IDs.
	Fields: []*discordgo.MessageEmbedField{},
}
