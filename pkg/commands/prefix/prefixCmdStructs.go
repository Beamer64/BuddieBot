package prefix

import (
	"github.com/bwmarrin/discordgo"
)

// releaseNotesEmbed is the embed for the release notes
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
	// Fields is intended to hold the per-release highlights. Edit this
	// slice fresh for each release before running $release — the
	// dynamic command-mention IDs (`</foo:NNN>`) only render correctly
	// for the currently-registered command IDs, so don't carry stale
	// entries across releases.
	Fields: []*discordgo.MessageEmbedField{},
}
