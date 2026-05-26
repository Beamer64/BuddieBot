package slash

import (
	"fmt"
	"runtime/debug"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// slashHandler is the shared shape for slash and message-component handlers.
type slashHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error

// errorLogger and userNotifier are dispatch-time deps; tests inject fakes
// via wrapWithDeps. Logger takes the full *config.Configs so tests can stub
// without populating ErrorLogChannelID.
type errorLogger func(s *discordgo.Session, cfg *config.Configs, err error, guildID string)
type userNotifier func(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error

func defaultErrorLogger(s *discordgo.Session, cfg *config.Configs, err error, guildID string) {
	helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, err, guildID)
}

// wrap adapts a slashHandler to discordgo's dispatcher signature, logs
// returned errors to the error channel, and recovers from panics.
func wrap(h slashHandler) func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
	return wrapWithDeps(h, defaultErrorLogger, helper.SendEphemeralError)
}

func wrapWithDeps(h slashHandler, logErr errorLogger, notifyUser userNotifier) func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in command handler: %v\n%s", r, debug.Stack())
				logErr(s, cfg, err, i.GuildID)
				_ = notifyUser(s, i, "Something went wrong. The error has been logged.")
			}
		}()

		if err := h(s, i, cfg); err != nil {
			logErr(s, cfg, err, i.GuildID)
		}
	}
}

// ComponentHandlers — keys are matched by prefix by the event handler.
var ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
	"horo-select": wrap(sendHoroscopeCompResponse),
	"wyr-reroll":  wrap(sendWYRrerollResp),
	"wyr-votes":   wrap(sendWYRvotesResp),
}

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
	"animals":   wrap(sendAnimalsResponse),
	"txt":       wrap(sendTxtResponse),
	"rate-this": wrap(sendRateThisResponse),
	"get":       wrap(sendGetResponse),
	"image":     wrap(sendImgResponse),
	"daily":     wrap(sendDailyResponse),
	"pick":      wrap(sendPickResponse),
	"tuuck":     wrap(sendTuuckResponse),
	"game":      wrap(sendGameResponse),
	"generate":  wrap(sendGenerateResponse),
	"audio":     wrap(sendAudioResponse),
}
