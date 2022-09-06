package games

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

/*func SendNimEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	embed := &discordgo.MessageEmbed{
		Title:       "12 Coin Nim",
		Color:       helper.RangeIn(1, 16777215),
		Description: doggoObj[0].Breeds[0].Temperament,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Weight",
				Value:  fmt.Sprintf("%s lbs / %s kg", impWeight, metWeight),
				Inline: true,
			},
			{
				Name:   "Height",
				Value:  fmt.Sprintf("%s in / %s cm", impHeight, metHeight),
				Inline: true,
			},
		},
	}
	err := startNim(s, i)
	if err != nil {
		return err
	}

	return nil
}*/

func startNim(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	return nil
}

func sendPlayerInvite(s *discordgo.Session, m *discordgo.MessageCreate, user string) (bool, error) {
	usrID := strings.SplitAfter(user, "!")
	userID := strings.Split(usrID[1], ">")[0]

	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return false, err
	}

	_, err = s.ChannelMessageSend(channel.ID, m.Author.Username+" has requested to play Nim with you. Would you like to accept?")
	if err != nil {
		return false, err
	}

	return false, nil
}
