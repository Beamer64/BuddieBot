package database

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
)

type DBguildSettingsItem struct {
	ModerateSpam      bool `json:"moderateSpam"`
	ModerateProfanity bool `json:"moderateProfanity"`
	DisableNSFW       bool `json:"disableNSFW"`
}

type DBguildItem struct {
	GuildID       string `json:"guildID"`
	Name          string `json:"name"`
	OwnerID       string `json:"ownerID"`
	GuildSettings DBguildSettingsItem
	Members       []DBmemberItem
}

func getGuildItem(dbClient *dynamodb.DynamoDB, cfg *config.Configs, guildID string) (DBguildItem, error) {
	var guildObj DBguildItem

	item, err := dbClient.GetItem(
		&dynamodb.GetItemInput{
			TableName: aws.String(cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(guildID),
				},
			},
		},
	)
	if err != nil {
		return guildObj, err
	}

	err = dynamodbattribute.UnmarshalMap(item.Item, &guildObj)
	if err != nil {
		return guildObj, err
	}

	return guildObj, nil
}

func InsertDBguildItem(dbClient *dynamodb.DynamoDB, g *discordgo.Guild, cfg *config.Configs) error {
	//check for existing db item
	item, err := getGuildItem(dbClient, cfg, g.ID)
	if err != nil {
		return err
	}

	// guild does not exist in DB
	if item.GuildID == "" {
		//create db item
		var memberList []DBmemberItem
		for _, v := range g.Members {
			member := DBmemberItem{
				UserName: v.User.Username,
				UserID:   v.User.ID,
				Roles:    v.Roles,
				Inventory: DBinventoryItem{
					Currency: 69420,
				},
			}
			memberList = append(memberList, member)
		}

		newItem := DBguildItem{
			GuildID: g.ID,
			Name:    g.Name,
			OwnerID: g.OwnerID,
			GuildSettings: DBguildSettingsItem{
				ModerateSpam:      false,
				ModerateProfanity: false,
				DisableNSFW:       false,
			},
			Members: memberList,
		}

		av, err := dynamodbattribute.MarshalMap(newItem)
		if err != nil {
			return err
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(cfg.Configs.Database.TableName),
		}

		_, err = dbClient.PutItem(input)
		if err != nil {
			return err
		}

		fmt.Printf("Inserted DB Guild Item for %s\n", g.ID)
	}
	return nil
}

func DeleteDBguildData(dbClient *dynamodb.DynamoDB, g *discordgo.Guild, cfg *config.Configs) error {
	//check for existing db item
	item, err := getGuildItem(dbClient, cfg, g.ID)
	if err != nil {
		return err
	}

	// guild exists
	if item.GuildID != "" {
		input := &dynamodb.DeleteItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(g.ID),
				},
			},
			TableName: aws.String(cfg.Configs.Database.TableName),
		}

		_, err := dbClient.DeleteItem(input)
		if err != nil {
			return err
		}

		fmt.Printf("Removed DB Guild Item for %s\n", g.ID)
	}
	return nil
}
