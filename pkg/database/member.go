package database

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
)

type DBinventoryItem struct {
	Currency int `json:"currency"`
}

type DBmemberItem struct {
	UserName  string   `json:"userName"`
	UserID    string   `json:"userID"`
	Roles     []string `json:"roles"`
	Inventory DBinventoryItem
}

func InsertDBmemberData(dbClient *dynamodb.DynamoDB, m *discordgo.Member, cfg *config.Configs) error {
	item, err := getGuildItem(dbClient, cfg, m.GuildID)
	if err != nil {
		return err
	}

	//guild exists
	if item.GuildID != "" {
		memberItemList := make(map[string]bool)
		for _, v := range item.Members {
			memberItemList[v.UserID] = true
		}

		if !memberItemList[m.User.ID] {
			member := DBmemberItem{
				UserName: m.User.Username,
				UserID:   m.User.ID,
				Roles:    m.Roles,
				Inventory: DBinventoryItem{
					Currency: 69420,
				},
			}

			marshalMember, err := dynamodbattribute.MarshalMap(member)
			if err != nil {
				return err
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
						S: aws.String(m.GuildID),
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
				return err
			}

			fmt.Printf("Inserted DB Member Item in guild %s\n", m.GuildID)
		}
	}

	return nil
}

func DeleteDBmemberData(dbClient *dynamodb.DynamoDB, m *discordgo.Member, cfg *config.Configs) error {
	//check for existing db item
	item, err := getGuildItem(dbClient, cfg, m.GuildID)
	if err != nil {
		return err
	}

	if item.GuildID != "" {
		for i, v := range item.Members {
			if v.UserID == m.User.ID {
				//remove member
				input := &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"guildID": {
							S: aws.String(m.GuildID),
						},
					},
					TableName:        aws.String(cfg.Configs.Database.TableName),
					ReturnValues:     aws.String("UPDATED_NEW"),
					UpdateExpression: aws.String(fmt.Sprintf("REMOVE Members[%v]", i)),
				}

				_, err = dbClient.UpdateItem(input)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}
