package commands

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/database"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"math/rand"
	"reflect"
	"strings"
)

type steamGames struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

type affirmation struct {
	Affirmation string `json:"affirmation"`
}

type kanye struct {
	Quote string `json:"quote"`
}

type advice struct {
	Slip struct {
		ID     int    `json:"id"`
		Advice string `json:"advice"`
	} `json:"slip"`
}

type doggo []struct {
	Breeds []struct {
		Weight struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"weight"`
		Height struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"height"`
		ID               int    `json:"id"`
		Name             string `json:"name"`
		BredFor          string `json:"bred_for"`
		BreedGroup       string `json:"breed_group"`
		LifeSpan         string `json:"life_span"`
		Temperament      string `json:"temperament"`
		Origin           string `json:"origin"`
		ReferenceImageID string `json:"reference_image_id"`
	} `json:"breeds"`
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type joke struct {
	ID   string `json:"id"`
	Joke string `json:"joke"`
}

type pickupLine struct {
	Category string `json:"category"`
	Joke     string `json:"joke"`
}

type wtp struct {
	Data struct {
		Type      []string `json:"Type"`
		Abilities []string `json:"abilities"`
		ASCII     string   `json:"ascii"`
		Height    float64  `json:"height"`
		ID        int      `json:"id"`
		Link      string   `json:"link"`
		Name      string   `json:"name"`
		Weight    int      `json:"weight"`
	} `json:"Data"`
	Answer   string `json:"answer"`
	Question string `json:"question"`
}

// Returns pseudo rand num between low and high.
// For random embed color: rangeIn(1, 16777215)
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// Checks if the value is empty and returns it if not.
// Otherwise, return 'N/A'
func checkIfEmpty(value string) string {
	if value != "" {
		return value
	}
	return "N/A"
}

func memberHasRole(session *discordgo.Session, i *discordgo.InteractionCreate, roleName string) bool {
	guildID := i.GuildID
	roleName = strings.ToLower(roleName)

	for _, roleID := range i.Member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}

// this is ugly and I have no pride
func getConfigSettingValue(settingName string, guildID string, cfg *config.Configs) (string, error) {
	dynamodbSess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(cfg.Configs.Database.Region),
			Credentials: credentials.NewStaticCredentials(cfg.Configs.Database.AccessKey, cfg.Configs.Database.SecretKey, ""),
		},
	)
	if err != nil {
		return "", err
	}

	dbClient := dynamodb.New(dynamodbSess)

	item, err := database.GetDBguildItemByID(dbClient, cfg, guildID)
	if err != nil {
		return "", err
	}

	settingsList := make(map[string]interface{})
	v := reflect.ValueOf(item.GuildSettings)
	for i := 0; i < v.NumField(); i++ {
		settingsList[v.Type().Field(i).Name] = v.Field(i).Interface()
	}

	value := ""
	switch settingName {
	case "CommandPrefix":
		value = reflect.ValueOf(settingsList["CommandPrefix"]).String()
		return fmt.Sprintf("Command Prefix - `CommandPrefix` - %s", value), nil

	case "ModerateProfanity":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["ModerateProfanity"]).Bool())
		if value == "false" {
			value = "Disabled ❌)"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Moderate Profanity - `ModerateProfanity` - %s", value), nil

	case "DisableNSFW":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["DisableNSFW"]).Bool())
		if value == "false" {
			value = "Disabled ❌)"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Disable NSFW - `DisableNSFW` - %s", value), nil

	case "ModerateSpam":
		value = fmt.Sprintf("%v", reflect.ValueOf(settingsList["ModerateSpam"]).Bool())
		if value == "false" {
			value = "Disabled ❌)"
		} else {
			value = "Enabled ✔️"
		}
		return fmt.Sprintf("Moderate Spam - `ModerateSpam` - %s", value), nil
	}

	return "N/A", nil
}
