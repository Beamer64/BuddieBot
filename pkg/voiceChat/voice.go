package voiceChat

import (
	"github.com/beamer64/discordBot/pkg/webscrape"
	"github.com/bwmarrin/discordgo"
)

type VoiceConnection struct {
	Dgv *discordgo.VoiceConnection
}

func ConnectVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate, guildID string, errChannelID string) (*discordgo.VoiceConnection, error) {
	vc := VoiceConnection{}

	if vc.Dgv == nil {
		voiceState, err := s.State.VoiceState(guildID, m.Author.ID)
		if err != nil {
			return nil, err
		}

		vc.Dgv, err = s.ChannelVoiceJoin(guildID, voiceState.ChannelID, false, true)
		if err != nil {
			if _, ok := s.VoiceConnections[guildID]; ok {
				vc.Dgv = s.VoiceConnections[guildID]
			} else {
				return nil, err
			}
		}

		_, err = s.ChannelMessageSend(errChannelID, "Voice Channel Joined")
		if err != nil {
			return nil, err
		}

		err = vc.Dgv.Speaking(true)
		if err != nil {
			return nil, err
		}

		if !vc.Dgv.Ready {
			vc.Dgv.Ready = true
		}

		if webScrape.StopPlaying == nil {
			webScrape.StopPlaying = make(chan bool)
		}
	}

	return vc.Dgv, nil
}
