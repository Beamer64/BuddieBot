package slash

import (
	"fmt"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func sendTxtResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
			go func() {
				err = helper.SendResponseErrorToUser(s, i, "Unable to convert text atm, try again later.")
			}()
			return err
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
