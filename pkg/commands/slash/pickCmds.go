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
	"net/url"
	"strings"
)

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
		// we need to separate by commas and spaces and add brackets because API bad
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
	// gameURL := fmt.Sprintf("steam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Appid)
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
	msgEdit.Embeds = &[]*discordgo.MessageEmbed{embed}

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
