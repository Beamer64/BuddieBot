package webScrape

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
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
