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
	Fields: []*discordgo.MessageEmbedField{
		{
			Name:   "New Command Group: </txt 1337:996597081085067404>",
			Value:  "Alters and makes some cool text effects.",
			Inline: false,
		},
		{
			Name:   "New Commands: </animals katz:938651875409019000>",
			Value:  "It's uhh...cats.",
			Inline: false,
		},
		{
			Name:   "New Commands: </get fake-person:941607156325703750>",
			Value:  "Generates a fake person with a name, age, and bio, etc.",
			Inline: false,
		},
		{
			Name:   "New Commands: </play just-lost:935687318281551879>",
			Value:  "New fun game to play with friends.",
			Inline: false,
		},
		{
			Name:   "New Commands: $cistercian",
			Value:  "Its very cool. Follow the link in the embed to learn more. Just try it.",
			Inline: false,
		},
		{
			Name:   "Bug Fix: Audio Not Playing",
			Value:  "Fixed a bug that wouldn't let you play audio. I honestly can't remember this, so I hope its actually fixed.",
			Inline: false,
		},
		{
			Name:   "Enhancement: Audio Not Found Message",
			Value:  "When searching for a YouTube audio to play, a \"Sorry! We couldn't find this file\" message will be sent if the audio is not found. This is to help with any confusion as to why the audio is not playing. This could also result from a timeout. Either way, you aren't getting the audio, wah wah.",
			Inline: false,
		},
		{
			Name:   "Enhancement: Repo Renamed",
			Value:  "The BuddieBot repo was renamed from \"DiscordBot\" to \"BuddieBot\" to better reflect my unique and creative naming skills.\n\n",
			Inline: false,
		},
	},
}
