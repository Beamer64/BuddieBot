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

func TestGetYoutubeURL(t *testing.T) {
	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	query := "https://www.youtube.com/watch?v=72hjeHtSEfg&pp=sAQA"
	apiKey := cfg.ExternalServicesConfig.YoutubeAPIKey
	y := new(youtube)

	lenQuery := len(query)
	if lenQuery < 4 || query[:4] != "http" {
		link, errr := searchYoutube(query, apiKey)
		if errr != nil {
			t.Fatal(err)
		}
		query = link
	}

	err = y.findVideoID(query)
	if err != nil {
		t.Fatal(err)
	}

	err = y.getVideoInfo(apiKey)
	if err != nil {
		t.Fatal(err)
	}

	err = y.parseVideoInfo()
	if err != nil {
		t.Fatal(err)
	}

	targetStream := y.streamList[0]
	downloadURL := targetStream["url"] + "&signature=" + targetStream["sig"]

	fmt.Println("Download URL: ", downloadURL)
	fmt.Println("Target Stream: ", targetStream["title"])
	// return downloadURL, targetStream["title"], nil
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

	session, err = discordgo.New("Bot " + cfg.ExternalServicesConfig.Token)
	if err != nil {
		t.Fatal(err)
	}

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		t.Fatal(err)
	}

	insult, err := GetInsult(cfg.ExternalServicesConfig.InsultAPI)
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
	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	insultURL := cfg.ExternalServicesConfig.InsultAPI
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
