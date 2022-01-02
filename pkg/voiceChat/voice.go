package voiceChat

import (
	"github.com/bwmarrin/discordgo"
)

func ConnectVoiceChannel(s *discordgo.Session, guildID, channelID string) error {
	// Connect to voice channel.
	// NOTE: Setting mute to false, deaf to true.
	dgv, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	return err
}
