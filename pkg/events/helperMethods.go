package events

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

// IsLaunchedByDebugger Determines if application is being run by the debugger.
func IsLaunchedByDebugger() bool {
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return true
	}
	return false
}

func getRandomLoadingMessage(possibleMessages []string) string {
	rand.Seed(time.Now().Unix())
	return possibleMessages[rand.Intn(len(possibleMessages))]
}

// GetGuildMembers Discordgo and the discord api are broken atm so niether will get member list
func GetGuildMembers(guildID string, cfg *config.Configs) ([]*discordgo.Member, error) {
	token := ""
	if IsLaunchedByDebugger() {
		token = cfg.Configs.Keys.TestBotToken
	} else {
		token = cfg.Configs.Keys.ProdBotToken
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://discord.com/api/guilds/%s/members", guildID), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req.Header.Add("Authorization", "Bot "+token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var memberObj []*discordgo.Member

	err = json.NewDecoder(res.Body).Decode(&memberObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return memberObj, nil
}

func (d *MessageCreateHandler) memberHasRole(session *discordgo.Session, message *discordgo.MessageCreate, roleName string) bool {
	guildID := message.GuildID
	roleName = strings.ToLower(roleName)

	for _, roleID := range message.Member.Roles {
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

func (g *GuildHandler) insertDBguildData(e *discordgo.Guild) error {
	//check for existing db item
	item, err := g.dbClient.GetItem(
		&dynamodb.GetItemInput{
			TableName: aws.String(g.cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(e.ID),
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if item.Item == nil {
		//create db item
		var memberList []DBmemberItem
		/*for _, v := range e.Members {
			member := DBmemberItem{
				UserName: v.User.Username,
				UserID:   v.User.ID,
				Roles:    v.Roles,
				Inventory: DBinventoryItem{
					Currency: 69420,
				},
			}
			memberList = append(memberList, member)
		}*/

		newItem := DBguildItem{
			GuildID: e.ID,
			Name:    e.Name,
			OwnerID: e.OwnerID,
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
			TableName: aws.String(g.cfg.Configs.Database.TableName),
		}

		_, err = g.dbClient.PutItem(input)
		if err != nil {
			return err
		}

		fmt.Printf("Inserted DB Guild Item for %s\n", e.ID)
	}
	return nil
}

func (g *GuildHandler) deleteDBguildData(e *discordgo.Guild) error {
	//check for existing db item
	item, err := g.dbClient.GetItem(
		&dynamodb.GetItemInput{
			TableName: aws.String(g.cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(e.ID),
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if item.Item != nil {
		//create db item
		input := &dynamodb.DeleteItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(e.ID),
				},
			},
			TableName: aws.String(g.cfg.Configs.Database.TableName),
		}

		_, err = g.dbClient.DeleteItem(input)
		if err != nil {
			return err
		}

		fmt.Printf("Removed DB Guild Item for %s\n", e.ID)
	}
	return nil
}

// todo add in later
func (g *GuildHandler) deleteDBmemberData(e *discordgo.Member) error {
	//check for existing db item
	item, err := g.dbClient.GetItem(
		&dynamodb.GetItemInput{
			TableName: aws.String(g.cfg.Configs.Database.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(e.GuildID),
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if item.Item != nil {
		guildItem := DBguildItem{}
		err = dynamodbattribute.UnmarshalMap(item.Item, &guildItem)
		if err != nil {
			return err
		}

		for i, v := range guildItem.Members {
			if v.UserID == e.User.ID {
				//remove member
				guildItem.Members = append(guildItem.Members[:i], guildItem.Members[i+1:]...)
				break
			}
		}

		type members struct {
			Members []DBmemberItem
		}

		updateData, err := dynamodbattribute.MarshalMap(
			members{
				Members: guildItem.Members,
			},
		)

		input := &dynamodb.UpdateItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"guildID": {
					S: aws.String(e.GuildID),
				},
			},
			TableName: aws.String(g.cfg.Configs.Database.TableName),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":m": {
					M: updateData,
				},
			},
			ReturnValues:     aws.String("UPDATED_NEW"),
			UpdateExpression: aws.String("set Members = :m"),
		}

		_, err = g.dbClient.UpdateItem(input)
		if err != nil {
			return err
		}

	}
	return nil
}
