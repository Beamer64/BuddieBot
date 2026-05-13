package slash

import (
	"github.com/bwmarrin/discordgo"
)

// Commands assembles every registered slash command. Each command's spec
// (a *discordgo.ApplicationCommand) lives next to its handler in the same
// per-feature file (e.g. animalsSpec is in animalCmds.go).
//
// All commands and options must have a description — Discord rejects the
// registration otherwise.
var Commands = []*discordgo.ApplicationCommand{
	animalsSpec(),
	rateThisSpec(),
	getSpec(),
	imageSpec(),
	dailySpec(),
	pickSpec(),
	playSpec(),
	txtSpec(),
	tuuckSpec(),
}
