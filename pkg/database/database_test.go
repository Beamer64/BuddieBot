package database

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/beamer64/discordBot/pkg/config"
	"os"
	"testing"
)

func TestGetDBguildItem(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
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

	var guildObj DBguildItem
	item, err := dbClient.GetItem(
		&dynamodb.GetItemInput{
			TableName: aws.String(cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(cfg.Configs.DiscordIDs.TestGuildID),
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = dynamodbattribute.UnmarshalMap(item.Item, &guildObj)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(guildObj)
}

func TestInsertDBmemberData(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
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

	memberItemList := make(map[string]bool)
	memberItemList["866151939472883762"] = true
	memberItemList["123456789"] = true

	if !memberItemList["866151939472883762"] {
		var roles = []string{
			"11111111111",
		}
		member := DBmemberItem{
			UserName: "BuddieBot",
			UserID:   "866151939472883762",
			Roles:    roles,
			Inventory: DBinventoryItem{
				Currency: 69420,
			},
		}

		//marshal member struct
		marshalMember, err := dynamodbattribute.MarshalMap(member)
		if err != nil {
			t.Fatal(err)
		}

		var members []*dynamodb.AttributeValue

		//create attribute value
		av := &dynamodb.AttributeValue{
			M: marshalMember,
		}
		members = append(members, av)

		input := &dynamodb.UpdateItemInput{
			TableName: aws.String(cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(cfg.Configs.DiscordIDs.TestGuildID),
				},
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":r": {
					L: members,
				},
				":empty_list": {
					L: []*dynamodb.AttributeValue{},
				},
			},
			ReturnValues:     aws.String("UPDATED_NEW"),
			UpdateExpression: aws.String("SET Members = list_append(if_not_exists(Members, :empty_list), :r)"),
		}

		_, err = dbClient.UpdateItem(input)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		fmt.Println("Already in list")
	}
}

func TestDeleteDBmemberData(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
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

	item, err := getGuildItem(dbClient, cfg, cfg.Configs.DiscordIDs.TestGuildID)
	if err != nil {
		t.Fatal(err)
	}

	if item.GuildID != "" {
		for i, v := range item.Members {
			if v.UserID == "866151939472883762" {
				input := &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"guildID": {
							S: aws.String(cfg.Configs.DiscordIDs.TestGuildID),
						},
					},
					TableName:        aws.String(cfg.Configs.Database.TableName),
					ReturnValues:     aws.String("UPDATED_NEW"),
					UpdateExpression: aws.String(fmt.Sprintf("REMOVE Members[%v]", i)),
				}

				_, err = dbClient.UpdateItem(input)
				if err != nil {
					t.Fatal(err)
				}

				fmt.Println("Removed DB Member item")
				break
			}
		}
	} else {
		fmt.Println("No Guild Item found")
	}
}
