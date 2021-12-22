package webScrape

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestScrapeHoroscope(t *testing.T) {
	found := false
	signNum := ""
	horoscope := ""
	sign := "Gemini"

	switch strings.ToLower(sign) {
	case "aries":
		signNum = "1"
	case "taurus":
		signNum = "2"
	case "gemini":
		signNum = "3"
	case "cancer":
		signNum = "4"
	case "leo":
		signNum = "5"
	case "virgo":
		signNum = "6"
	case "libra":
		signNum = "7"
	case "scorpio":
		signNum = "8"
	case "sagittarius":
		signNum = "9"
	case "capricorn":
		signNum = "10"
	case "aquarius":
		signNum = "11"
	case "pisces":
		signNum = "12"
	}

	c := colly.NewCollector()

	// On every p element which has style attribute call callback
	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			// link := e.Attr("font-size:16px;")

			if !found {
				if e.Text != "" {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	// Before making a request print "Visiting ..."
	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		},
	)

	// Start scraping on https://www.horoscope.com
	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(horoscope)
	fmt.Println(found)
}

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

func TestGroupChat(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

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

	/*channel, err := session.UserChannelCreate("289217573004902400")
	if err != nil {
		t.Fatal(err)
	}*/

	chn, err := session.GuildChannelCreate(cfg.ExternalServicesConfig.GuildID, "Test", 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = session.ChannelMessageSend(chn.ID, "test test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMembers(t *testing.T) {
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

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://discord.com/api/guilds/%s/members", cfg.ExternalServicesConfig.GuildID), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", "Bot "+cfg.ExternalServicesConfig.Token)
	req.Header.Add("User-Agent", "DiscordBot")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	/*res, err := http.Get("https://discord.com/api/guilds/293416960237240320/members")
	if err != nil {
		t.Fatal(err)
	}*/

	var member []*discordgo.Member

	err = json.NewDecoder(res.Body).Decode(&member)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(member)
	for _, mem := range member {
		fmt.Println(mem.User.Username)
	}
}
