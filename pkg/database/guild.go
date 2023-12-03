package database

import (
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"reflect"
)

// GetConfigSettingValueByName this is ugly and I have no pride WIP
func GetConfigSettingValueByName(settingName string, guildID string, cfg *config.Configs) (string, error) {
	settingsList := make(map[string]interface{})

	value := ""
	switch settingName {
	case "CommandPrefix":
		value = reflect.ValueOf(settingsList["CommandPrefix"]).String()
		return fmt.Sprintf("Command Prefix - `CommandPrefix` - %s", value), nil

	case "ModerateProfanity":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["ModerateProfanity"]).Bool())
		if value == "false" {
			value = "Disabled ❌"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Moderate Profanity - `ModerateProfanity` - %s", value), nil

	case "DisableNSFW":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["DisableNSFW"]).Bool())
		if value == "false" {
			value = "Disabled ❌"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Disable NSFW - `DisableNSFW` - %s", value), nil

	case "ModerateSpam":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["ModerateSpam"]).Bool())
		if value == "false" {
			value = "Disabled ❌"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Moderate Spam - `ModerateSpam` - %s", value), nil
	}

	return "N/A", nil
}
