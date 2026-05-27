package slash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendTuuckResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "cmd-list":
		return tuuckCmdList(s, i, sub.Options)
	default:
		return helper.ReturnUserError(s, i, "Unknown tuuck subcommand.", fmt.Errorf("unknown tuuck subcommand: %s", sub.Name))
	}
}

// tuuckCmdList opens an ephemeral, paginated command list — a compact
// name + description per command, several per page. Content is derived from
// the live specs, so there's nothing to maintain by hand. Rendering is
// instant, so no defer is needed.
func tuuckCmdList(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	// Optional command arg picks the scope: empty = the top-level list, a
	// command name = that command's subcommands/options.
	scope := ""
	for _, opt := range opts {
		if opt.Name == "command" {
			scope = strings.TrimPrefix(strings.TrimSpace(opt.StringValue()), "/")
		}
	}

	embed, components, ok := tuuckPage(scope, 0)
	if !ok {
		return helper.ReturnUserError(s, i, fmt.Sprintf("Unknown command: %q. Use /tuuck cmd-list (no command) to browse them.", scope), nil)
	}
	return s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:      discordgo.MessageFlagsEphemeral,
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		},
	)
}

// sendTuuckPageResponse handles the Prev/Next buttons. The custom ID carries
// both the scope and the target page (tuuck-page:<scope>:<page>), so it
// re-renders the right list and edits the (ephemeral) message in place.
func sendTuuckPageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	scope, page := splitScopePage(i.MessageComponentData().CustomID)
	embed, components, ok := tuuckPage(scope, page)
	if !ok {
		// Scope's command no longer exists (removed since the message was sent);
		// fall back to the top-level list.
		embed, components, _ = tuuckPage("", 0)
	}
	return s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		},
	)
}

// splitScopePage parses "tuuck-page:<scope>:<page>". scope may be empty (the
// top-level list) and never contains a colon, so the last colon delimits page.
func splitScopePage(customID string) (scope string, page int) {
	rest := strings.TrimPrefix(customID, "tuuck-page:")
	idx := strings.LastIndex(rest, ":")
	if idx < 0 {
		return "", 0
	}
	page, _ = strconv.Atoi(rest[idx+1:])
	return rest[:idx], page
}

// tuuckFieldsPerPage caps entries per cmd-list page at Discord's hard limit of
// 25 embed fields. Keeps page counts low while staying within the API cap.
const tuuckFieldsPerPage = 25

// tuuckListItem is one entry on a cmd-list page: a field name + value.
type tuuckListItem struct {
	name  string
	value string
}

// tuuckListItems returns the title and entries for a cmd-list scope. An empty
// scope is the top-level command list; a command-name scope is that command's
// subcommands (flattened across groups) or, for a type-choice command, its
// choices (with the invocation as the value, since choices carry no description).
func tuuckListItems(scope string) (title string, items []tuuckListItem, ok bool) {
	if scope == "" {
		items = make([]tuuckListItem, 0, len(Commands))
		for _, spec := range Commands {
			items = append(items, tuuckListItem{name: "/" + spec.Name, value: spec.Description})
		}
		return "Available Commands", items, true
	}

	spec := findCommandSpec(scope)
	if spec == nil {
		return "", nil, false
	}

	for _, opt := range spec.Options {
		switch {
		case opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup:
			for _, sub := range opt.Options {
				items = append(
					items, tuuckListItem{
						name:  opt.Name + " " + sub.Name,
						value: orPlaceholder(sub.Description),
					},
				)
			}
		case opt.Type == discordgo.ApplicationCommandOptionSubCommand:
			items = append(
				items, tuuckListItem{
					name:  opt.Name,
					value: orPlaceholder(opt.Description),
				},
			)
		case len(opt.Choices) > 0:
			for _, c := range opt.Choices {
				items = append(
					items, tuuckListItem{
						name:  c.Name,
						value: fmt.Sprintf("`/%s %s:%v`", spec.Name, opt.Name, c.Value),
					},
				)
			}
		}
	}
	return "/" + spec.Name + " — options", items, true
}

func orPlaceholder(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// tuuckPage renders one page of a cmd-list scope as inline fields, plus its
// Prev/Next buttons (custom IDs carry the scope + target page). Content and
// page count derive from the live specs, so nothing needs hand-maintaining.
// Buttons are omitted when everything fits on a single page. ok is false for
// an unknown command scope.
func tuuckPage(scope string, page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent, bool) {
	title, items, ok := tuuckListItems(scope)
	if !ok {
		return nil, nil, false
	}

	totalPages := (len(items) + tuuckFieldsPerPage - 1) / tuuckFieldsPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * tuuckFieldsPerPage
	end := start + tuuckFieldsPerPage
	if end > len(items) {
		end = len(items)
	}

	fields := make([]*discordgo.MessageEmbedField, 0, end-start)
	for _, it := range items[start:end] {
		fields = append(
			fields, &discordgo.MessageEmbedField{
				Name:   it.name,
				Value:  it.value,
				Inline: true,
			},
		)
	}

	embed := &discordgo.MessageEmbed{
		Title:  title,
		Color:  helper.RandomDiscordColor(),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{Text: tuuckFooter(scope, page+1, totalPages)},
	}
	if scope != "" {
		if spec := findCommandSpec(scope); spec != nil {
			if ex, ok := helper.CommandExamples[spec.Name]; ok && ex != "" {
				embed.Description = "Example: `" + ex + "`"
			}
		}
	}
	if embed.Description == "" && len(fields) == 0 {
		embed.Description = "No subcommands."
	}

	var components []discordgo.MessageComponent
	if totalPages > 1 {
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Prev",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("tuuck-page:%s:%d", scope, page-1),
						Disabled: page <= 0,
					},
					discordgo.Button{
						Label:    "Next",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("tuuck-page:%s:%d", scope, page+1),
						Disabled: page >= totalPages-1,
					},
				},
			},
		}
	}
	return embed, components, true
}

// tuuckFooter hints at the next drill-down step per scope.
func tuuckFooter(scope string, page, total int) string {
	if scope == "" {
		return fmt.Sprintf("Page %d of %d  •  /tuuck cmd-list command:<name> to drill into one", page, total)
	}
	return fmt.Sprintf("Page %d of %d  •  /tuuck cmd-list to see all commands", page, total)
}

// findCommandSpec matches Name case-insensitively. Returns nil if none.
func findCommandSpec(name string) *discordgo.ApplicationCommand {
	for _, spec := range Commands {
		if strings.EqualFold(spec.Name, name) {
			return spec
		}
	}
	return nil
}

func tuuckSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "tuuck",
		Description: "Help — browse commands and see how to use them",
		Contexts:    helper.GuildOnly,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "cmd-list",
				Description: "Browse every command",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "command",
						Description: "Command name (e.g. audio, image, daily)",
						Required:    false,
					},
				},
			},
		},
	}
}
