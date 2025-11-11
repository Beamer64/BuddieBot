package slash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/chromedp"
	"github.com/mitchellh/mapstructure"
)

func sendGetResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var embed *discordgo.MessageEmbed
	var data *discordgo.MessageSend
	var err error

	errRespMsg := "Unable to fetch data atm, Try again later."

	// defer the interaction response to avoid timeout
	// sends a "Bot is thinking..." message
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	)
	if err != nil {
		return fmt.Errorf("error sending deferred Interaction for /get command %s: %v", options.Name, err)
	}

	switch options.Name {
	case "rekd":
		clientData, err := client.Roast()
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
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

		data = &discordgo.MessageSend{
			Content: fmt.Sprintf("%s\n(ง ͠° ͟ل͜ ͡°)ง", content),
		}

	case "landsat":
		text := options.Options[0].StringValue()
		embed, err = getLandSatImageEmbed(cfg, text)

		data = &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "joke":
		clientData, err := client.Joke()
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		var jokeObj joke
		err = mapstructure.Decode(clientData, &jokeObj)
		if err != nil {
			return err
		}

		data = &discordgo.MessageSend{
			Content: fmt.Sprintf("%s", jokeObj.Joke),
		}

	case "8ball":
		clientData, err := client.Eightball()
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.MessageSend{
			Content: fmt.Sprintf("%s", clientData),
		}

	case "yomomma":
		clientData, err := client.Yomama()
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
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

		data = &discordgo.MessageSend{
			Content: content,
		}

	case "pickup-line":
		clientData, err := client.PickupLine()
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		var pickupObj pickupLine
		err = mapstructure.Decode(clientData, &pickupObj)
		if err != nil {
			return err
		}

		data = &discordgo.MessageSend{
			Content: fmt.Sprintf("%s", pickupObj.Joke),
		}

	case "fake-person":
		personData, err := callFakePersonAPI(cfg)
		if err != nil {
			go func() {
				_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		embed = getFakePersonEmbed(personData)

		data = &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "xkcd":
		embed, err = getXkcdEmbed(cfg)

		data = &discordgo.MessageSend{
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

	// delete the interaction response
	err = s.InteractionResponseDelete(i.Interaction)
	if err != nil {
		return err
	}

	// send the new response
	// data must be of type *discordgo.MessageSend
	_, err = s.ChannelMessageSendComplex(i.ChannelID, data)
	if err != nil {
		return fmt.Errorf("error sending Interaction for command %s: %v", options.Name, err)
	}

	return nil
}

func getLandSatImageEmbed(cfg *config.Configs, text string) (*discordgo.MessageEmbed, error) {
	imgPath, err := getLandsatImage(cfg, text)
	if err != nil {
		return nil, err
	}

	imgURL, err := helper.GetImgbbUploadURL(cfg, imgPath, 10)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Landsat, more like...landFLAT...amirite non-round supporters??.",
		Color: helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: imgURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    cfg.Configs.ApiURLs.LandsatAPI,
			IconURL: imgURL,
		},
	}

	return embed, nil
}

func getLandsatImage(cfg *config.Configs, text string) (string, error) {
	landsatUrl := cfg.Configs.ApiURLs.LandsatAPI

	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	// Navigate to the page, insert text, and click the button
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(landsatUrl),
		chromedp.WaitVisible(`#nameInput`),
		chromedp.SendKeys(`#nameInput`, text, chromedp.NodeVisible),
		chromedp.WaitVisible(`#enterButton`),
		chromedp.Click(`#enterButton`),
		chromedp.Sleep(5*time.Second),
		chromedp.Screenshot("#nameBoxes", &buf, chromedp.NodeVisible),
	)
	if err != nil {
		return "", err
	}

	filename := "../../res/genFiles/landSat.png"
	if err = os.WriteFile(filename, buf, 0644); err != nil {
		return "", err
	}

	return filename, nil
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	imgURL := "https:"
	doc.Find("#comic img").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			if src, exists := s.Attr("src"); exists {
				imgURL = imgURL + src
				return false // Stop after finding the first image
			}
			return true
		},
	)

	embed := &discordgo.MessageEmbed{
		Title: "People used to read comics.",
		Color: helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: imgURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Generated with xkcd.com",
			IconURL: imgURL,
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
