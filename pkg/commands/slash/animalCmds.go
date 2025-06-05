package slash

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"io"
	"math/rand"
	"net/http"
)

func sendAnimalsResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	commandName := i.ApplicationCommandData().Options[0].Name

	var embed *discordgo.MessageEmbed
	var data *discordgo.MessageSend
	var err error

	errRespMsg := "Unable to make call at this moment, please try later :("

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	)
	if err != nil {
		return fmt.Errorf("error sending deferred Interaction for /get command %s: %v", commandName, err)
	}

	switch commandName {
	case "doggo":
		embed, err = getDoggoEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "katz":
		embed, err = getKatzEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, errRespMsg)
			}()
			return err
		}

		data = &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}
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
		return fmt.Errorf("error sending Interaction for command %s: %v", commandName, err)
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
