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
		return helper.ReturnUserErrorDeferred(s, i, "Unknown admin subcommand.", fmt.Errorf("unknown admin subcommand: %s", sub.Name))
	}
}

func adminSetPrefix(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	// new-prefix is optional; omitting (or blanking) it resets to the default.
	newPrefix := ""
	for _, opt := range opts {
		if opt.Name == "new-prefix" {
			newPrefix = strings.TrimSpace(opt.StringValue())
		}
	}

	if newPrefix != "" {
		if len(newPrefix) > 5 {
			return helper.ReturnUserErrorDeferred(s, i, "Prefix must be 5 characters or fewer.", nil)
		}
		if strings.ContainsAny(newPrefix, " \t\n") {
			return helper.ReturnUserErrorDeferred(s, i, "Prefix can't contain spaces.", nil)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cfg.DB.SetGuildPrefixOverride(ctx, i.GuildID, newPrefix); err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Couldn't update the prefix.", fmt.Errorf("set prefix: %w", err))
	}

	content := fmt.Sprintf("Prefix set to `%s` for this server.", newPrefix)
	if newPrefix == "" {
		content = "Prefix reset to the default `$`."
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}); err != nil {
		return fmt.Errorf("send /admin set-prefix response: %w", err)
	}
	return nil
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
				Description: "Set this server's $-command prefix (omit to reset to default)",
				Options: []*discordgo.ApplicationCommandOption{
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
