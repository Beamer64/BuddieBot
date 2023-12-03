package database

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"os"
	"strings"
	"testing"
)

func TestChangeConfigSettingValueByName(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	dynamodbSess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(cfg.Configs.Database.Region),
			Credentials: credentials.NewStaticCredentials(cfg.Configs.Database.AccessKey, cfg.Configs.Database.SecretKey, ""),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	dbClient := dynamodb.New(dynamodbSess)

	settingName := "CommandPrefix"
	settingValue := "%"

	newVal := ""
	expr := map[string]*dynamodb.AttributeValue{}
	switch settingName {
	case "CommandPrefix":
		newVal = strings.Trim(settingValue, " ")
		expr = map[string]*dynamodb.AttributeValue{
			":v": {
				S: aws.String(newVal),
			},
		}
	case "ModerateProfanity":
		if helper.StringInSlice(strings.ToLower(settingValue), helper.ApprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else if helper.StringInSlice(strings.ToLower(settingValue), helper.DisapprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else {

		}
	case "DisableNSFW":
		if helper.StringInSlice(strings.ToLower(settingValue), helper.ApprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else if helper.StringInSlice(strings.ToLower(settingValue), helper.DisapprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else {

		}
	case "ModerateSpam":
		if helper.StringInSlice(strings.ToLower(settingValue), helper.ApprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else if helper.StringInSlice(strings.ToLower(settingValue), helper.DisapprovalWords) {
			newVal = strings.Trim(settingValue, " ")
			expr = map[string]*dynamodb.AttributeValue{
				":v": {
					S: aws.String(newVal),
				},
			}
		} else {

		}

	default:

	}

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"guildID": {
				S: aws.String(cfg.Configs.DiscordIDs.TestGuildID),
			},
		},
		ExpressionAttributeValues: expr,
		TableName:                 aws.String(cfg.Configs.Database.TableName),
		ReturnValues:              aws.String("UPDATED_NEW"),
		UpdateExpression:          aws.String(fmt.Sprintf("SET GuildSettings.%s = :v", settingName)),
	}

	_, err = dbClient.UpdateItem(input)
	if err != nil {
		t.Fatal(err)
	}
}
