package webScrape

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
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

	guild, err := session.State.Guild("293416960237240320")
	if err != nil {
		t.Fatal(err)
	}

	members := guild.Members

	atMember := ""
	for _, memb := range members {
		if memb.User.Username == "Beamer64" {
			atMember = "<@" + memb.User.ID + ">"
		}
	}

	if atMember != "" {
		fmt.Println("Member: ", atMember)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Insult: ", insult)
		if err != nil {
			t.Fatal(err)
		}

	} else {
		fmt.Println("Couldn't find User.")
		if err != nil {
			t.Fatal(err)
		}
	}
}
