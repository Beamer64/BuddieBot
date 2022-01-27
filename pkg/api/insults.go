package api

import (
	"encoding/json"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"io"
	"net/http"
)

type insult struct {
	Insult string `json:"insult"`
}

func GetInsult(insultURL string) (string, error) {
	res, err := http.Get(insultURL)
	if err != nil {
		return "", err
	}

	var insultObj insult

	err = json.NewDecoder(res.Body).Decode(&insultObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return insultObj.Insult, nil
}

func GetInsultEmbed(randColorInt int, cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
	retVal := ""
	if cfg.Configs.Keys.InsultAPI != "" { // check if insult API is set up
		insult, err := GetInsult(cfg.Configs.Keys.InsultAPI)
		if err != nil {
			return nil, err
		}

		retVal = insult
	} else {
		retVal = cfg.Cmd.Msg.InsultAPIError
	}

	embed := &discordgo.MessageEmbed{
		Title:       "(ง ͠° ͟ل͜ ͡°)ง",
		Color:       randColorInt,
		Description: retVal,
	}

	return embed, nil
}
