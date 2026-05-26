package slash

import (
	"fmt"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/bb_data/textfonts"
	"github.com/bwmarrin/discordgo"
)

func sendTxtResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	optMap := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, opt := range i.ApplicationCommandData().Options {
		optMap[opt.Name] = opt
	}
	cmdType := optMap["type"].StringValue()

	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get type %s: %w", cmdType, err)
	}

	content := ""

	switch cmdType {
	case "clapback":
		text := optMap["text"].StringValue()
		content = strings.ReplaceAll(text, " ", " 👏 ") + " 👏"

	case "bubble", "1337", "cursive", "flipped", "cursed":
		text := strings.ToLower(optMap["text"].StringValue())
		content = textfonts.Convert(text, cmdType)

	case "emoji-letters":
		text := strings.ToLower(optMap["text"].StringValue())
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

	default:
		return fmt.Errorf("unknown option: %s", cmdType)

	}

	webhookEdit := &discordgo.WebhookEdit{
		Content: &content,
	}

	if _, err := s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		return fmt.Errorf("send /get response for type %s: %w", cmdType, err)
	}

	return nil
}

func txtSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "txt",
		Description: "Funky Texts",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Text responsibly",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "1337", Value: "1337"},
					{Name: "bubble", Value: "bubble"},
					{Name: "clapback", Value: "clapback"},
					{Name: "cursed", Value: "cursed"},
					{Name: "cursive", Value: "cursive"},
					{Name: "emoji-letters", Value: "emoji-letters"},
					{Name: "flipped", Value: "flipped"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text",
				Required:    true,
			},
		},
	}
}
