package slash

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
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
	dog, err := retryDoggoFetch(cfg)
	if err != nil {
		return nil, fmt.Errorf("error retrieving doggo data after max # of attempts: %w", err)
	}

	get := helper.CheckIfStructValueISEmpty

	embed := &discordgo.MessageEmbed{
		Title:       dog.Name,
		Color:       helper.RandomDiscordColor(),
		Description: dog.Temperament,
		Image: &discordgo.MessageEmbedImage{
			URL: dog.Image.URL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Weight", Value: fmt.Sprintf("%slbs / %skg", get(dog.Weight.Imperial), get(dog.Weight.Metric)), Inline: true},
			{Name: "Breed Group", Value: get(dog.BreedGroup), Inline: true},
			{Name: "Origin", Value: get(dog.Origin), Inline: true},
			{Name: "Height", Value: fmt.Sprintf("%sin / %scm", get(dog.Height.Imperial), get(dog.Height.Metric)), Inline: true},
			{Name: "Life Span", Value: get(dog.LifeSpan), Inline: true},
			{Name: "Good Pup", Value: "10/10", Inline: true},
			{Name: "Bred For", Value: get(dog.BredFor), Inline: false},
		},
	}

	return embed, nil
}

// retryDoggoFetch picks random breed IDs and calls the per-breed endpoint
// until it gets a valid hit or runs out of attempts. About 90% of IDs in the
// range are valid (the rest 404), so 5 attempts is effectively guaranteed to
// succeed under normal conditions.
func retryDoggoFetch(cfg *config.Configs) (doggo, error) {
	const maxAttempts = 5
	const maxDoggoBreedID = 697

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		id := rand.Intn(maxDoggoBreedID) + 1
		dog, err := callDoggoAPI(cfg, id)
		if err == nil {
			return dog, nil
		}
		lastErr = err
	}
	return doggo{}, fmt.Errorf("no valid dog data after %d attempts: %w", maxAttempts, lastErr)
}

func callDoggoAPI(cfg *config.Configs, id int) (doggo, error) {
	url := cfg.Configs.ApiURLs.DoggoAPI + strconv.Itoa(id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return doggo{}, fmt.Errorf("creating request to Doggo API: %w", err)
	}

	req.Header.Set("x-api-key", cfg.Configs.Keys.DoggoAPIkey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return doggo{}, fmt.Errorf("making request to Doggo API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return doggo{}, fmt.Errorf("breed %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		return doggo{}, fmt.Errorf("doggo API returned status code %d", resp.StatusCode)
	}

	var dog doggo
	if err := json.NewDecoder(resp.Body).Decode(&dog); err != nil {
		return doggo{}, fmt.Errorf("decoding Doggo API response: %w", err)
	}

	return dog, nil
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
		Title: get(breed.Name), Description: "Good Beans", Color: helper.RandomDiscordColor(),
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request katz API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	var katzObj []katz
	if err = json.NewDecoder(resp.Body).Decode(&katzObj); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return katzObj, nil
}

func createNinjaAPIrequest(cfg *config.Configs, url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("x-api-key", cfg.Configs.Keys.NinjaAPIKey)
	return req, nil
}

func animalsSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "animals",
		Description: "So CUTE",
		Options: []*discordgo.ApplicationCommandOption{
			/*{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "doggo",
				Description: "🐕",
				Required:    false,
			},*/
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "katz",
				Description: "😻",
				Required:    false,
			},
		},
	}
}
