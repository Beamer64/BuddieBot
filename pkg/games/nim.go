package games

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func StartNim(s *discordgo.Session, m *discordgo.MessageCreate, user string, versus bool) error {
	if versus {
		accepted, err := SendPlayerInvite(s, m, user)
		if err != nil {
			return err
		}

		if accepted {

		} else {

		}

	} else {

	}
	return nil
}

func SendPlayerInvite(s *discordgo.Session, m *discordgo.MessageCreate, user string) (bool, error) {
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
