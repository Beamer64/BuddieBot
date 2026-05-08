package slash

import (
	"fmt"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendTxtResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	options := i.ApplicationCommandData().Options[0]

	var err error
	content := ""

	switch options.Name {
	case "clapback":
		text := options.Options[0].StringValue()
		content = strings.ReplaceAll(text, " ", " 👏 ") + " 👏"

	case "bubble", "1337", "cursive", "flipped", "cursed":
		text := strings.ToLower(options.Options[0].StringValue())
		content, err = helper.ToConvertedText(text, options.Name)
		if err != nil {
			return helper.ReturnUserError(s, i, "Unable to convert text atm, try again later.", err)
		}

	case "emojiletters":
		text := strings.ToLower(options.Options[0].StringValue())
		words := strings.Split(text, " ")

		for _, v := range words {
			replacer := strings.NewReplacer(
				"a", "🅰️ ", "b", "🅱️ ", "c", "🇨 ", "d", "🇩 ", "e", "🇪 ", "f", "🇫 ", "g", "🇬 ", "h", "🇭 ", "i", "ℹ️ ", "j", "🇯 ", "k", "🇰 ", "l", "🇱 ", "m", "〽️",
				"n", "🇳 ", "o", "⭕ ", "p", "🅿️ ", "q", "🇶 ", "r", "🇷 ", "s", "🇸 ", "t", "🇹 ", "u", "🇺 ", "v", "🇻 ", "w", "🇼 ", "x", "❎ ", "y", "🇾 ", "z", "🇿 ",
				"0", " ️0️⃣ ", "1", "1️⃣ ", "2", "2️⃣ ", "3", "3️⃣ ", "4", "4️⃣ ", "5", "5️⃣ ", "6", "6️⃣ ", "7", "7️⃣ ", "8", "8️⃣ ", "9", "9️⃣ ",
				"?", "❓ ", "!", "❗ ", "#", "#️⃣ ", "*", "✳️ ", "$", "💲 ", "<", "⏪ ", ">", "⏩ ", "-", "➖ ", "--", "➖ ", "+", "➕ ",
			)
			v = replacer.Replace(v)
			content = fmt.Sprintf("%s%s   ", content, v)
		}

	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func txtSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "txt",
		Description: "Funky Texts",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "1337",
				Description: "1337C0D3",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "73X7 70 CH4N63",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bubble",
				Description: "Bubble Text",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "clapback",
				Description: "Say it with sass",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "cursed",
				Description: "Cursed Text",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "cursive",
				Description: "Say it with class",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "emojiletters",
				Description: "Emoji Text",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "flipped",
				Description: "bǝqqilᖷ",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "Text to change",
						Required:    true,
					},
				},
			},
		},
	}
}
