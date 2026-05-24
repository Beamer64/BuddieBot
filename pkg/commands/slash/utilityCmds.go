package slash

import (
	"fmt"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendTuuckResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return sendTuuckCommands(s, i)
	}

	cmdName := strings.TrimSpace(options[0].StringValue())
	cmdName = strings.TrimPrefix(cmdName, "/")

	spec := findCommandSpec(cmdName)
	if spec == nil {
		return helper.ReturnUserErrorDeferred(s, i, fmt.Sprintf("Invalid command: %s", cmdName), nil)
	}

	embed := &discordgo.MessageEmbed{
		Title: "/" + spec.Name,
		Color: helper.RandomDiscordColor(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Description",
				Value:  spec.Description,
				Inline: false,
			},
		},
	}
	if usage := buildTuuckUsage(spec); usage != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Subcommands / options",
			Value:  usage,
			Inline: false,
		})
	}

	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		},
	); err != nil {
		return fmt.Errorf("send /tuuck %s response: %w", cmdName, err)
	}
	return nil
}

// sendTuuckCommands lists every registered top-level slash command from
// the in-memory Commands slice — no more reflection over a yaml-loaded
// struct that drifts every time a command is renamed.
func sendTuuckCommands(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var content strings.Builder
	content.WriteString("A list of current slash commands:\n```\n")
	for _, spec := range Commands {
		fmt.Fprintf(&content, "/%-12s — %s\n", spec.Name, spec.Description)
	}
	content.WriteString("```\nUse `/tuuck <command-name>` for more detail on a specific command.")

	contentStr := content.String()
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &contentStr,
		},
	); err != nil {
		return fmt.Errorf("send /tuuck commands list: %w", err)
	}
	return nil
}

// findCommandSpec returns the ApplicationCommand whose Name matches
// (case-insensitively), or nil if none.
func findCommandSpec(name string) *discordgo.ApplicationCommand {
	for _, spec := range Commands {
		if strings.EqualFold(spec.Name, name) {
			return spec
		}
	}
	return nil
}

// buildTuuckUsage renders the spec's subcommands or top-level options
// as a bulleted multi-line string. Returns "" if the command has no
// options worth showing.
func buildTuuckUsage(spec *discordgo.ApplicationCommand) string {
	if len(spec.Options) == 0 {
		return ""
	}
	var b strings.Builder
	for _, opt := range spec.Options {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommand,
			discordgo.ApplicationCommandOptionSubCommandGroup:
			fmt.Fprintf(&b, "• `%s` — %s\n", opt.Name, opt.Description)
		default:
			req := ""
			if opt.Required {
				req = " (required)"
			}
			fmt.Fprintf(&b, "• `%s`%s — %s\n", opt.Name, req, opt.Description)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func tuuckSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "tuuck",
		Description: "I've fallen and can't get up!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "command",
				Description: "Specify a command for a description",
				Required:    false,
			},
		},
	}
}
