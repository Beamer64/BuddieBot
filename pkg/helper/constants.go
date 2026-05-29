package helper

import (
	"github.com/bwmarrin/discordgo"
)

const (
	// PollMessageContent — the reaction handler matches messages by this body.
	PollMessageContent = "Poll Time!"

	// LmgtfyMsgEmoji triggers the "Let Me Google That For You" reaction handler.
	LmgtfyMsgEmoji    = "🔎"
	TestLmgtfyEmojiID = "lmgtfy:1509929622412984420"
	ProdLmgtfyEmojiID = "lmgtfy:1509929740180520972"

	// ErrorReaction — fallback when LogAndReact can't DM the user.
	ErrorReaction = "⚠️"

	CistercianMin = -9999
	CistercianMax = 9999
)

// GuildOnly restricts a command to the guild context, keeping it out of DMs
var GuildOnly = &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild}

// CommandExamples holds a representative usage example per top-level command,
// keyed by command name. /tuuck shows it when present; a missing entry simply
// omits the example. The command list itself comes from the live specs, so
// this is the only manually-curated piece (and forgetting an entry is harmless).
var CommandExamples = map[string]string{
	"animals":   "/animals doggo",
	"audio":     "/audio play url-1:https://youtu.be/dQw4w9WgXcQ",
	"daily":     "/daily type:horoscope",
	"game":      "/game wyr",
	"generate":  "/generate type:fake-person",
	"get":       "/get type:joke",
	"image":     "/image filter blur user:@friend",
	"pick":      "/pick choices 1st:pizza 2nd:tacos",
	"rate-this": "/rate-this type:simp user:@friend",
	"txt":       "/txt type:bubble text:hello",
	"tuuck":     "/tuuck cmd-help command:audio",
	"user":      "/user profile",
	"admin":     "/admin set-prefix new-prefix:!",
}

var CistOnes = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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

var CistTens = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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
var CistHunds = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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
var CistThous = map[string]struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
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
