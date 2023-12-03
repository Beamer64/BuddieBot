package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/api"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/database"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"github.com/mitchellh/mapstructure"
	"io"
	"math/rand"
	"mvdan.cc/xurls/v2"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// functions here should mostly be used for the slash commands

//region Utility Commands

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
		return helper.SendResponseError(s, i, fmt.Sprintf("Invalid command: %s", cmdName))
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
		//send setting list
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
			//send setting list
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
			//send setting list
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

//endregion

//region Play Commands

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var embed *discordgo.MessageEmbed
	var data *discordgo.InteractionResponseData
	var err error

	switch options.Name {
	case "coin-flip":
		embed, err = getCoinFlipEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, "Unable to flip coin atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "just-lost":
		embed = &discordgo.MessageEmbed{
			Title:       "You just lost The Game.",
			Color:       helper.RangeIn(1, 16777215),
			Description: "..Told you not to play.",
		}

		channel, err := s.UserChannelCreate(i.Member.User.ID)
		if err != nil {
			return err
		}

		_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
		if err != nil {
			return err
		}

		embed = &discordgo.MessageEmbed{
			Title: "Check your DM's  👀",
			Color: helper.RangeIn(1, 16777215),
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	// todo finish this
	case "nim":
		/*err := games.SendNimEmbed(s, i, cfg)
		if err != nil {
			return err
		}*/

	// todo finish this
	case "typeracer":

	case "gtl":
		clientData, err := client.GTL()
		if err != nil {
			return err
		}

		embed, err = getGTLembed(clientData)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "wtp":
		clientData, err := client.WTP()
		if err != nil {
			return err
		}

		embed, err = getWTPembed(clientData, false)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "wyr":
		embed, err = getWYREmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Another One! (▀̿Ĺ̯▀̿ ̿)",
							Style:    1,
							CustomID: "wyr-button",
						},
					},
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}

func getGTLembed(data interface{}) (*discordgo.MessageEmbed, error) {
	var gtlObj gtl
	err := mapstructure.Decode(data, &gtlObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Clue",
				Value:  gtlObj.Clue,
				Inline: false,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: gtlObj.Question,
		},
	}

	return embed, nil
}

func getWYREmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.WYRAPI)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var wyrObj wyr
	err = json.NewDecoder(res.Body).Decode(&wyrObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Would You Rather?",
		Color:       helper.RangeIn(1, 16777215),
		Description: wyrObj.Data,
	}

	return embed, nil
}

func getWTPembed(data interface{}, isAnswer bool) (*discordgo.MessageEmbed, error) {
	var wtpObj wtp
	err := mapstructure.Decode(data, &wtpObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{}

	if isAnswer {

	} else {
		embed = &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: wtpObj.Question,
			},
		}
	}

	return embed, nil
}

func getCoinFlipEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	search, results := "", ""
	if randNum%2 == 0 {
		search = "Coin Flip Heads"
		results = "Heads"

	} else {
		search = "Coin Flip Tails"
		results = "Tails"
	}

	gifURL, err := api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Flipping...",
		Description: fmt.Sprintf("It's %s!", results),
		Color:       helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}

	return embed, nil
}

//endregion

//region Animal Commands

func sendAnimalsResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	commandName := i.ApplicationCommandData().Options[0].Name

	var embed *discordgo.MessageEmbed
	var data *discordgo.InteractionResponseData
	var err error
	errRespMsg := "Unable to make call at this moment, please try later :("

	switch commandName {
	case "doggo":
		embed, err = getDoggoEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "katz":
		embed, err = getKatzEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}
	}
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}

func getDoggoEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	callCount := 1

	doggoObj, err := callDoggoAPI(cfg)
	if err != nil {
		return nil, err
	}

	for len(doggoObj) == 0 || len(doggoObj[0].Breeds) == 0 {
		doggoObj, err = callDoggoAPI(cfg)
		if err != nil {
			return nil, err
		}
		callCount++

		if callCount == 5 {
			return nil, fmt.Errorf("error retrieving doggoObj. Attempts made: %v", callCount)
		}
	}

	breed := doggoObj[0].Breeds[0]

	impWeight := helper.CheckIfStructValueISEmpty(breed.Weight.Imperial)
	metWeight := helper.CheckIfStructValueISEmpty(breed.Weight.Metric)
	impHeight := helper.CheckIfStructValueISEmpty(breed.Height.Imperial)
	metHeight := helper.CheckIfStructValueISEmpty(breed.Height.Metric)

	embed := &discordgo.MessageEmbed{
		Title:       breed.Name,
		Color:       helper.RangeIn(1, 16777215),
		Description: breed.Temperament,
		Image: &discordgo.MessageEmbedImage{
			URL: doggoObj[0].URL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Weight",
				Value:  fmt.Sprintf("%slbs / %skg", impWeight, metWeight),
				Inline: true,
			},
			{
				Name:   "Breed Group",
				Value:  helper.CheckIfStructValueISEmpty(breed.BreedGroup),
				Inline: true,
			},
			{
				Name:   "Origin",
				Value:  helper.CheckIfStructValueISEmpty(breed.Origin),
				Inline: true,
			},
			{
				Name:   "Height",
				Value:  fmt.Sprintf("%sin / %scm", impHeight, metHeight),
				Inline: true,
			},
			{
				Name:   "Life Span",
				Value:  helper.CheckIfStructValueISEmpty(breed.LifeSpan),
				Inline: true,
			},
			{
				Name:   "Good Pup",
				Value:  "10/10",
				Inline: true,
			},
			{
				Name:   "Bred For",
				Value:  helper.CheckIfStructValueISEmpty(breed.BredFor),
				Inline: false,
			},
		},
	}

	return embed, nil
}

func getKatzEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	descTiers := map[int]string{
		0: "★★★★★",
		1: "⭐★★★★",
		2: "⭐⭐★★★",
		3: "⭐⭐⭐★★",
		4: "⭐⭐⭐⭐★",
		5: "⭐⭐⭐⭐⭐",
	}

	katzObj, err := callKatzAPI(cfg)
	if err != nil {
		return nil, err
	}

	i := rand.Intn(len(katzObj))
	breed := katzObj[i]

	name := helper.CheckIfStructValueISEmpty(breed.Name)
	origin := helper.CheckIfStructValueISEmpty(breed.Origin)
	length := helper.CheckIfStructValueISEmpty(breed.Length)

	minLife := helper.CheckIfStructValueISEmpty(breed.MinLifeExpectancy)
	maxLife := helper.CheckIfStructValueISEmpty(breed.MaxLifeExpectancy)
	lifeSpan := fmt.Sprintf("%s - %s years", minLife, maxLife)

	minWeight := helper.CheckIfStructValueISEmpty(breed.MinWeight)
	maxWeight := helper.CheckIfStructValueISEmpty(breed.MaxWeight)
	weight := fmt.Sprintf("%s - %s lbs", minWeight, maxWeight)

	embed := &discordgo.MessageEmbed{
		Title:       name,
		Description: "Good Beans",
		Color:       helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: breed.ImageLink,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Origin",
				Value:  origin,
				Inline: true,
			},
			{
				Name:   "Weight",
				Value:  weight,
				Inline: true,
			},
			{
				Name:   "Life Span",
				Value:  lifeSpan,
				Inline: true,
			},
			{
				Name:   "Grooming Req",
				Value:  descTiers[breed.Grooming],
				Inline: true,
			},
			{
				Name:   "Playfulness",
				Value:  descTiers[breed.Playfulness],
				Inline: true,
			},
			{
				Name:   "Affection",
				Value:  descTiers[breed.FamilyFriendly],
				Inline: true,
			},
			{
				Name:   "Pet Friendly",
				Value:  descTiers[breed.OtherPetsFriendly],
				Inline: true,
			},
			{
				Name:   "Children Friendly",
				Value:  descTiers[breed.ChildrenFriendly],
				Inline: true,
			},
			{
				Name:   "Intelligence",
				Value:  descTiers[breed.Intelligence],
				Inline: true,
			},
			{
				Name:   "General Health",
				Value:  descTiers[breed.GeneralHealth],
				Inline: true,
			},
			{
				Name:   "Length",
				Value:  length,
				Inline: true,
			},
		},
	}

	return embed, nil
}

func callDoggoAPI(cfg *config.Configs) ([]doggo, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", cfg.Configs.ApiURLs.DoggoAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Set("x-api-key", cfg.Configs.Keys.DoggoAPIkey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request doggoKatz URL: %v", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	var doggoObj []doggo
	err = json.NewDecoder(resp.Body).Decode(&doggoObj)
	if err != nil {
		return nil, err
	}

	return doggoObj, nil
}

func callKatzAPI(cfg *config.Configs) ([]katz, error) {
	charset := "abcdefghijklmnopqrstuvwxyz"
	c := string(charset[rand.Intn(len(charset))])

	client := http.Client{}
	req, err := createNinjaAPIrequest(cfg, cfg.Configs.ApiURLs.NinjaKatzAPI+c)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request doggoKatz URL: %v", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	var katzObj []katz
	err = json.NewDecoder(resp.Body).Decode(&katzObj)
	if err != nil {
		return nil, err
	}

	return katzObj, nil
}

func createNinjaAPIrequest(cfg *config.Configs, url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Set("x-api-key", cfg.Configs.Keys.NinjaAPIKey)
	return req, nil
}

//endregion

//region Get Commands

func sendGetResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var embed *discordgo.MessageEmbed
	var data *discordgo.InteractionResponseData
	var err error

	errRespMsg := "Unable to fetch data atm, Try again later."

	switch options.Name {
	case "rekd":
		clientData, err := client.Roast()
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		content := ""
		switch len(options.Options) {
		// todo add comments
		case 0:
			content = fmt.Sprintf("<@!%s>\n%s", i.Member.User.ID, clientData)

		case 1:
			user := options.Options[0].UserValue(s)

			content = fmt.Sprintf("<@!%s>\n%s", user.ID, clientData)
		}

		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s\n(ง ͠° ͟ل͜ ͡°)ง", content),
		}

	case "joke":
		clientData, err := client.Joke()
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		var jokeObj joke
		err = mapstructure.Decode(clientData, &jokeObj)
		if err != nil {
			return err
		}

		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s", jokeObj.Joke),
		}

	case "8ball":
		clientData, err := client.Eightball()
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s", clientData),
		}

	case "yomomma":
		clientData, err := client.Yomama()
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		content := ""
		switch len(options.Options) {
		case 0:
			content = fmt.Sprintf("<@!%s>\n%s", i.Member.User.ID, clientData)

		case 1:
			user := options.Options[0].UserValue(s)
			content = fmt.Sprintf("<@!%s>\n%s", user.ID, clientData)
		}

		data = &discordgo.InteractionResponseData{
			Content: content,
		}

	case "pickup-line":
		clientData, err := client.PickupLine()
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		var pickupObj pickupLine
		err = mapstructure.Decode(clientData, &pickupObj)
		if err != nil {
			return err
		}

		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s", pickupObj.Joke),
		}

	case "fake-person":
		personData, err := callFakePersonAPI(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		embed = getFakePersonEmbed(personData)

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "xkcd":
		embed, err = getXkcdEmbed(cfg)

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

		/*case "captcha":
		data, err := client.WTP()
		if err != nil {
			return err
		}*/

	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}

func getXkcdEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	resp, err := http.Get(cfg.Configs.ApiURLs.XkcdAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc := fmt.Sprintf("%s", html)

	rxRelaxed := xurls.Strict()
	links := rxRelaxed.FindAllString(doc, -1)

	img := ""
	for _, l := range links {
		if strings.Contains(l, "https://imgs.xkcd.com/comics/") {
			img = l
			break
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Sunfay Dunnies",
		Color: helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: img,
		},
	}

	return embed, nil
}

func callFakePersonAPI(cfg *config.Configs) (fakePerson, error) {
	var personObj fakePerson

	resp, err := http.Get(cfg.Configs.ApiURLs.FakePersonAPI)
	if err != nil {
		return personObj, err
	}

	if resp.StatusCode != http.StatusOK {
		return personObj, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&personObj)
	if err != nil {
		return personObj, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return personObj, nil
}

func getFakePersonEmbed(fakePersonObj fakePerson) *discordgo.MessageEmbed {
	fpObj := fakePersonObj.Results[0]
	dob := strings.Split(fpObj.Dob.Date, "T")

	embed := &discordgo.MessageEmbed{
		Title:       "Fake Person Generator",
		Description: "BuddieBot has created life!",
		Color:       helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Gender",
				Value:  fpObj.Gender,
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  fmt.Sprintf("%s %s %s", fpObj.Name.Title, fpObj.Name.First, fpObj.Name.Last),
				Inline: true,
			},
			{
				Name:   "DOB",
				Value:  dob[0],
				Inline: true,
			},
			{
				Name:   "Age",
				Value:  fmt.Sprintf("%v", fpObj.Dob.Age),
				Inline: true,
			},
			{
				Name: "Address",
				Value: fmt.Sprintf(
					"%v %s\n%s, %s, %v %s", fpObj.Location.Street.Number, fpObj.Location.Street.Name, fpObj.Location.City, fpObj.Location.State, fpObj.Location.Postcode,
					fpObj.Location.Country,
				),
				Inline: false,
			},
			{
				Name:   "Email",
				Value:  fpObj.Email,
				Inline: true,
			},
			{
				Name:   "Username",
				Value:  fpObj.Login.Username,
				Inline: true,
			},
			{
				Name:   "Password",
				Value:  fpObj.Login.Password,
				Inline: true,
			},
			{
				Name:   "Phone",
				Value:  fpObj.Phone,
				Inline: true,
			},
			{
				Name:   "Cell",
				Value:  fpObj.Cell,
				Inline: true,
			},
			{
				Name:   fpObj.ID.Name,
				Value:  fpObj.ID.Value,
				Inline: true,
			},
			{
				Name:   "Nationality",
				Value:  fpObj.Nat,
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fpObj.Picture.Large,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Generated with randomuser.me",
		},
	}

	return embed
}

//endregion

//region Img Commands

func sendImgResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var user *discordgo.User
	var imgName string
	var bufferImage []byte
	var err error

	errRespMsg := "Unable to edit image at this moment, please try later :("

	switch options.Name {
	case "pixelate":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Pixelate(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Pixelate.png"

	case "mirror":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Mirror(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Mirror.png"

	case "flip-image":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.FlipImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "FlipImage.png"

	case "colors":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Colors(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Colors.png"

	case "murica":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.America(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "America.png"

	case "communism":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Communism(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Communism.png"

	case "triggered":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Triggered(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Triggered.png"

	case "expand":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)

		}

		bufferImage, err = client.ExpandImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "ExpandImage.png"

	case "wasted":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)

		}

		bufferImage, err = client.Wasted(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Wasted.png"

	case "sketch":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)

		}

		bufferImage, err = client.Sketch(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sketch.png"

	case "spin":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.SpinImage(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "SpinImage.png"

	case "petpet":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.PetPet(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "PetPet.png"

	case "bonk":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Bonk(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Bonk.png"

	case "bomb":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Bomb(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Bomb.png"

	case "shake":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Shake(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Shake.png"

	case "invert":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Invert(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Invert.png"

	case "sobel":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Sobel(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sobel.png"

	case "hog":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Hog(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Hog.png"

	case "triangle":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Triangle(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Triangle.png"

	case "blur":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Blur(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Blur.png"

	case "rgb":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.RGB(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "RGB.png"

	case "angel":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Angel(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Angel.png"

	case "satan":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Satan(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Satan.png"

	case "delete":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Delete(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Delete.png"

	case "fedora":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Fedora(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Fedora.png"

	case "hitler":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Hitler(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Hitler.png"

	case "lego":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Lego(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Lego.png"

	case "wanted":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Wanted(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Wanted.png"

	case "stringify":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Stringify(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Stringify.png"

	case "burn":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Burn(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Burn.png"

	case "earth":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Earth(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Earth.png"

	case "freeze":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Freeze(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Freeze.png"

	case "ground":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Ground(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Ground.png"

	case "mosiac":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Mosiac(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Mosiac.png"

	case "sithlord":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Sithlord(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sithlord.png"

	case "jail":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Jail(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Jail.png"

	case "shatter":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Shatter(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Shatter.png"

	case "pride":
		flag := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Pride(user.AvatarURL("300"), flag)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "pride.png"

	case "trash":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Trash(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Trash.png"

	case "deepfry":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Deepfry(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "deepfry.png"

	case "ascii":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Ascii(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Ascii.png"

	case "charcoal":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Charcoal(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Charcoal.png"

	case "posterize":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Posterize(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Posterize.png"

	case "sepia":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Sepia(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Sepia.png"

	case "swirl":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Swirl(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Swirl.png"

	case "paint":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Paint(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Paint.png"

	case "night":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Night(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "night.png"

	case "rainbow":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Rainbow(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Rainbow.png"

	case "magik":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Magik(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "Magik.png"

	case "5guys1girl":
		guy := options.Options[0].UserValue(s)
		girl := options.Options[1].UserValue(s)

		bufferImage, err = client.FivegOneg(guy.AvatarURL("300"), girl.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "fiveGuys.png"

	case "slap":
		slapped := options.Options[0].UserValue(s)
		slapper := options.Options[1].UserValue(s)

		bufferImage, err = client.Slap(slapper.AvatarURL("300"), slapped.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "slap.png"

	case "obama":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "obama.png"

	case "tweet":
		tweet := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Tweet(user.AvatarURL("300"), user.Username, tweet)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "tweet.png"

	case "youtube":
		comment := options.Options[0].StringValue()
		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.YouTubeComment(user.AvatarURL("300"), user.Username, comment, false)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "youtube.png"

	case "discord":
		msg := options.Options[0].StringValue()
		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Discord(user.AvatarURL("300"), user.Username, msg, true)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "discord.png"

	case "retro-meme":
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 3:
			user = options.Options[2].UserValue(s)
		}

		bufferImage, err = client.Retromeme(user.AvatarURL("300"), topText, bottomText)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "retro-meme.png"

	case "motivational":
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 3:
			user = options.Options[2].UserValue(s)
		}

		bufferImage, err = client.Motivational(user.AvatarURL("300"), topText, bottomText)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "motivational.png"

	case "modern-meme":
		text := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 2:
			user = options.Options[1].UserValue(s)
		}

		bufferImage, err = client.Modernmeme(user.AvatarURL("300"), text)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "modern-meme.png"

	case "why_are_you_gay":
		user1 := options.Options[0].UserValue(s)
		user2 := options.Options[1].UserValue(s)

		bufferImage, err = client.WhyAreYouGay(user1.AvatarURL("300"), user2.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "why_are_you_gay.png"

	case "elmo":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Elmo(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "elmo.png"

	case "tv-static":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.TvStatic(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "static.png"

	case "rain":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Rain(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "rain.png"

	case "glitch":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Glitch(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "glitch.png"

	case "sȶǟȶɨƈ-ɢʟɨȶƈɦ":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}
		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.GlitchStatic(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "static.png"

	case "album":
		switch len(options.Options) {
		case 0:
			user, err = s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

		case 1:
			user = options.Options[0].UserValue(s)
		}

		bufferImage, err = client.Album(user.AvatarURL("300"))
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, errRespMsg)
			}()
			return err
		}

		imgName = "album.png"

	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files: []*discordgo.File{
					{
						Name:        imgName,
						ContentType: "image",
						Reader:      bytes.NewReader(bufferImage),
					},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}

//endregion

//region RateThis Commands

func sendRateThisResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options[0]
	user := fmt.Sprintf("<@!%s>", i.Member.User.ID)

	if len(options.Options) == 1 {
		userName := options.Options[0].UserValue(s)
		user = fmt.Sprintf("<@!%s>", userName.ID)
	}

	embed, err := getRateThisEmbed(options.Name, user)
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to Rate atm, try again later.")
		}()
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		},
	)

	return nil
}

func getRateThisEmbed(ratingName string, user string) (*discordgo.MessageEmbed, error) {
	score := strconv.Itoa(rand.Intn(100))
	rateTitle, rateDesc := getRateTitleAndDesc(ratingName, user, score)

	embed := &discordgo.MessageEmbed{
		Title:       rateTitle,
		Description: rateDesc,
		Color:       helper.RangeIn(1, 16777215),
	}

	return embed, nil
}

func getRateTitleAndDesc(ratingName string, user string, score string) (string, string) {
	switch ratingName {
	case "simp":
		return "Rate This Simp", fmt.Sprintf("%s's Simp Score: %s/100", user, score)
	case "dank":
		return "Dank Rating", fmt.Sprintf("%s's Dank Score: %s/100", user, score)
	case "epicgamer":
		return "Rate This Epic Gamer", fmt.Sprintf("%s's Epic Gamer Score: %s/100", user, score)
	case "gay":
		return "Gay Rating", fmt.Sprintf("%s's Gay Score: %s/100", user, score)
	case "schmeat":
		size := helper.RangeIn(1, 15)
		schmeat := "C" + strings.Repeat("=", size) + "8"
		return "Schmeat Size", fmt.Sprintf("%s's Thang Thangin' \n%s", user, schmeat)
	case "stinky":
		return "Rate This Stinky", fmt.Sprintf("%s's Stinky Score: %s/100", user, score)
	case "thot":
		return "Rate This Thot", fmt.Sprintf("%s's Thot Score: %s/100", user, score)
	case "pickme":
		return "Rate This Pick-Me", fmt.Sprintf("%s's Pick-Me Score: %s/100", user, score)
	case "neckbeard":
		return "Rate This Neck Beard", fmt.Sprintf("%s's Neck Beard Score: %s/100", user, score)
	case "looks":
		return "Rate These Looks", fmt.Sprintf("%s's Looks Score: %s/100", user, score)
	case "smarts":
		return "Rate These Smarts", fmt.Sprintf("%s's Smarts Score: %s/100", user, score)
	case "nerd":
		return "Rate This Nerd", fmt.Sprintf("%s's Nerd Score: %s/100", user, score)
	case "geek":
		return "Rate This Geek", fmt.Sprintf("%s's Geek Score: %s/100", user, score)
	default:
		return "", ""
	}
}

//endregion

//region Txt Commands

func sendTxtResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options[0]

	var err error
	content := ""

	switch options.Name {
	case "clapback":
		text := options.Options[0].StringValue()
		content = strings.ReplaceAll(text, " ", " 👏 ") + " 👏"

	case "bubble", "1337", "cursive", "flipped", "cursed":
		text := strings.ToLower(options.Options[0].StringValue())
		content, err = helper.ToConvertedText(text, options.Name)
		if err != nil {
			go func() {
				err = helper.SendResponseError(s, i, "Unable to convert text atm, try again later.")
			}()
			return err
		}

	case "emojiletters":
		text := strings.ToLower(options.Options[0].StringValue())
		words := strings.Split(text, " ")

		for _, v := range words {
			replacer := strings.NewReplacer(
				"a", "🅰️ ", "b", "🅱️ ", "c", "🇨 ", "d", "🇩 ", "e", "🇪 ", "f", "🇫 ", "g", "🇬 ", "h", "🇭 ", "i", "ℹ️ ", "j", "🇯 ", "k", "🇰 ", "l", "🇱 ", "m", "〽️",
				"n", "🇳 ", "o", "⭕ ", "p", "🅿️ ", "q", "🇶 ", "r", "🇷 ", "s", "🇸 ", "t", "🇹 ", "u", "🇺 ", "v", "🇻 ", "w", "🇼 ", "x", "❎ ", "y", "🇾 ", "z", "🇿 ",
				"0", " ️0️⃣ ", "1", "1️⃣ ", "2", "2️⃣ ", "3", "3️⃣ ", "4", "4️⃣ ", "5", "5️⃣ ", "6", "6️⃣ ", "7", "7️⃣ ", "8", "8️⃣ ", "9", "9️⃣ ",
				"?", "❓ ", "!", "❗ ", "#", "#️⃣ ", "*", "✳️ ", "$", "💲 ", "<", "⏪ ", ">", "⏩ ", "-", "➖ ", "--", "➖ ", "+", "➕ ",
			)
			v = replacer.Replace(v)
			content = fmt.Sprintf("%s%s   ", content, v)
		}

	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

//endregion

//region Daily Commands

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi

	optionName := i.ApplicationCommandData().Options[0].Name
	var err error
	var data *discordgo.InteractionResponseData

	switch optionName {
	case "advice":
		data, err = getDailyEmbed(cfg, getDailyAdviceEmbed)
	case "kanye":
		data, err = getDailyEmbed(cfg, getDailyKanyeEmbed)
	case "affirmation":
		data, err = getDailyEmbed(cfg, getDailyAffirmationEmbed)
	case "horoscope":
		data = getHoroscopeResponseData()
	case "fact":
		var clientData interface{}
		clientData, err = client.Fact()
		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s", clientData),
		}
	default:
		return fmt.Errorf("unknown option: %s", optionName)
	}
	if err != nil {
		go func() {
			_ = helper.SendResponseError(s, i, "Unable to fetch Daily command atm. Try again later.")
		}()
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func getDailyEmbed(cfg *config.Configs, embedFunc func(*config.Configs) (*discordgo.MessageEmbed, error)) (*discordgo.InteractionResponseData, error) {
	embed, err := embedFunc(cfg)
	if err != nil {
		return nil, err
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			embed,
		},
	}, nil
}

func getHoroscopeResponseData() *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: "Choose a zodiac sign",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "horo-select",
						Placeholder: "Zodiac",
						Options: []discordgo.SelectMenuOption{
							{Label: "Aquarius", Value: "aquarius", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🌊"}},
							{Label: "Aries", Value: "aries", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🐏"}},
							{Label: "Cancer", Value: "cancer", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🦀"}},
							{Label: "Capricorn", Value: "capricorn", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🐐"}},
							{Label: "Gemini", Value: "gemini", Default: false, Emoji: discordgo.ComponentEmoji{Name: "♊"}},
							{Label: "Leo", Value: "leo", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🦁"}},
							{Label: "Libra", Value: "libra", Default: false, Emoji: discordgo.ComponentEmoji{Name: "⚖️"}},
							{Label: "Pisces", Value: "pisces", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🐠"}},
							{Label: "Sagittarius", Value: "sagittarius", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🏹"}},
							{Label: "Scorpio", Value: "scorpio", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🦂"}},
							{Label: "Taurus", Value: "taurus", Default: false, Emoji: discordgo.ComponentEmoji{Name: "🐃"}},
							{Label: "Virgo", Value: "virgo", Default: false, Emoji: discordgo.ComponentEmoji{Name: "♍"}},
						},
					},
				},
			},
		},
	}
}

func getDailyAdviceEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.AdviceAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var adviceObj advice
	err = json.NewDecoder(res.Body).Decode(&adviceObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "( ಠ ͜ʖರೃ)",
		Color:       helper.RangeIn(1, 16777215),
		Description: adviceObj.Slip.Advice,
	}

	return embed, nil
}

func getDailyKanyeEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.KanyeAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var kanyeObj kanye
	err = json.NewDecoder(res.Body).Decode(&kanyeObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  fmt.Sprintf("\"%s\"", kanyeObj.Quote),
				Value: "- Kanye",
			},
		},
	}

	return embed, nil
}

func getDailyAffirmationEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.AffirmationAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var affirmObj affirmation
	err = json.NewDecoder(res.Body).Decode(&affirmObj)
	if err != nil {
		return nil, err
	}

	current := time.Now()
	timeFormat := current.Format("Jan 2, 2006")

	embed := &discordgo.MessageEmbed{
		Title:       timeFormat,
		Color:       helper.RangeIn(1, 16777215),
		Description: affirmObj.Affirmation,
	}

	return embed, nil
}

func getHoroscopeEmbed(sign string) (*discordgo.MessageEmbed, error) {
	signNum := getSignNumber(sign)
	if signNum == "" {
		return nil, fmt.Errorf("invalid sign: %s", sign)
	}

	horoscope, err := scrapeHoroscope(signNum)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       sign,
		Description: horoscope,
		Color:       helper.RangeIn(1, 16777215),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Via Horoscopes.com",
			IconURL: "https://www.horoscope.com/images-US/horoscope-logo.svg",
		},
	}

	return embed, nil
}

func scrapeHoroscope(signNum string) (string, error) {
	c := colly.NewCollector()

	var horoscope string
	found := false

	todayFormat := time.Now().Format("Jan 2, 2006")
	yesterdayFormat := time.Now().AddDate(0, 0, -1).Format("Jan 2, 2006")

	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			if !found {
				if strings.Contains(e.Text, todayFormat) || strings.Contains(e.Text, yesterdayFormat) {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		return "", fmt.Errorf("failed to scrape horoscope: %v", err)
	}

	return horoscope, nil
}

func getSignNumber(sign string) string {
	sign = strings.ToLower(sign)
	signNum := ""

	switch sign {
	case "aries":
		signNum = "1"
	case "taurus":
		signNum = "2"
	case "gemini":
		signNum = "3"
	case "cancer":
		signNum = "4"
	case "leo":
		signNum = "5"
	case "virgo":
		signNum = "6"
	case "libra":
		signNum = "7"
	case "scorpio":
		signNum = "8"
	case "sagittarius":
		signNum = "9"
	case "capricorn":
		signNum = "10"
	case "aquarius":
		signNum = "11"
	case "pisces":
		signNum = "12"
	default:
		signNum = ""
	}
	return signNum
}

//endregion

//region Pick Commands

func sendPickResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	option := strings.ToLower(i.ApplicationCommandData().Options[0].Name)

	var data *discordgo.InteractionResponseData
	var err error

	switch option {
	case "steam":
		data, err = sendSteamPickResponse(cfg)
	case "choices":
		data = sendChoicesPickResponse(i)
	case "album":
		data, err = sendAlbumPickResponse(i, cfg)
	case "poll":
		return sendPollResponse(s, i, cfg)
	default:
		return fmt.Errorf("unknown option: %s", option)
	}
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to pick atm, try again later.")
		}()
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func sendSteamPickResponse(cfg *config.Configs) (*discordgo.InteractionResponseData, error) {
	gameURL, err := getSteamGame(cfg)
	if err != nil {
		return nil, err
	}

	data := &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("I have Chosen...\n %s \n☝(°ロ°)☝", gameURL),
	}

	return data, nil
}

func sendChoicesPickResponse(i *discordgo.InteractionCreate) *discordgo.InteractionResponseData {
	content := ""
	for _, v := range i.ApplicationCommandData().Options[0].Options {
		content = content + fmt.Sprintf("[%s] ", v.StringValue())
	}

	content = strings.TrimSpace(content)
	content = fmt.Sprintf("*%s*", content)

	randomIndex := rand.Intn(len(i.ApplicationCommandData().Options[0].Options))
	choice := i.ApplicationCommandData().Options[0].Options[randomIndex].StringValue()

	embed := &discordgo.MessageEmbed{
		Title: "I have chosen...",
		Color: helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   choice,
				Value:  "☝(°ロ°)",
				Inline: true,
			},
		},
	}

	data := &discordgo.InteractionResponseData{
		Content: content,
		Embeds:  []*discordgo.MessageEmbed{embed},
	}

	return data
}

func sendAlbumPickResponse(i *discordgo.InteractionCreate, cfg *config.Configs) (*discordgo.InteractionResponseData, error) {
	var tags []string
	for _, v := range i.ApplicationCommandData().Options[0].Options {
		tags = append(tags, v.StringValue())
	}

	albums, err := callAlbumPickerAPI(cfg, tags, "")
	if err != nil {
		return nil, err
	}

	tagsStr := strings.Join(tags, ", ")

	data := &discordgo.InteractionResponseData{
		Content: "Here are some hand-picked albums",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "album-suggest",
						Placeholder: "Album",
						Options: []discordgo.SelectMenuOption{
							{Label: albums[0].AlbumName, Value: tagsStr + "*{1}*", Default: false},
							{Label: albums[1].AlbumName, Value: tagsStr + "*{2}*", Default: false},
							{Label: albums[2].AlbumName, Value: tagsStr + "*{3}*", Default: false},
							{Label: albums[3].AlbumName, Value: tagsStr + "*{4}*", Default: false},
							{Label: albums[4].AlbumName, Value: tagsStr + "*{5}*", Default: false},
						},
					},
				},
			},
		},
	}

	return data, err
}

func sendPollResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	question := i.ApplicationCommandData().Options[0].Options[0]
	var fields []*discordgo.MessageEmbedField
	var emojis []string

	for _, v := range i.ApplicationCommandData().Options[0].Options {
		emoji := helper.GetRandomStringFromSet(cfg.Emojis)
		if v.Name != "request" {
			f := &discordgo.MessageEmbedField{
				Name:   v.StringValue(),
				Value:  emoji,
				Inline: false,
			}
			fields = append(fields, f)
			emojis = append(emojis, emoji)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:  question.StringValue(),
		Color:  helper.RangeIn(1, 16777215),
		Fields: fields,
	}

	data := &discordgo.InteractionResponseData{
		Content: "Poll Time!",
		Embeds:  []*discordgo.MessageEmbed{embed},
	}

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return err
	}

	err = addPollReactions(emojis, i, s)
	if err != nil {
		return err
	}

	return nil
}

func callAlbumPickerAPI(cfg *config.Configs, tagSlice []string, tagStr string) ([]albumPicker, error) {
	var albumPickerObjs []albumPicker

	urlTags := ""
	if tagStr == "" {
		//we need to separate by commas and spaces and add brackets because API bad
		urlTags = strings.Join(tagSlice, ", ")
	} else {
		urlTags = tagStr
	}

	URL := cfg.Configs.ApiURLs.AlbumPickerAPI + url.PathEscape("["+urlTags+"]")

	res, err := http.Get(URL)
	if err != nil {
		return albumPickerObjs, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&albumPickerObjs)
	if err != nil {
		return albumPickerObjs, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return albumPickerObjs, nil
}

func getAlbumPickerEmbed(tags string, cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	index := getAlbumPickerIndex(tags)

	replacer := strings.NewReplacer("*{1}*", "", "*{2}*", "", "*{3}*", "", "*{4}*", "", "*{5}*", "")
	tags = replacer.Replace(tags)

	albumPickerObj, err := callAlbumPickerAPI(cfg, nil, tags)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Check out these albums!",
		Color: helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: albumPickerObj[index].URL,
		},
		Footer: &discordgo.MessageEmbedFooter{

			Text: "http://www.albumrecommender.com",
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Album Name", Value: helper.CheckIfStructValueISEmpty(albumPickerObj[index].AlbumName), Inline: true},
			{Name: "Album Artist", Value: helper.CheckIfStructValueISEmpty(albumPickerObj[index].Artist), Inline: true},
			{Name: "Genres", Value: helper.CheckIfStructValueISEmpty(albumPickerObj[index].Genres), Inline: false},
			{Name: "Secondary Genres", Value: helper.CheckIfStructValueISEmpty(albumPickerObj[index].SecGenres), Inline: false},
			{Name: "Descriptors", Value: helper.CheckIfStructValueISEmpty(albumPickerObj[index].Descriptors), Inline: false},
		},
	}

	return embed, nil
}

func getAlbumPickerIndex(tags string) int {
	switch {
	case strings.Contains(tags, "*{1}*"):
		return 0
	case strings.Contains(tags, "*{2}*"):
		return 1
	case strings.Contains(tags, "*{3}*"):
		return 2
	case strings.Contains(tags, "*{4}*"):
		return 3
	case strings.Contains(tags, "*{5}*"):
		return 4
	}
	return 0
}

func getSteamGame(cfg *config.Configs) (string, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.SteamAPI)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var steamObj steamGames
	err = json.NewDecoder(res.Body).Decode(&steamObj)
	if err != nil {
		return "", err
	}

	randomIndex := rand.Intn(len(steamObj.Applist.Apps))
	for steamObj.Applist.Apps[randomIndex].Name == "" {
		randomIndex = rand.Intn(len(steamObj.Applist.Apps))
	}
	//gameURL := fmt.Sprintf("steam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Appid)
	gameURL := fmt.Sprintf("https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Appid)

	return gameURL, nil
}

func addPollReactions(emojis []string, i *discordgo.InteractionCreate, s *discordgo.Session) error {
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		return err
	}

	pollMsgID := channel.LastMessageID

	for _, v := range emojis {
		err = s.MessageReactionAdd(channel.ID, pollMsgID, v)
		if err != nil {
			err = fmt.Errorf("Emoji: %s \n %s", v, err)
			return err
		}
	}

	return nil
}

//endregion

//region Component Commands

func sendHoroscopeCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	sign := i.MessageComponentData().Values[0]

	embed, err := getHoroscopeEmbed(sign)
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to fetch Horoscope atm, try again later.")
		}()
		return err
	}

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	msgContent := ""
	msgEdit.Content = &msgContent
	msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

	// edit response (i.Interaction) and replace with embed
	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return err
	}

	// 'This interaction failed' will show if not included
	// todo fix later
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot send an empty message") {
			return err
		}
	}

	return nil
}

func sendAlbumPickCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	tags := i.MessageComponentData().Values[0]

	embed, err := getAlbumPickerEmbed(tags, cfg)
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to fetch Albums atm, try again later.")
		}()
		return err
	}

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	msgContent := ""
	msgEdit.Content = &msgContent
	msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

	// edit response (i.Interaction) and replace with embed
	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return err
	}

	// 'This interaction failed' will show if not included
	// todo fix later
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot send an empty message") {
			return err
		}
	}

	return nil
}

func sendWYRCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	embed, err := getWYREmbed(cfg)
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to fetch WYR atm, try again later.")
		}()
		return err
	}

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	msgContent := ""
	msgEdit.Content = &msgContent
	msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

	// edit response (i.Interaction) and replace with embed
	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return err
	}

	// 'This interaction failed' will show if not included
	// todo fix later
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot send an empty message") {
			return err
		}
	}

	return nil
}

//endregion
