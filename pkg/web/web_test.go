package web

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	gomail "gopkg.in/mail.v2"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSendEmail(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	m := gomail.NewMessage()

	// receivers
	m.SetHeader("To", cfg.Configs.Settings.Email)

	// sender
	m.SetHeader("From", cfg.Configs.Settings.Email)

	// subject
	m.SetHeader("Subject", "Test Subject")

	// E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "Test Email sent.")

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, cfg.Configs.Settings.Email, cfg.Configs.Settings.EmailPassword)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err = d.DialAndSend(m); err != nil {
		t.Fatal(err)
	}
}

func TestAudioLink(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	/*url := "https://www.youtube.com/watch?v=VXtVrNdD3YA"

	newURL := strings.Replace(url, "youtube", "youtubex2", 1)

	link, fileName, err := GetYtAudioLink("https://www.youtube.com/watch?v=VXtVrNdD3YA")
	if err != nil {
		t.Fatal(err)
	}*/
}

func TestFormatAudioFileName(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

	fileName := "293416960237240320/Audio/welcome_to_the_internet_bo_burnham_frominside-7031555991360336165.mp3"

	//split at "/"
	splitName := strings.SplitAfterN(fileName, "/", 3)
	fileName = splitName[2]

	//replace characters
	replacer := strings.NewReplacer("/", "", "_", " ", "-", "", ".mp3", "")
	fileName = replacer.Replace(fileName)

	//remove numbers
	numRegex, err := regexp.Compile("[0-9]")
	fileName = numRegex.ReplaceAllString(fileName, "")
	if err != nil {
		t.Fatal(err)
	}

	//capitalize first letters
	caser := cases.Title(language.AmericanEnglish)
	fileName = caser.String(fileName)

	fmt.Println(fileName)
}

func TestScrapeHoroscope(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

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

	// StartServer scraping on https://www.horoscope.com
	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(horoscope)
	fmt.Println(found)
}

func TestGroupChat(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	var err error
	var session *discordgo.Session

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	session, err = discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
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

	/*chn, err := session.GuildChannelCreate(cfg.Configuration.GuildID, "Test", 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = session.ChannelMessageSend(chn.ID, "test test")
	if err != nil {
		t.Fatal(err)
	}*/
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

	session, err = discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
	if err != nil {
		t.Fatal(err)
	}

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://discord.com/api/guilds/%s/members", ""), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", "Bot "+cfg.Configs.Keys.TestBotToken)
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
