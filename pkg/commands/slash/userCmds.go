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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First view of a (user, guild) pair creates the row here.
	user, err := cfg.DB.EnsureUser(ctx, i.GuildID, target.ID)
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Couldn't load that profile.", fmt.Errorf("profile ensure user: %w", err))
	}

	// SQLite CURRENT_TIMESTAMP is "YYYY-MM-DD HH:MM:SS" in UTC
	memberSince := user.CreatedAt
	if t, parseErr := time.Parse("2006-01-02", user.CreatedAt); parseErr == nil {
		memberSince = fmt.Sprintf("<t:%d:D>", t.Unix())
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Dosh", Value: fmt.Sprintf("%d", user.Dosh), Inline: true},
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
			Name: "Member Since", Value: memberSince, Inline: false,
		},
	)

	embed := &discordgo.MessageEmbed{
		Title:     target.Username + "'s Profile",
		Color:     helper.RandomDiscordColor(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: target.AvatarURL("128")},
		Fields:    fields,
	}

	embeds := []*discordgo.MessageEmbed{embed}
	if _, err = s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds},
	); err != nil {
		return fmt.Errorf("send /user response: %w", err)
	}

	return nil
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
