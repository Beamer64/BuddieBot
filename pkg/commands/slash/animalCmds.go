package slash

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"net/http"
)

func sendAnimalsResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	commandName := i.ApplicationCommandData().Options[0].Name
	errRespMsg := "Unable to make call at this moment, please try later :("

	// Defer the interaction to prevent timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get command %s: %w", commandName, err)
	}

	var webhookEdit *discordgo.WebhookEdit
	var err error

	// Map command names to embed functions
	embeds := map[string]func(*config.Configs) (*discordgo.MessageEmbed, error){
		"doggo": getDoggoEmbed,
		"katz":  getKatzEmbed,
	}

	switch commandName {
	case "doggo", "katz":
		var embed *discordgo.MessageEmbed

		embed, err = embeds[commandName](cfg)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
		}
	default:
		return fmt.Errorf("unknown option: %s", commandName)
	}
	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("error in animalCmds.sendAnimalsResponse() : %w", err)
	}

	// Edit the interaction response with the generated data
	if _, err = s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("failed to send message for command %s: %w", commandName, err)
	}

	return nil
}

func getDoggoEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	doggoObj, err := retryDoggoFetch(cfg)
	if err != nil {
		return nil, fmt.Errorf("error retrieving doggo data after max # of attempts: %w", err)
	}

	dog := doggoObj[0]
	breed := dog.Breeds[0]

	get := helper.CheckIfStructValueISEmpty

	embed := &discordgo.MessageEmbed{
		Title:       breed.Name,
		Color:       helper.RangeIn(1, 16777215),
		Description: breed.Temperament,
		Image: &discordgo.MessageEmbedImage{
			URL: dog.URL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Weight", Value: fmt.Sprintf("%slbs / %skg", get(breed.Weight.Imperial), get(breed.Weight.Metric)), Inline: true},
			{Name: "Breed Group", Value: get(breed.BreedGroup), Inline: true},
			{Name: "Origin", Value: get(breed.Origin), Inline: true},
			{Name: "Height", Value: fmt.Sprintf("%sin / %scm", get(breed.Height.Imperial), get(breed.Height.Metric)), Inline: true},
			{Name: "Life Span", Value: get(breed.LifeSpan), Inline: true},
			{Name: "Good Pup", Value: "10/10", Inline: true},
			{Name: "Bred For", Value: get(breed.BredFor), Inline: false},
		},
	}

	return embed, nil
}

// retryDoggoFetch encapsulates retry logic for fetching valid dog data
func retryDoggoFetch(cfg *config.Configs) ([]doggo, error) {
	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		doggoObj, err := callDoggoAPI(cfg)
		if err != nil {
			return nil, err
		}
		if len(doggoObj) > 0 && len(doggoObj[0].Breeds) > 0 {
			return doggoObj, nil
		}
	}
	return nil, fmt.Errorf("no valid dog data after %d attempts", maxAttempts)
}

func callDoggoAPI(cfg *config.Configs) ([]doggo, error) {
	req, err := http.NewRequest("GET", cfg.Configs.ApiURLs.DoggoAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request to Doggo API: %w", err)
	}

	req.Header.Set("x-api-key", cfg.Configs.Keys.DoggoAPIkey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request to Doggo API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("doggo API returned status code %d", resp.StatusCode)
	}

	var doggoObj []doggo
	if err := json.NewDecoder(resp.Body).Decode(&doggoObj); err != nil {
		return nil, fmt.Errorf("decoding Doggo API response: %w", err)
	}

	return doggoObj, nil
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

	// Pick a random cat
	breed := katzObj[rand.Intn(len(katzObj))]

	// Helper to simplify empty value check
	get := helper.CheckIfStructValueISEmpty

	// Compose complex values
	lifeSpan := fmt.Sprintf("%s - %s years", get(breed.MinLifeExpectancy), get(breed.MaxLifeExpectancy))
	weight := fmt.Sprintf("%s - %s lbs", get(breed.MinWeight), get(breed.MaxWeight))

	fields := []*discordgo.MessageEmbedField{
		{Name: "Origin", Value: get(breed.Origin), Inline: true},
		{Name: "Weight", Value: weight, Inline: true},
		{Name: "Life Span", Value: lifeSpan, Inline: true},
		{Name: "Grooming Req", Value: descTiers[breed.Grooming], Inline: true},
		{Name: "Playfulness", Value: descTiers[breed.Playfulness], Inline: true},
		{Name: "Affection", Value: descTiers[breed.FamilyFriendly], Inline: true},
		{Name: "Pet Friendly", Value: descTiers[breed.OtherPetsFriendly], Inline: true},
		{Name: "Children Friendly", Value: descTiers[breed.ChildrenFriendly], Inline: true},
		{Name: "Intelligence", Value: descTiers[breed.Intelligence], Inline: true},
		{Name: "General Health", Value: descTiers[breed.GeneralHealth], Inline: true},
		{Name: "Length", Value: get(breed.Length), Inline: true},
	}

	embed := &discordgo.MessageEmbed{
		Title: get(breed.Name), Description: "Good Beans", Color: helper.RangeIn(1, 16777215),
		Image:  &discordgo.MessageEmbedImage{URL: breed.ImageLink},
		Fields: fields,
	}

	return embed, nil
}

func callKatzAPI(cfg *config.Configs) ([]katz, error) {
	// Choose a random letter to search with
	randomChar := string("abcdefghijklmnopqrstuvwxyz"[rand.Intn(26)])

	url := cfg.Configs.ApiURLs.NinjaKatzAPI + randomChar

	req, err := createNinjaAPIrequest(cfg, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request katz API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	var katzObj []katz
	if err = json.NewDecoder(resp.Body).Decode(&katzObj); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
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
