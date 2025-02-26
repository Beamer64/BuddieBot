package commands

import (
	"github.com/bwmarrin/discordgo"
)

// ReleaseNotesEmbed is the embed for the release notes
var ReleaseNotesEmbed = &discordgo.MessageEmbed{
	Title: "Release Notes!",
	URL:   "https://github.com/Beamer64/BuddieBot/blob/master/res/release.md",
	Description: "SUM MOAR BIG BOI CHANGES\n\nDetailed list can be found in the Title link above." +
		"\nCheck it out\n-----------------------------------------------------------------------------\n\n- Command changes:",
	Color: 11091696,
	Fields: []*discordgo.MessageEmbedField{
		{
			Name:   "New Command Group: /ratethis",
			Value:  "Give/Get some new ratings",
			Inline: false,
		},
		{
			Name:   "New Commands: /pick album",
			Value:  "<@!282722418093719556>'s Album recommender api. Recommends a music album based on liked tags.",
			Inline: false,
		},
		{
			Name:   "New Commands: /pick poll",
			Value:  "Poll comman..for polling things..",
			Inline: false,
		},
		{
			Name:   "New Commands: ${COMMAND} SpongeBob easter egg",
			Value:  "It's my bot, I can do what I want.",
			Inline: false,
		},
		{
			Name:   "Bug Fix: Youtube mobile links",
			Value:  "(When working..) Audio will play with the mobile link 'm.youtube.com...'",
			Inline: false,
		},
		{
			Name:   "Enhancement: Audio Queue",
			Value:  "The Audio Queue will show a cleaned title from the old 'Name-Title-Sum_Numbers.mp3'.",
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

type imgBBData struct {
	Data    imgData `json:"data"`
	Success bool    `json:"success"`
	Status  int     `json:"status"`
}
type imageInfo struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type thumb struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type medium struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type imgData struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	URLViewer  string    `json:"url_viewer"`
	URL        string    `json:"url"`
	DisplayURL string    `json:"display_url"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Size       int       `json:"size"`
	Time       int       `json:"time"`
	Expiration int       `json:"expiration"`
	Image      imageInfo `json:"image"`
	Thumb      thumb     `json:"thumb"`
	Medium     medium    `json:"medium"`
	DeleteURL  string    `json:"delete_url"`
}
