package events

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestCommandUsageKey(t *testing.T) {
	opt := func(name string, t discordgo.ApplicationCommandOptionType, val any, sub ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
		return &discordgo.ApplicationCommandInteractionDataOption{
			Name:    name,
			Type:    t,
			Value:   val,
			Options: sub,
		}
	}

	cases := []struct {
		name string
		data discordgo.ApplicationCommandInteractionData
		want string
	}{
		{
			name: "no options",
			data: discordgo.ApplicationCommandInteractionData{Name: "tuuck"},
			want: "tuuck",
		},
		{
			name: "subcommand",
			data: discordgo.ApplicationCommandInteractionData{
				Name:    "audio",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{opt("play", discordgo.ApplicationCommandOptionSubCommand, nil)},
			},
			want: "audio play",
		},
		{
			name: "subcommand group then subcommand",
			data: discordgo.ApplicationCommandInteractionData{
				Name: "image",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					opt("filter", discordgo.ApplicationCommandOptionSubCommandGroup, nil,
						opt("blur", discordgo.ApplicationCommandOptionSubCommand, nil)),
				},
			},
			want: "image filter blur",
		},
		{
			name: "type choice dispatcher",
			data: discordgo.ApplicationCommandInteractionData{
				Name:    "daily",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{opt("type", discordgo.ApplicationCommandOptionString, "horoscope")},
			},
			want: "daily horoscope",
		},
		{
			name: "non-type string option is ignored",
			data: discordgo.ApplicationCommandInteractionData{
				Name:    "tuuck",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{opt("command", discordgo.ApplicationCommandOptionString, "audio")},
			},
			want: "tuuck",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := commandUsageKey(tc.data); got != tc.want {
				t.Errorf("commandUsageKey() = %q, want %q", got, tc.want)
			}
		})
	}
}
