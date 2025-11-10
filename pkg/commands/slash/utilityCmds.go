package slash

import (
	"fmt"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"reflect"
	"strings"
)

func sendTuuckResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return sendTuuckCommands(s, i, cfg)
	}

	cmdName := options[0].StringValue()
	if strings.HasPrefix(cmdName, "/") {
		cmdName = cmdName[1:]
	}

	cmdInfo := getCommandInfo(cmdName, cfg)
	if cmdInfo == nil {
		return helper.SendResponseErrorToUser(s, i, fmt.Sprintf("Invalid command: %s", cmdName))
	}

	embed := &discordgo.MessageEmbed{
		Title: cmdInfo.Name + " info",
		Color: helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Description",
				Value:  cmdInfo.Desc,
				Inline: false,
			},
			{
				Name:   "Usage",
				Value:  "`" + cmdInfo.Name + "`",
				Inline: false,
			},
			{
				Name:   "Example",
				Value:  cmdInfo.Example,
				Inline: false,
			},
		},
	}

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		},
	)

	return err
}

func sendTuuckCommands(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	var content strings.Builder
	content.WriteString("A list of current Slash command groups\n```\n")

	v := reflect.ValueOf(&cfg.Cmd.SlashName).Elem()

	for n := 0; n < v.NumField(); n++ {
		field := v.Type().Field(n)
		_, err := fmt.Fprintf(&content, "%s\n", field.Name)
		if err != nil {
			return fmt.Errorf("error formatting string: %v", err)
		}
	}

	content.WriteString("```\nYou can get more information about a command by using `/tuuck <command_name>`")

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content.String(),
			},
		},
	)

	return err
}

func getCommandInfo(cmdName string, cfg *config.Configs) *tuuckCmdInfo {
	var info tuuckCmdInfo

	n := reflect.ValueOf(&cfg.Cmd.SlashName).Elem()
	d := reflect.ValueOf(&cfg.Cmd.Desc).Elem()
	e := reflect.ValueOf(&cfg.Cmd.Example).Elem()

	for i := 0; i < n.NumField(); i++ {
		field := n.Type().Field(i)
		if strings.EqualFold(field.Name, cmdName) {
			info.Name = fmt.Sprintf("%s", n.Field(i).Interface())
			break
		}
	}

	for i := 0; i < d.NumField(); i++ {
		field := d.Type().Field(i)
		if strings.EqualFold(field.Name, cmdName) {
			info.Desc = fmt.Sprintf("%s", d.Field(i).Interface())
			break
		}
	}

	for i := 0; i < e.NumField(); i++ {
		field := e.Type().Field(i)
		if strings.EqualFold(field.Name, cmdName) {
			info.Example = fmt.Sprintf("%s", e.Field(i).Interface())
			break
		}
	}

	if info.Name != "" {
		return &info
	} else {
		return nil
	}
}

func sendConfigResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	settingListEmbed, err := getSettingsListEmbed(i.GuildID, cfg)
	if err != nil {
		return err
	}

	if !helper.MemberHasRole(s, i.Member, i.GuildID, cfg.Configs.Settings.BotAdminRole) {
		// send setting list
		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						settingListEmbed,
					},
				},
			},
		)
		if err != nil {
			return err
		}
	} else {
		switch i.ApplicationCommandData().Options[0].Name {
		case "list":
			// send setting list
			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							settingListEmbed,
						},
					},
				},
			)
			if err != nil {
				return err
			}

		// TODO fix this
		case "setting":
			/*settingName := i.ApplicationCommandData().Options[0].Options[0].StringValue()
			newSettingValue := i.ApplicationCommandData().Options[0].Options[1].StringValue()*/

		default:
			// send setting list
			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							settingListEmbed,
						},
					},
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getSettingsListEmbed(guildID string, cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	cmdPrefix, err := database.GetConfigSettingValueByName("CommandPrefix", guildID, cfg)
	if err != nil {
		return nil, err
	}
	modProfanity, err := database.GetConfigSettingValueByName("ModerateProfanity", guildID, cfg)
	if err != nil {
		return nil, err
	}
	disableNSFW, err := database.GetConfigSettingValueByName("DisableNSFW", guildID, cfg)
	if err != nil {
		return nil, err
	}
	modSpam, err := database.GetConfigSettingValueByName("ModerateSpam", guildID, cfg)
	if err != nil {
		return nil, err
	}

	settingListEmbed := &discordgo.MessageEmbed{
		Title:       "BuddieBot Server Settings",
		Description: "These are the current settings for the server and can only be changed by holders of the **Bot Admin Role**.",
		Color:       helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   cmdPrefix,
				Value:  "This is the prefix to any non-slash commands. The default prefix is '$'.",
				Inline: false,
			},
			{
				Name:   modProfanity,
				Value:  "When enabled, BuddieBot will do it's best to filter out any profane chats.",
				Inline: false,
			},
			{
				Name:   disableNSFW,
				Value:  "When disabled, the server will not have access to any NSFW commands/content.",
				Inline: false,
			},
			{
				Name:   modSpam,
				Value:  "When enabled, BuddieBot will remove any unwanted chat spamming.",
				Inline: false,
			},
		},
	}
	return settingListEmbed, nil
}
