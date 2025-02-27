package commands

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
var ones = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 20, 140, 20},
	"2": {100, 60, 140, 60},
	"3": {100, 20, 140, 60},
	"4": {100, 60, 140, 20},
	"5": {100, 20, 140, 20},
	"6": {140, 20, 140, 60},
	"7": {100, 20, 140, 20},
	"8": {100, 60, 140, 60},
	"9": {100, 20, 140, 20},
}

var tens = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 20, 60, 20},
	"2": {100, 60, 60, 60},
	"3": {100, 20, 60, 60},
	"4": {100, 60, 60, 20},
	"5": {100, 20, 60, 20},
	"6": {60, 20, 60, 60},
	"7": {100, 20, 60, 20},
	"8": {100, 60, 60, 60},
	"9": {100, 20, 60, 20},
}
var hunds = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 180, 140, 180},
	"2": {100, 140, 140, 140},
	"3": {100, 180, 140, 140},
	"4": {100, 140, 140, 180},
	"5": {100, 180, 140, 180},
	"6": {140, 180, 140, 140},
	"7": {100, 180, 140, 180},
	"8": {100, 140, 140, 140},
	"9": {100, 180, 140, 180},
}
var thous = map[string]struct {
	x1 int
	y1 int
	x2 int
	y2 int
}{
	"1": {100, 180, 60, 180},
	"2": {100, 140, 60, 140},
	"3": {100, 180, 60, 140},
	"4": {100, 140, 60, 180},
	"5": {100, 180, 60, 180},
	"6": {60, 180, 60, 140},
	"7": {100, 180, 60, 180},
	"8": {100, 140, 60, 140},
	"9": {100, 180, 60, 180},
}
