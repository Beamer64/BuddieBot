package slash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendAdminResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Admin actions are ephemeral — no need to broadcast config changes.
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /admin: %w", err)
	}

	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "set-prefix":
		return adminSetPrefix(s, i, cfg, sub.Options)
	default:
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Unknown admin subcommand.", fmt.Errorf("unknown admin subcommand: %s", sub.Name))
	}
}

func adminSetPrefix(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	// new-prefix is optional; omitting (or blanking) it resets to the default.
	newPrefix := ""
	trailingSpace := false
	for _, opt := range opts {
		switch opt.Name {
		case "new-prefix":
			newPrefix = strings.TrimLeft(opt.StringValue(), " ")
		case "trailing-space":
			trailingSpace = opt.BoolValue()
		}
	}

	// Discord strips trailing whitespace from String options at the client
	// layer, so a "plz " style prefix can't arrive in newPrefix directly. The
	// trailing-space boolean is the user's opt-in to that mode — append the
	// space here, BEFORE validation, so the 5-char cap counts it.
	if trailingSpace && newPrefix != "" && !strings.HasSuffix(newPrefix, " ") {
		newPrefix += " "
	}

	if newPrefix != "" {
		// 5-char max applies to the full prefix INCLUDING trailing spaces —
		// keeps the rendered prefix short regardless of style choice.
		if len(newPrefix) > 5 {
			return helper.EditMsgPostDeferred(s, i, "Prefix must be 5 characters or fewer.")
		}
		if strings.ContainsAny(newPrefix, "\t\n\r") {
			return helper.EditMsgPostDeferred(s, i, "Prefix can't contain tabs or newlines.")
		}
		// Trailing spaces are allowed — e.g. setting "plz " lets users
		// type "plz roman" naturally. Embedded spaces are not.
		if strings.Contains(strings.TrimRight(newPrefix, " "), " ") {
			return helper.EditMsgPostDeferred(s, i, "Spaces are only allowed at the end of the prefix.")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cfg.DB.SetGuildPrefixOverride(ctx, i.GuildID, newPrefix); err != nil {
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Couldn't update the prefix.", fmt.Errorf("set prefix: %w", err))
	}

	content := fmt.Sprintf("Prefix set to `%s` for this server.", newPrefix)
	if newPrefix == "" {
		content = "Prefix reset to the default `$`."
		newPrefix = "$"
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}); err != nil {
		return fmt.Errorf("send /admin set-prefix response: %w", err)
	}

	if err := sendPrefixUpdateMsg(s, i, newPrefix); err != nil {
		return fmt.Errorf("send /admin sendPrefixUpdateMsg: %w", err)
	}

	return nil
}

func sendPrefixUpdateMsg(s *discordgo.Session, i *discordgo.InteractionCreate, newPrefix string) error {
	// Discord visually collapses trailing whitespace inside single backticks,
	// so when the prefix ends in a space the rendered `plz ` looks identical
	// to `plz`. Call that out in plain text so members know to type the space.
	extra := ""
	if strings.HasSuffix(newPrefix, " ") {
		extra = " (the trailing space is required)"
	}

	desc := fmt.Sprintf(
		"A server admin has updated the BuddieBot command prefix to `%s`%s.\nExample Usage: `%sgoodboy`",
		newPrefix, extra, newPrefix,
	)

	_, err := s.ChannelMessageSendComplex(
		i.ChannelID, &discordgo.MessageSend{
			Content: "@everyone",
			Embed: &discordgo.MessageEmbed{
				Color:       helper.RandomDiscordColor(),
				Title:       "(°ロ°)👇",
				Description: desc,
			},
		},
	)

	return err
}

func adminSpec() *discordgo.ApplicationCommand {
	perm := int64(discordgo.PermissionManageGuild)
	return &discordgo.ApplicationCommand{
		Name:                     "admin",
		Description:              "Server-admin configuration commands",
		Contexts:                 helper.GuildOnly,
		DefaultMemberPermissions: &perm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set-prefix",
				Description: "Set this server's $-command prefix. Call this where everyone can see.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "trailing-space",
						Description: "Require a space between prefix and command? (e.g. 'plz roman') Off by default.",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "new-prefix",
						Description: "New prefix; leave empty to reset to the default ($)",
						Required:    false,
					},
				},
			},
		},
	}
}
