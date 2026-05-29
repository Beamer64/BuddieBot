package slash

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendRateThisResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	optMap := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, opt := range i.ApplicationCommandData().Options {
		optMap[opt.Name] = opt
	}
	cmdType := optMap["type"].StringValue()

	// The target defaults to the invoker (always human — bots can't run slash
	// commands); the optional 'user' arg redirects the rating at someone else,
	// which is where a bot can sneak in.
	target := i.Member.User
	if userOpt, ok := optMap["user"]; ok {
		if u := userOpt.UserValue(s); u != nil {
			target = u
		}
	}
	userMention := fmt.Sprintf("<@!%s>", target.ID)

	embed, value := getRateThisEmbed(cmdType, userMention)

	// Persist the rating against the *target* user, but only for humans —
	// we don't materialize bot rows in the DB. DB hiccups shouldn't kill the
	// embed; log locally and let the rating still display.
	if !target.Bot {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if u, dbErr := cfg.DB.EnsureUser(ctx, i.GuildID, target.ID); dbErr != nil {
			log.Printf("rate-this ensure user %s in guild %s: %v", target.ID, i.GuildID, dbErr)
		} else if u != nil {
			if dbErr := cfg.DB.SetUserRating(ctx, u.ID, cmdType, value); dbErr != nil {
				log.Printf("rate-this set rating %s for user %d: %v", cmdType, u.ID, dbErr)
			}
		}
	}

	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		},
	); err != nil {
		return fmt.Errorf("send /rate-this %s response: %w", cmdType, err)
	}
	return nil
}

// getRateThisEmbed generates the value for this rating, formats it for display,
// and wraps it in an embed. Returning the raw value lets the caller persist it.
func getRateThisEmbed(ratingName, user string) (*discordgo.MessageEmbed, int) {
	value := helper.RandomRatingValue(ratingName)
	title, desc := formatRating(ratingName, user, value)
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       helper.RandomDiscordColor(),
	}, value
}

// standardRatings maps the rating's option name to its human-readable label
// pair: the embed title and the score label. Schmeat is handled separately
// because it has no /100 score and renders an ASCII strip instead.
var standardRatings = map[string]struct{ Title, ScoreLabel string }{
	"simp":       {"Rate This Simp", "Simp"},
	"dank":       {"Dank Rating", "Dank"},
	"epic-gamer": {"Rate This Epic Gamer", "Epic Gamer"},
	"gay":        {"Gay Rating", "Gay"},
	"stinky":     {"Rate This Stinky", "Stinky"},
	"thot":       {"Rate This Thot", "Thot"},
	"pickme":     {"Rate This Pick-Me", "Pick-Me"},
	"neckbeard":  {"Rate This Neck Beard", "Neck Beard"},
	"looks":      {"Rate These Looks", "Looks"},
	"smarts":     {"Rate These Smarts", "Smarts"},
	"nerd":       {"Rate This Nerd", "Nerd"},
	"geek":       {"Rate This Geek", "Geek"},
	"npc":        {"Rate This NPC", "NPC"},
}

// formatRating renders a rating's title + description for the /rate-this
// embed. Empty strings come back for unknown rating types — caller decides
// whether that's worth surfacing (it shouldn't happen, since the option's
// choices are fixed in the spec).
func formatRating(ratingName, user string, value int) (string, string) {
	if ratingName == "schmeat" {
		return "Schmeat Size", fmt.Sprintf("%s's Thang Thangin' \n%s", user, helper.SchmeatString(value))
	}
	r, ok := standardRatings[ratingName]
	if !ok {
		return "", ""
	}
	return r.Title, fmt.Sprintf("%s's %s Score: %d/100", user, r.ScoreLabel, value)
}

func rateThisSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "rate-this",
		Description: "I give free ratings!",
		Contexts:    helper.GuildOnly,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "I'll rate the pick too",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "dank", Value: "dank"},
					{Name: "epic-gamer", Value: "epic-gamer"},
					{Name: "gay", Value: "gay"},
					{Name: "geek", Value: "geek"},
					{Name: "looks", Value: "looks"},
					{Name: "neckbeard", Value: "neckbeard"},
					{Name: "nerd", Value: "nerd"},
					{Name: "npc", Value: "npc"},
					{Name: "pickme", Value: "pickme"},
					{Name: "schmeat", Value: "schmeat"},
					{Name: "simp", Value: "simp"},
					{Name: "smarts", Value: "smarts"},
					{Name: "stinky", Value: "stinky"},
					{Name: "thot", Value: "thot"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Rated user",
				Required:    false,
			},
		},
	}
}
