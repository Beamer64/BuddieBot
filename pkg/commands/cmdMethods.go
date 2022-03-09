package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/games"
	"github.com/beamer64/godagpi/dagpi"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"github.com/mitchellh/mapstructure"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// functions here should mostly be used for the slash commands

//region Utility Commands

func sendTuuckResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	n := reflect.ValueOf(&cfg.Cmd.SlashName).Elem()
	d := reflect.ValueOf(&cfg.Cmd.Desc).Elem()
	e := reflect.ValueOf(&cfg.Cmd.Example).Elem()

	if len(i.ApplicationCommandData().Options) > 0 {
		cmdOption := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())
		slashCmd := ""
		cmdName := ""
		if strings.Contains(cmdOption, "/") {
			slashCmd = cmdOption
			cmdName = strings.ReplaceAll(slashCmd, "/", "")
		} else {
			slashCmd = fmt.Sprintf("/%s", cmdOption)
			cmdName = cmdOption
		}

		title := ""
		for t := 0; t < n.NumField(); t++ {
			if strings.Contains(fmt.Sprintf("%s", n.Field(t).Interface()), cmdName) {
				title = fmt.Sprintf("%s info", n.Field(t).Interface())
				break
			}
		}

		desc := ""
		for de := 0; de < d.NumField(); de++ {
			cmdDesc := strings.ReplaceAll(cmdName, " ", "")
			lowerDesc := strings.ToLower(d.Type().Field(de).Name)
			if strings.Contains(lowerDesc, cmdDesc) {
				desc = fmt.Sprintf("%s", d.Field(de).Interface())
				break
			}
		}

		example := ""
		for ex := 0; ex < e.NumField(); ex++ {
			cmdExample := strings.ReplaceAll(cmdName, " ", "")
			lowerExample := strings.ToLower(e.Type().Field(ex).Name)
			if strings.Contains(lowerExample, cmdExample) {
				example = fmt.Sprintf("%s", e.Field(ex).Interface())
				break
			}
		}
		usage := fmt.Sprintf("`%s`", slashCmd)

		if title == "" {
			title = fmt.Sprintf("Invalid Command: %s", cmdOption)
		} else if desc == "" {
			desc = "Command not found"
			usage = "N/A"
		} else if example == "" {
			example = "Command not found"
		}

		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Description",
					Value:  desc,
					Inline: false,
				},
				{
					Name:   "Usage",
					Value:  usage,
					Inline: false,
				},
				{
					Name:   "Example",
					Value:  example,
					Inline: false,
				},
			},
		}
		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	} else {
		content := "A list of current Slash command groups\n```\n"

		for i := 0; i < n.NumField(); i++ {
			content = content + fmt.Sprintf("%s \n", n.Field(i).Interface())
		}
		content = content + "```\nYou can get more information about a command by using /tuuck <command_name>"

		err := s.InteractionRespond(
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
	}

	return nil
}

//endregion

//region Play Commands

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, client dagpi.Client) error {
	options := i.ApplicationCommandData().Options[0]

	switch options.Name {
	case "coin-flip":
		embed, err := getCoinFlipEmbed(cfg)
		if err != nil {
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
		if err != nil {
			return err
		}

	// todo finish this
	case "nim":
		err := games.PlayNIM(s, i, cfg, "")
		if err != nil {
			return err
		}

	case "typeracer":

	case "gtl":

	case "wtp":
		data, err := client.WTP()
		if err != nil {
			return err
		}

		embed, err := getWTPembed(data, false)
		if err != nil {
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
		if err != nil {
			return err
		}

		timer := time.NewTimer(60 * time.Second)
		go func() {
			<-timer.C
			fmt.Println("You took too long to respond!")
		}()
	}

	return nil
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
	gifURL, err := api.RequestGifURL("Coin Flip", cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Flipping...",
		Color: 16761856,
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}

	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	search := ""
	results := ""
	if randNum%2 == 0 {
		search = "Coin Flip Heads"
		results = "Heads"
		gifURL, err = api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return nil, err
		}

	} else {
		search = "Coin Flip Tails"
		results = "Tails"
		gifURL, err = api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return nil, err
		}
	}

	embed.Description = fmt.Sprintf("It's %s!", results)
	embed.Image = &discordgo.MessageEmbedImage{
		URL: gifURL,
	}

	return embed, nil
}

//endregion

//region Animal Commands

func sendAnimalsResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	switch i.ApplicationCommandData().Options[0].Name {
	case "doggo":
		embed, err := getDoggoEmbed(cfg)
		if err != nil {
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
		if err != nil {
			return err
		}
	}

	return nil
}

func getDoggoEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	// a data scientist had to fix this...
	doggoObj, err := callDoggoAPI(cfg)
	if err != nil {
		return nil, err
	}

	if len(doggoObj[0].Breeds) < 1 {
		doggoObj, err = callDoggoAPI(cfg)
		if err != nil {
			return nil, err
		}
	}

	impWeight := checkIfEmpty(doggoObj[0].Breeds[0].Weight.Imperial)
	metWeight := checkIfEmpty(doggoObj[0].Breeds[0].Weight.Metric)
	impHeight := checkIfEmpty(doggoObj[0].Breeds[0].Height.Imperial)
	metHeight := checkIfEmpty(doggoObj[0].Breeds[0].Height.Metric)

	embed := &discordgo.MessageEmbed{
		Title:       doggoObj[0].Breeds[0].Name,
		Color:       rangeIn(1, 16777215),
		Description: doggoObj[0].Breeds[0].Temperament,
		Image: &discordgo.MessageEmbedImage{
			URL: doggoObj[0].URL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Weight",
				Value:  fmt.Sprintf("%s lbs / %s kg", impWeight, metWeight),
				Inline: true,
			},
			{
				Name:   "Height",
				Value:  fmt.Sprintf("%s in / %s cm", impHeight, metHeight),
				Inline: true,
			},
			{
				Name:   "Origin",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].Origin),
				Inline: true,
			},
			{
				Name:   "Bred For",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].BredFor),
				Inline: true,
			},
			{
				Name:   "Breed Group",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].BreedGroup),
				Inline: true,
			},
			{
				Name:   "Life Span",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].LifeSpan),
				Inline: true,
			},
		},
	}

	return embed, nil
}

func callDoggoAPI(cfg *config.Configs) (doggo, error) {
	res, err := http.Get(cfg.Configs.Keys.DoggoAPI)
	if err != nil {
		return nil, err
	}

	var doggoObj doggo

	err = json.NewDecoder(res.Body).Decode(&doggoObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return doggoObj, nil
}

//endregion

//region Get Commands

func sendGetResponse(s *discordgo.Session, i *discordgo.InteractionCreate, client dagpi.Client) error {
	options := i.ApplicationCommandData().Options[0]

	switch options.Name {

	case "insult":
		insultMsg, err := client.Roast()
		if err != nil {
			return err
		}

		content := ""
		switch len(options.Options) {
		case 0:
			content = fmt.Sprintf("<@!%s>\n%s", i.Member.User.ID, insultMsg)

		case 1:
			user := options.Options[0].UserValue(s)

			content = fmt.Sprintf("<@!%s>\n%s", user.ID, insultMsg)
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s\n(ง ͠° ͟ل͜ ͡°)ง", content),
				},
			},
		)
		if err != nil {
			return err
		}

	case "joke":
		data, err := client.Joke()
		if err != nil {
			return err
		}

		var jokeObj joke
		err = mapstructure.Decode(data, &jokeObj)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s", jokeObj.Joke),
				},
			},
		)
		if err != nil {
			return err
		}

	case "8ball":
		data, err := client.Eightball()
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s", data),
				},
			},
		)
		if err != nil {
			return err
		}

	case "yomomma":
		data, err := client.Yomama()
		if err != nil {
			return err
		}

		content := ""
		switch len(options.Options) {
		case 0:
			content = fmt.Sprintf("<@!%s>\n%s", i.Member.User.ID, data)

		case 1:
			user := options.Options[0].UserValue(s)

			content = fmt.Sprintf("<@!%s>\n%s", user.ID, data)
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

	case "pickup-line":
		data, err := client.PickupLine()
		if err != nil {
			return err
		}

		var pickupObj pickupLine
		err = mapstructure.Decode(data, &pickupObj)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s", pickupObj.Joke),
				},
			},
		)
		if err != nil {
			return err
		}

		/*case "captcha":
		data, err := client.WTP()
		if err != nil {
			return err
		}*/

	}

	return nil
}

//endregion

//region Img Commands

func sendImgResponse(s *discordgo.Session, i *discordgo.InteractionCreate, client dagpi.Client) error {
	options := i.ApplicationCommandData().Options[0]

	switch options.Name {
	case "pixelate":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Pixelate(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Pixelate(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Pixelate.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "mirror":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Mirror(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Mirror(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Mirror.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "flip-image":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.FlipImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.FlipImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "FlipImage.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "colors":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Colors(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Colors(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Colors.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "murica":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.America(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.America(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "America.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "communism":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Communism(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Communism(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Communism.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "triggered":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Triggered(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Triggered(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Triggered.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "expand":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.ExpandImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.ExpandImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "ExpandImage.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "wasted":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Wasted(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Wasted(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Wasted.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "sketch":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Sketch(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Sketch(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Sketch.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "spin":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.SpinImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.SpinImage(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "SpinImage.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "petpet":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.PetPet(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.PetPet(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "PetPet.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "bonk":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Bonk(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Bonk(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Bonk.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "bomb":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Bomb(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Bomb(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Bomb.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "shake":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Shake(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Shake(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Shake.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "invert":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Invert(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Invert(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Invert.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "sobel":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Sobel(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Sobel(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Sobel.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "hog":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Hog(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Hog(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Hog.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "triangle":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Triangle(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Triangle(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Triangle.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "blur":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Blur(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Blur(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Blur.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "rgb":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.RGB(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.RGB(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "RGB.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "angel":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Angel(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Angel(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Angel.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "satan":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Satan(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Satan(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Satan.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "delete":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Delete(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Delete(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Delete.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "fedora":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Fedora(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Fedora(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Fedora.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "hitler":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Hitler(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Hitler(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Hitler.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "lego":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Lego(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Lego(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Lego.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "wanted":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Wanted(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Wanted(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Wanted.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "stringify":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Stringify(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Stringify(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Stringify.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "burn":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Burn(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Burn(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Burn.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "earth":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Earth(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Earth(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Earth.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "freeze":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Freeze(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Freeze(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Freeze.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "ground":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Ground(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Ground(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Ground.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "mosiac":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Mosiac(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Mosiac(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Mosiac.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "sithlord":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Sithlord(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Sithlord(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Sithlord.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "jail":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Jail(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Jail(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Jail.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "shatter":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Shatter(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Shatter(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Shatter.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "pride":
		var buffer []byte
		flag := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			buffer, err = client.Pride(user.AvatarURL("300"), flag)
			if err != nil {
				return err
			}

		case 2:
			user := options.Options[1].UserValue(s)

			bufferImg, err := client.Pride(user.AvatarURL("300"), flag)
			if err != nil {
				return err
			}

			buffer = bufferImg
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "pride.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "trash":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Trash(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Trash(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Trash.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "deepfry":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Deepfry(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Deepfry(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "deepfry.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "ascii":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Ascii(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Ascii(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Ascii.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "charcoal":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Charcoal(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Charcoal(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Charcoal.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "posterize":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Posterize(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Posterize(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Posterize.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "sepia":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Sepia(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Sepia(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Sepia.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "swirl":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Swirl(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Swirl(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Swirl.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "paint":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Paint(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Paint(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Paint.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "night":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Night(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Night(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "night.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "rainbow":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Rainbow(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Rainbow(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Rainbow.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "magik":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Magik(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Magik(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "Magik.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "5guys1girl":
		guy := options.Options[0].UserValue(s)
		girl := options.Options[1].UserValue(s)

		buffer, err := client.FivegOneg(guy.AvatarURL("300"), girl.AvatarURL("300"))
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "fiveGuys.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "slap":
		slapped := options.Options[0].UserValue(s)
		slapper := options.Options[1].UserValue(s)

		buffer, err := client.Slap(slapper.AvatarURL("300"), slapped.AvatarURL("300"))
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "slap.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "obama":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Obama(user.AvatarURL("300"), user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "obama.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "tweet":
		var buffer []byte
		tweet := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			buffer, err = client.Tweet(user.AvatarURL("300"), user.Username, tweet)
			if err != nil {
				return err
			}

		case 2:
			user := options.Options[1].UserValue(s)

			bufferImage, err := client.Tweet(user.AvatarURL("300"), user.Username, tweet)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "tweet.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "youtube":
		comment := options.Options[0].StringValue()
		var buffer []byte

		switch len(options.Options) {
		case 1:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			buffer, err = client.YouTubeComment(user.AvatarURL("300"), user.Username, comment, false)
			if err != nil {
				return err
			}

		case 2:
			user := options.Options[1].UserValue(s)

			bufferImage, err := client.YouTubeComment(user.AvatarURL("300"), user.Username, comment, false)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "youtube.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "discord":
		msg := options.Options[0].StringValue()
		var buffer []byte

		switch len(options.Options) {
		case 1:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			buffer, err = client.Discord(user.AvatarURL("300"), user.Username, msg, true)
			if err != nil {
				return err
			}

		case 2:
			user := options.Options[1].UserValue(s)

			bufferImage, err := client.Discord(user.AvatarURL("300"), user.Username, msg, true)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "discord.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "retro-meme":
		var buffer []byte
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Retromeme(user.AvatarURL("300"), topText, bottomText)
			if err != nil {
				return err
			}

			buffer = bufferImage

		case 3:
			user := options.Options[2].UserValue(s)

			bufferImage, err := client.Retromeme(user.AvatarURL("300"), topText, bottomText)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "retro-meme.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "motivational":
		var buffer []byte
		topText := options.Options[0].StringValue()
		bottomText := options.Options[1].StringValue()

		switch len(options.Options) {
		case 2:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Motivational(user.AvatarURL("300"), topText, bottomText)
			if err != nil {
				return err
			}

			buffer = bufferImage

		case 3:
			user := options.Options[2].UserValue(s)

			bufferImage, err := client.Motivational(user.AvatarURL("300"), topText, bottomText)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "motivational.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "modern-meme":
		var buffer []byte
		text := options.Options[0].StringValue()

		switch len(options.Options) {
		case 1:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Modernmeme(user.AvatarURL("300"), text)
			if err != nil {
				return err
			}

			buffer = bufferImage

		case 2:
			user := options.Options[1].UserValue(s)

			bufferImage, err := client.Modernmeme(user.AvatarURL("300"), text)
			if err != nil {
				return err
			}

			buffer = bufferImage
		}
		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "modern-meme.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "why_are_you_gay":
		user1 := options.Options[0].UserValue(s)
		user2 := options.Options[1].UserValue(s)

		buffer, err := client.WhyAreYouGay(user1.AvatarURL("300"), user2.AvatarURL("300"))
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "why_are_you_gay.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "elmo":
		var buffer []byte

		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Elmo(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Elmo(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "elmo.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "tv-static":
		var buffer []byte
		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.TvStatic(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.TvStatic(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "static.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "rain":
		var buffer []byte
		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Rain(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Rain(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "rain.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "glitch":
		var buffer []byte
		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Glitch(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Glitch(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "glitch.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "sȶǟȶɨƈ-ɢʟɨȶƈɦ":
		var buffer []byte
		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.GlitchStatic(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.GlitchStatic(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "static.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "album":
		var buffer []byte
		switch len(options.Options) {
		case 0:
			user, err := s.User(i.Member.User.ID)
			if err != nil {
				return err
			}

			bufferImage, err := client.Album(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		case 1:
			user := options.Options[0].UserValue(s)

			bufferImage, err := client.Album(user.AvatarURL("300"))
			if err != nil {
				return err
			}

			buffer = bufferImage
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Files: []*discordgo.File{
						{
							Name:        "album.png",
							ContentType: "image",
							Reader:      bytes.NewReader(buffer),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	}

	return nil
}

//endregion

//region Daily Commands

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, client dagpi.Client) error {
	switch i.ApplicationCommandData().Options[0].Name {
	case "advice":
		embed, err := getDailyAdviceEmbed(cfg)
		if err != nil {
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
		if err != nil {
			return err
		}

	case "kanye":
		embed, err := getDailyKanyeEmbed(cfg)
		if err != nil {
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
		if err != nil {
			return err
		}

	case "affirmation":
		embed, err := getDailyAffirmationEmbed(cfg)
		if err != nil {
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
		if err != nil {
			return err
		}

	case "horoscope":
		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Choose a zodiac sign",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID:    "horo-select",
									Placeholder: "Zodiac",
									Options: []discordgo.SelectMenuOption{
										{
											Label:   "Aquarius",
											Value:   "aquarius",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🌊"},
										},
										{
											Label:   "Aries",
											Value:   "aries",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐏"},
										},
										{
											Label:   "Cancer",
											Value:   "cancer",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦀"},
										},
										{
											Label:   "Capricorn",
											Value:   "capricorn",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐐"},
										},
										{
											Label:   "Gemini",
											Value:   "gemini",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "♊"},
										},
										{
											Label:   "Leo",
											Value:   "leo",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦁"},
										},
										{
											Label:   "Libra",
											Value:   "libra",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "⚖️"},
										},
										{
											Label:   "Pisces",
											Value:   "pisces",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐠"},
										},
										{
											Label:   "Sagittarius",
											Value:   "sagittarius",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🏹"},
										},
										{
											Label:   "Scorpio",
											Value:   "scorpio",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦂"},
										},
										{
											Label:   "Taurus",
											Value:   "taurus",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐃"},
										},
										{
											Label:   "Virgo",
											Value:   "virgo",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "♍"},
										},
									},
								},
							},
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "fact":
		data, err := client.Fact()
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s", data),
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDailyAdviceEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.Keys.AdviceAPI)
	if err != nil {
		return nil, err
	}

	var adviceObj advice

	err = json.NewDecoder(res.Body).Decode(&adviceObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	embed := &discordgo.MessageEmbed{
		Title:       "( ಠ ͜ʖರೃ)",
		Color:       rangeIn(1, 16777215),
		Description: adviceObj.Slip.Advice,
	}

	return embed, nil
}

func getDailyKanyeEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.Keys.KanyeAPI)
	if err != nil {
		return nil, err
	}

	var kanyeObj kanye

	err = json.NewDecoder(res.Body).Decode(&kanyeObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	embed := &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: rangeIn(1, 16777215),
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
	res, err := http.Get(cfg.Configs.Keys.AffirmationAPI)
	if err != nil {
		return nil, err
	}

	var affirmObj affirmation

	err = json.NewDecoder(res.Body).Decode(&affirmObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	current := time.Now()
	timeFormat := current.Format("Jan 2, 2006")

	embed := &discordgo.MessageEmbed{
		Title:       timeFormat,
		Color:       rangeIn(1, 16777215),
		Description: affirmObj.Affirmation,
	}

	return embed, nil
}

func getHoroscopeEmbed(sign string) (*discordgo.MessageEmbed, error) {
	horoscope := ""
	found := false
	signNum := ""

	switch strings.ToLower(sign) {
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
	}

	c := colly.NewCollector()

	// Before making a request print "Visiting ..."
	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		},
	)

	// this is ugly, and I'd like to do away with it eventually
	today := time.Now()
	todayFormat := today.Format("Jan 2, 2006")

	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayFormat := yesterday.Format("Jan 2, 2006")

	// On every p element which has style attribute call callback
	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			if !found {
				if strings.Contains(e.Text, todayFormat) {
					horoscope = e.Text
					found = true

					// website hasn't updated yet
				} else if strings.Contains(e.Text, yesterdayFormat) {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	// StartServer scraping on https://www.horoscope.com
	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		return nil, nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       sign,
		Description: horoscope,
		Color:       rangeIn(1, 16777215),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Via Horoscopes.com",
			IconURL: "https://www.horoscope.com/images-US/horoscope-logo.svg",
		},
	}

	return embed, nil
}

//endregion

//region Pick Commands

func sendPickResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	switch strings.ToLower(i.ApplicationCommandData().Options[0].Name) {
	case "steam":
		choice, err := getSteamGame(cfg)
		if err != nil {
			return err
		}

		embed := &discordgo.MessageEmbed{
			Title: "I have Chosen...",
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   choice,
					Value:  "☝(°ロ°)",
					Inline: true,
				},
			},
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Choosing Steam Game",
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "choices":
		content := ""
		for _, v := range i.ApplicationCommandData().Options[0].Options {
			content = content + fmt.Sprintf("[%s] ", v.StringValue())
		}
		content = strings.TrimSpace(content)
		content = fmt.Sprintf("*%s*", content)

		rand.Seed(time.Now().UnixNano())

		randomIndex := rand.Intn(len(i.ApplicationCommandData().Options[0].Options))
		choice := i.ApplicationCommandData().Options[0].Options[randomIndex].StringValue()

		embed := &discordgo.MessageEmbed{
			Title: "I have chosen...",
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   choice,
					Value:  "☝(°ロ°)",
					Inline: true,
				},
			},
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getSteamGame(cfg *config.Configs) (string, error) {
	res, err := http.Get(cfg.Configs.Keys.SteamAPI)
	if err != nil {
		return "", err
	}

	var steamObj steamGames

	err = json.NewDecoder(res.Body).Decode(&steamObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)
	if err != nil {
		return "", err
	}

	randomIndex := rand.Intn(len(steamObj.Applist.Apps))
	for steamObj.Applist.Apps[randomIndex].Name == "" {
		randomIndex = rand.Intn(len(steamObj.Applist.Apps))
	}
	choice := fmt.Sprintf("%s\nsteam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Name, steamObj.Applist.Apps[randomIndex].Appid)

	return choice, nil
}

//endregion

//region Component Commands

func sendHoroscopeCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	sign := i.MessageComponentData().Values[0]
	embed, err := getHoroscopeEmbed(sign)
	if err != nil {
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
