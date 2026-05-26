package slash

import (
	"github.com/bwmarrin/discordgo"
)

// Commands assembles every registered slash command. Specs live next to
// their handlers (animalsSpec in animalCmds.go, etc.). All commands and
// options must have a description — Discord rejects registration otherwise.
var Commands = []*discordgo.ApplicationCommand{
	animalsSpec(),
	rateThisSpec(),
	getSpec(),
	imageSpec(),
	dailySpec(),
	pickSpec(),
	gameSpec(),
	txtSpec(),
	tuuckSpec(),
	generateSpec(),
	audioSpec(),
}
