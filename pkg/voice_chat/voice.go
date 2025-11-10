package voice_chat

import (
	"github.com/Beamer64/BuddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
)

func ConnectVoiceChannel(s *discordgo.Session, userID string, guildID string) (*discordgo.VoiceConnection, error) {
	vc, err := getExistingVoiceConnection(s, guildID)
	if err != nil {
		return nil, err
	}

	if vc == nil {
		voiceState, err := s.State.VoiceState(guildID, userID)
		if err != nil {
			return nil, err
		}

		vc, err = s.ChannelVoiceJoin(guildID, voiceState.ChannelID, false, true)
		if err != nil {
			return nil, err
		}

		err = vc.Speaking(true)
		if err != nil {
			return nil, err
		}

		if !vc.Ready {
			vc.Ready = true
		}

		prepareWebPlayback()

		vc = updateVoiceConnection(s, guildID, vc)
	}

	return vc, nil
}

func getExistingVoiceConnection(s *discordgo.Session, guildID string) (*discordgo.VoiceConnection, error) {
	if vc, ok := s.VoiceConnections[guildID]; ok {
		return vc, nil
	}
	return nil, nil
}

func updateVoiceConnection(s *discordgo.Session, guildID string, vc *discordgo.VoiceConnection) *discordgo.VoiceConnection {
	if _, ok := s.VoiceConnections[guildID]; !ok {
		s.VoiceConnections[guildID] = vc
	}
	return vc
}

func prepareWebPlayback() {
	select {
	case <-web.StopPlaying:
		web.StopPlaying = make(chan bool)
	default:
		if !web.IsPlaying {
			web.StopPlaying = make(chan bool)
		}
	}
}
