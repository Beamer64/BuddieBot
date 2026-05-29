package prefix

import (
	"fmt"

	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// buildReleaseNotesEmbed returns a fresh release-notes embed per call. Command mentions
// resolve via helper.CommandMention against the IDs captured at registration; they fall
// back to plain "/cmd" text if the registry isn't populated.
func buildReleaseNotesEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "Release Notes — May 2026",
		URL:   "https://github.com/Beamer64/BuddieBot/blob/master/res/release.md",
		Description: "BuddieBot's been hitting the gym. 🏋️ Here's what's new — under-the-hood " +
			"glow-ups, a smarter help command, and the groundwork for some big stuff coming soon.\n\n" +
			"Full notes live in the title link above.",
		Color:  11091696,
		Author: &discordgo.MessageEmbedAuthor{},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "📈 We're back, baby. This time for good...probably.",
				Value: "The bot is now being self hosted to ensure we have as much uptime as possible " +
					"while paying as little as possible. (I'm doing this for free and I'm cheap.) " +
					"Only power and internet outages can stop us now.",
			},
			{
				Name: "🧱 Built to last",
				Value: "The image and data commands now run on BuddieBot's own in-house libraries instead of " +
					"leaning on third-party services. Fewer outages, snappier effects, and 60+ image filters, " +
					"distorts, and memes that keep working even when someone else's API falls over.",
			},
			{
				Name: "📖 Help that actually helps",
				Value: fmt.Sprintf(
					"Run %s to flip through every command, or add `command:<name>` to dig into one and see its "+
						"options plus a real usage example. It reads from the live command list, so it never goes stale.",
					helper.CommandMention("tuuck", "cmd-list"),
				),
			},
			{
				Name: "💾 BuddieBot has a brain now",
				Value: "There's a real datastore behind the scenes powering per-server settings — and laying the " +
					"foundation for the big one: a **coming-soon economy** so you can earn, hoard, and flex " +
					"the imaginary currency, Dosh! 💰 (Still waiting on focus group data for the name..)",
			},
			{
				Name: "🎵 Audio, all in one place",
				Value: fmt.Sprintf(
					"Music now lives under %s — play, queue, skip, clear, stop, and resume-queue. Paste a YouTube "+
						"playlist URL and it'll queue the whole thing.\n*Heads up: audio is enabled in select servers "+
						"only — if it's quiet here, that's why.*",
					helper.CommandMention("audio", "play"),
				),
			},
			{
				Name: "🙋 Your data, your call",
				Value: fmt.Sprintf(
					"%s shows what BuddieBot has on you; %s wipes everything across every server, no questions "+
						"asked. Tracking is opt-in and your messages are never logged.",
					helper.CommandMention("user", "profile"),
					helper.CommandMention("user", "forget-me"),
				),
			},
			{
				Name: "⚙️ Make it yours + steadier",
				Value: fmt.Sprintf(
					"Admins can set a custom command prefix with %s. More server customization coming and the heavier commands are now rate-limited "+
						"so one button-masher can't bog things down for everyone.",
					helper.CommandMention("admin", "set-prefix"),
				),
			},
			{
				Name: "👁️ Big Brother is always watching",
				Value: "Enhancements to all logging now ensures any/all errors will come across my desk. " +
					"And if any do not, then I'll panic and cry.",
			},
		},
	}
}
