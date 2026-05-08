package slash

import (
	"fmt"
	"runtime/debug"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// slashHandler is the unified shape for slash and message-component handlers.
// Every handler takes (session, interaction, cfg) — even if cfg is unused —
// so a single wrap() helper covers the whole dispatch table.
type slashHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error

// errorLogger and userNotifier are dispatch-time dependencies. Production
// code injects helper.LogErrorsToErrorChannel-shaped logging and
// helper.SendResponseErrorToUser via wrap(); tests use wrapWithDeps to inject
// fakes for assertions.
//
// The logger receives the full *config.Configs so it can extract whatever
// channel ID it needs — keeping the type erasure here (rather than passing a
// pre-extracted ID) means tests can use a stub logger without constructing a
// fully-populated Configs.
type errorLogger func(s *discordgo.Session, cfg *config.Configs, err error, guildID string)
type userNotifier func(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error

// defaultErrorLogger is the production error logger — extracts the error
// channel ID from cfg and forwards to helper.LogErrorsToErrorChannel.
func defaultErrorLogger(s *discordgo.Session, cfg *config.Configs, err error, guildID string) {
	helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, guildID)
}

// wrap converts a slashHandler into the func signature the discordgo
// dispatcher expects. It logs returned errors to the configured error channel
// and recovers from panics so a single bad invocation can't crash the bot.
func wrap(h slashHandler) func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
	return wrapWithDeps(h, defaultErrorLogger, helper.SendResponseErrorToUser)
}

// wrapWithDeps is wrap() with injectable dependencies for testing.
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

// ComponentHandlers handles message components (buttons, dropdowns, etc.)
// in interactions. Keys are matched by prefix in CommandHandler.
var ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
	"horo-select":   wrap(sendHoroscopeCompResponse),
	"album-suggest": wrap(sendAlbumPickCompResponse),
	"wyr-reroll":    wrap(sendWYRrerollResp),
	"wyr-votes":     wrap(sendWYRvotesResp),
}

// CommandHandlers handles top-level slash command interactions.
var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
	"animals":   wrap(sendAnimalsResponse),
	"txt":       wrap(sendTxtResponse),
	"rate-this": wrap(sendRateThisResponse),
	"get":       wrap(sendGetResponse),
	"img-b":     wrap(sendImgResponse),
	"img-e":     wrap(sendImgResponse),
	"img-wbs":   wrap(sendImgResponse),
	"daily":     wrap(sendDailyResponse),
	"pick":      wrap(sendPickResponse),
	"tuuck":     wrap(sendTuuckResponse),
	"play":      wrap(sendPlayResponse),
}
