package api

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestGetGifURL(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/
	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	type gif struct {
		Results []struct {
			Media []struct {
				Gif struct {
					Size    int    `json:"size"`
					Preview string `json:"preview"`
					Dims    []int  `json:"dims"`
					URL     string `json:"url"`
				} `json:"gif"`
			} `json:"media"`
		} `json:"results"`
	}

	URL := fmt.Sprintf("https://g.tenor.com/v1/search?q=cat&key=%s&limit=1", cfg.Configs.Keys.TenorAPIkey)

	res, err := http.Get(URL)
	if err != nil {
		t.Fatal(err)
	}

	var gifObj gif

	err = json.NewDecoder(res.Body).Decode(&gifObj)
	if err != nil {
		t.Fatal(err)
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(res.Body)

	gifURL := gifObj.Results[0].Media[0].Gif.URL

	fmt.Println(gifURL)
}

func TestPostInsult(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	var err error
	var session *discordgo.Session

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	session, err = discordgo.New("Bot " + cfg.Configs.Keys.BotToken)
	if err != nil {
		t.Fatal(err)
	}

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		t.Fatal(err)
	}

	insult, err := GetInsult(cfg.Configs.Keys.InsultAPI)
	if err != nil {
		t.Fatal(err)
	}

	memberName := "me"
	if !strings.HasPrefix(memberName, "<@") {
		channel, err := session.UserChannelCreate("289217573004902400")
		if err != nil {
			t.Fatal(err)
		}

		_, err = session.ChannelMessageSend(channel.ID, "You need to '@Mention' the user for insults. eg: @UserName")
		if err != nil {
			t.Fatal(err)
		}

	} else {
		if strings.ToLower(memberName) == "me" || strings.ToLower(memberName) == "@me" {
			fmt.Println(memberName)

			fmt.Println(insult)
		}
	}
}

func TestGetInsult(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	insultURL := cfg.Configs.Keys.InsultAPI
	res, err := http.Get(insultURL)
	if err != nil {
		t.Fatal(err)
	}

	var insultObj insult

	err = json.NewDecoder(res.Body).Decode(&insultObj)
	if err != nil {
		t.Fatal(err)
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	fmt.Println(insultObj.Insult)
}
