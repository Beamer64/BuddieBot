package slash

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// profilePageCount is the number of profile pages currently in rotation. The
// pagination plumbing (custom IDs, button render) is fully in place; adding a
// new page is mostly a new switch case in renderProfilePage.
const profilePageCount = 1

// recentRatingsShown is how many of the user's most recently updated ratings
// appear in the top-of-profile corner block.
const recentRatingsShown = 3

// sendUserResponse dispatches /user subcommands. It does NOT defer up-front —
// each subcommand owns its response type (profile defers a public embed;
// forget-me sends an immediate ephemeral confirmation).
func sendUserResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// With subcommands, Discord always sends exactly one subcommand option.
	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "profile":
		// Target defaults to the invoker; the optional 'target' arg overrides it.
		target := i.Member.User
		for _, opt := range sub.Options {
			if opt.Name == "target" {
				if u := opt.UserValue(s); u != nil {
					target = u
				}
			}
		}
		return sendProfileResponse(s, i, cfg, target)
	case "forget-me":
		return userForgetMePrompt(s, i)
	default:
		return helper.ReturnUserError(s, i, "Unknown user subcommand.", fmt.Errorf("unknown user subcommand: %s", sub.Name))
	}
}

func sendProfileResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, target *discordgo.User) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /user profile: %w", err)
	}

	// Bots don't get materialized in the User table — short-circuit with a
	// friendly stub so the command still "works" against them without writing.
	if target.Bot {
		embed := botProfileEmbed(target)
		embeds := []*discordgo.MessageEmbed{embed}
		if _, err := s.InteractionResponseEdit(
			i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds},
		); err != nil {
			return fmt.Errorf("send /user profile bot stub: %w", err)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embed, components, err := profilePage(ctx, cfg.DB, i.GuildID, target, i.Member.User.ID, 0)
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Couldn't load that profile.", err)
	}

	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
		},
	); err != nil {
		return fmt.Errorf("send /user response: %w", err)
	}
	return nil
}

// botProfileEmbed is the stub shown when /user profile targets a bot. We
// don't create a User row for bots, so there's no DB-backed profile to render.
func botProfileEmbed(target *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       target.Username + "'s Profile",
		Description: "Bots don't have BuddieBot profiles.",
		Color:       helper.RandomDiscordColor(),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: target.AvatarURL("128")},
	}
}

// sendProfilePageResponse handles Prev/Next clicks on a /user profile embed.
// The custom ID carries the target's Discord ID, the original invoker's
// Discord ID (for ownership check), and the page to render — so the handler
// is fully self-contained.
func sendProfilePageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	targetID, invokerID, page := parseProfilePageID(i.MessageComponentData().CustomID)

	// Ownership: only the user who invoked /user profile can flip pages.
	if i.Member == nil || i.Member.User.ID != invokerID {
		return helper.SendEphemeralError(s, i, "That isn't your profile to flip.")
	}

	// s.User caches when possible; if Discord refuses (rare), fall back to a
	// minimal stub so we still re-render something rather than 500ing.
	target, err := s.User(targetID)
	if err != nil || target == nil {
		target = &discordgo.User{ID: targetID, Username: "Unknown User"}
	}

	// Bot stubs never get pagination buttons rendered against them, but stay
	// defensive in case a stale message somehow drives a click here.
	if target.Bot {
		return s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{botProfileEmbed(target)},
					Components: []discordgo.MessageComponent{},
				},
			},
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embed, components, err := profilePage(ctx, cfg.DB, i.GuildID, target, invokerID, page)
	if err != nil {
		return helper.SendEphemeralError(s, i, "Couldn't reload that profile page.")
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

// parseProfilePageID splits "profile-page:<targetID>:<invokerID>:<page>".
// Returns zeroes on malformed input; the caller's ownership check then fails
// naturally.
func parseProfilePageID(customID string) (target, invoker string, page int) {
	parts := strings.Split(strings.TrimPrefix(customID, "profile-page:"), ":")
	if len(parts) < 3 {
		return "", "", 0
	}
	page, _ = strconv.Atoi(parts[2])
	return parts[0], parts[1], page
}

// profilePage renders one page of /user profile plus its pagination buttons.
// It ensures the target's User row exists (so a freshly-rated person you've
// never interacted with still has data to show) and returns the assembled
// embed + components. Buttons only render when there's more than one page.
func profilePage(
	ctx context.Context,
	db *database.DB,
	discordGuildID string,
	target *discordgo.User,
	invokerID string,
	page int,
) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	if page < 0 {
		page = 0
	}
	if page >= profilePageCount {
		page = profilePageCount - 1
	}

	user, err := db.EnsureUser(ctx, discordGuildID, target.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("profile ensure user: %w", err)
	}

	embed, err := renderProfilePage(ctx, db, target, user, page)
	if err != nil {
		return nil, nil, err
	}

	components := profilePageButtons(target.ID, invokerID, page, profilePageCount)
	return embed, components, nil
}

// renderProfilePage builds the embed body for the given page index. Today
// there's only page 0 (Basics + corner-3); future pages slot in as new cases.
func renderProfilePage(
	ctx context.Context,
	db *database.DB,
	target *discordgo.User,
	user *database.User,
	page int,
) (*discordgo.MessageEmbed, error) {
	embed := &discordgo.MessageEmbed{
		Title:     target.Username + "'s Profile",
		Color:     helper.RandomDiscordColor(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: target.AvatarURL("128")},
		Footer:    &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Page %d of %d", page+1, profilePageCount)},
	}

	switch page {
	case 0:
		fields, err := profileBasicsFields(ctx, db, user)
		if err != nil {
			return nil, err
		}
		embed.Fields = fields
	}
	return embed, nil
}

// profileBasicsFields composes the Basics page: a top-of-profile inline block
// of the user's most-recently-updated ratings, followed by Dosh, command-usage
// stats, the Day-One badge when present, and BB User Since.
func profileBasicsFields(ctx context.Context, db *database.DB, user *database.User) ([]*discordgo.MessageEmbedField, error) {
	recent, err := db.GetRecentUserRatings(ctx, user.ID, recentRatingsShown)
	if err != nil {
		return nil, fmt.Errorf("recent ratings: %w", err)
	}

	total, err := db.GetUserCommandTotal(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("command total: %w", err)
	}
	topName, topCount, err := db.GetUserTopCommand(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("top command: %w", err)
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Recent Ratings", Value: formatRecentRatings(recent), Inline: true},
		{Name: " ", Value: " ", Inline: true},
		{Name: "Dosh", Value: fmt.Sprintf("`🪙 %d`", user.Dosh), Inline: true},
		{Name: "Commands", Value: formatCommandStats(total, topName, topCount), Inline: false},
	}
	if user.IsDayOne {
		fields = append(
			fields, &discordgo.MessageEmbedField{
				Name: "Day One", Value: "Original tester", Inline: true,
			},
		)
	}
	fields = append(
		fields, &discordgo.MessageEmbedField{
			Name: "BB User Since", Value: memberSinceDisplay(user.CreatedAt), Inline: false,
		},
	)
	return fields, nil
}

// formatRecentRatings renders a stack of "Label: value" lines for the top
// inline field. "—" stands in when the user has no ratings yet (which only
// happens if seeding failed; new rows get 3 seeded values).
func formatRecentRatings(ratings []*database.UserRating) string {
	if len(ratings) == 0 {
		return "—"
	}
	lines := make([]string, 0, len(ratings))
	for _, r := range ratings {
		lines = append(lines, formatRecentRatingLine(r))
	}
	return strings.Join(lines, "\n")
}

// formatCommandStats renders the Commands field: a Total line and a
// Most-used line. topName is the storage key (e.g. "image filter blur");
// we prepend "/" so it reads as a command. Empty topName means the user
// has nothing tracked yet — render an em-dash in that slot.
func formatCommandStats(total int64, topName string, topCount int64) string {
	most := "—"
	if topName != "" {
		most = fmt.Sprintf("`/%s` (`%d×`)", topName, topCount)
	}
	return fmt.Sprintf("**Total**: `%d`\n**Most used**: %s", total, most)
}

// formatRecentRatingLine renders one corner-block line. Schmeat keeps its
// ASCII-strip style; everything else is "Pretty Label: X/100".
func formatRecentRatingLine(r *database.UserRating) string {
	if r.RatingName == "schmeat" {
		return fmt.Sprintf("**Schmeat**: `%s`", helper.SchmeatString(r.Value))
	}
	if pretty, ok := standardRatings[r.RatingName]; ok {
		return fmt.Sprintf("**%s**: `%d/100`", pretty.ScoreLabel, r.Value)
	}
	return fmt.Sprintf("**%s**: `%d`", r.RatingName, r.Value)
}

func memberSinceDisplay(createdAt string) string {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, createdAt); err == nil {
			return fmt.Sprintf("<t:%d:D>", t.Unix())
		}
	}
	if len(createdAt) >= 10 {
		return createdAt[:10]
	}
	return createdAt
}

// profilePageButtons returns Prev/Next buttons for the given page, or nil
// when the profile has only one page
func profilePageButtons(targetID, invokerID string, page, totalPages int) []discordgo.MessageComponent {
	if totalPages <= 1 {
		return nil
	}
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Prev",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("profile-page:%s:%s:%d", targetID, invokerID, page-1),
					Disabled: page <= 0,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("profile-page:%s:%s:%d", targetID, invokerID, page+1),
					Disabled: page >= totalPages-1,
				},
			},
		},
	}
}

// userForgetMePrompt sends an ephemeral confirmation with Confirm/Cancel buttons.
func userForgetMePrompt(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "This permanently deletes **all** your BuddieBot data across **every** server (balance, profile, everything). This can't be undone.\n\nAre you sure?",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Delete my data",
								Style:    discordgo.DangerButton,
								CustomID: "forget-me-confirm:" + i.Member.User.ID,
							},
							discordgo.Button{
								Label:    "Cancel",
								Style:    discordgo.SecondaryButton,
								CustomID: "forget-me-cancel:" + i.Member.User.ID,
							},
						},
					},
				},
			},
		},
	); err != nil {
		return fmt.Errorf("send forget-me prompt: %w", err)
	}
	return nil
}

func forgetMeConfirm(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	wantID := strings.TrimPrefix(i.MessageComponentData().CustomID, "forget-me-confirm:")
	// Ephemeral already scopes the buttons to the invoker; this is belt-and-suspenders.
	if i.Member == nil || i.Member.User.ID != wantID {
		return helper.SendEphemeralError(s, i, "That confirmation isn't yours.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n, err := cfg.DB.ForgetUser(ctx, wantID)
	if err != nil {
		return helper.ReturnUserError(s, i, "Couldn't delete your data — try again.", fmt.Errorf("forget user: %w", err))
	}

	msg := "Your BuddieBot data has been deleted. You'll be added fresh next time you use a command."
	if n == 0 {
		msg = "You had no stored data — nothing to delete."
	}
	return updateForgetMeMessage(s, i, msg)
}

func forgetMeCancel(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	return updateForgetMeMessage(s, i, "Cancelled — your data is untouched.")
}

// updateForgetMeMessage replaces the ephemeral confirmation with a result
// and strips the buttons (empty, non-nil slice clears them).
func updateForgetMeMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    content,
				Components: []discordgo.MessageComponent{},
			},
		},
	); err != nil {
		return fmt.Errorf("update forget-me message: %w", err)
	}
	return nil
}

func userSpec() *discordgo.ApplicationCommand {
	// only InteractionContextGuild keeps the command out of DMs
	return &discordgo.ApplicationCommand{
		Name:        "user",
		Description: "User profile and account commands",
		Contexts:    helper.GuildOnly,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "profile",
				Description: "View your BuddieBot profile (or someone else's)",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "target",
						Description: "Whose profile to view (defaults to you)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "forget-me",
				Description: "Delete all your BuddieBot data (asks for confirmation first)",
			},
		},
	}
}
