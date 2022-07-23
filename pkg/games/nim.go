package games

import (
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func PlayNIM(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, user string) error {
	/*if strings.HasPrefix(param, "<@") {
		err := startNim(s, m, param, true)
		if err != nil {
			return err
		}

	} else {
		if param == "" {
			err := startNim(s, m, param, false)
			if err != nil {
				return err
			}

		} else {
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				return err
			}
		}
	}*/
	return nil
}

func startNim(s *discordgo.Session, m *discordgo.MessageCreate, user string, versus bool) error {
	if versus {
		accepted, err := sendPlayerInvite(s, m, user)
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
