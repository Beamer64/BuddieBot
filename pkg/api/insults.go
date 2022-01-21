package api

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"io"
	"net/http"
	"strings"
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

func PostInsult(user string, cfg *config.ConfigStructs) (string, error) {
	retVal := ""
	if cfg.Configs.Keys.InsultAPI != "" { // check if insult API is set up
		if !strings.HasPrefix(user, "<@") {
			retVal = "You need to '@Mention' the user for insults, Dingus. eg: @UserName"

			return retVal, nil
		}

		insult, err := GetInsult(cfg.Configs.Keys.InsultAPI)
		if err != nil {
			return "", err
		}

		retVal = fmt.Sprintf("%s %s", user, insult)

	} else {
		retVal = cfg.Cmd.Msg.InsultAPIError
	}

	return retVal, nil
}
