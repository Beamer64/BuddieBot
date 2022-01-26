package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/voice_chat"
	"github.com/beamer64/discordBot/pkg/web_scrape"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"time"
)

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func coinFlip(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
	gifURL, err := api.RequestGifURL("Coin Flip", cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Flipping...",
		Color: 16761856,
		Author: &discordgo.MessageEmbedAuthor{
			Name: "BuddieBot",
			IconURL: "https://camo.githubusercontent.com/97c16e17070b00f5c5db3447703233bf007dd60706c46db66aa5042a417277a7" +
				"/68747470733a2f2f696d6167652e666c617469636f6e2e636f6d2f69636f6e732f706e672f3531322f343639382f343639383738372e706e67",
		},
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

func playYoutubeLink(s *discordgo.Session, i *discordgo.InteractionCreate, param string) error {
	msg, err := s.ChannelMessageSend(i.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	//yas
	if i.Member.User.ID == "932843527870742538" {
		param = "https://www.youtube.com/watch?v=kJQP7kiw5Fk"
	}

	link, fileName, err := web_scrape.GetYtAudioLink(s, msg, param)
	if err != nil {
		return err
	}

	err = web_scrape.DownloadMpFile(link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
	if err != nil {
		return err
	}

	err = web_scrape.PlayAudioFile(dgv, fileName, i.ChannelID, s)
	if err != nil {
		return err
	}

	return nil
}

func stopAudioPlayback() error {
	//vc := voice_chat.VoiceConnection{}

	if web_scrape.StopPlaying != nil {
		close(web_scrape.StopPlaying)
		web_scrape.IsPlaying = false

		/*if vc.Dgv != nil {
			vc.Dgv.Close()

		}*/
	}

	return nil
}

func skipPlayback(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(web_scrape.MpFileQueue) > 0 {
		err := stopAudioPlayback()
		if err != nil {
			return err
		}

		dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
		if err != nil {
			return err
		}

		err = web_scrape.PlayAudioFile(dgv, "", i.ChannelID, s)
		if err != nil {
			return err
		}
	}

	return nil
}
