package events

import (
	"encoding/json"
	"fmt"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"testing"
)

func TestGetAuditLog(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	session, err := discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
	if err != nil {
		t.Fatal(err)
	}

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		t.Fatal(err)
	}

	// AuditLogActionMemberKick = 20
	log, err := session.GuildAuditLog(cfg.Configs.DiscordIDs.TestGuildID, "", "", 20, 100)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(log.AuditLogEntries)
}

func TestStateMembers(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	session, err := discordgo.New("Bot " + cfg.Configs.Keys.TestBotToken)
	if err != nil {
		t.Fatal(err)
	}

	// Open the websocket and begin listening.
	err = session.Open()
	if err != nil {
		t.Fatal(err)
	}

	guild, err := session.State.Guild(cfg.Configs.DiscordIDs.TestGuildID)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(guild)
}

type NSFWResp struct {
	StatusCode string `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Meta       struct {
		Tag struct {
			Timestamp float64 `json:"timestamp"`
			Model     string  `json:"model"`
			Config    string  `json:"config"`
		} `json:"tag"`
	} `json:"meta"`
	Results []struct {
		Docid      float64 `json:"docid"`
		URL        string  `json:"url"`
		StatusCode string  `json:"status_code"`
		StatusMsg  string  `json:"status_msg"`
		LocalID    string  `json:"local_id"`
		Result     struct {
			Tag struct {
				Classes []string  `json:"classes"`
				Probs   []float64 `json:"probs"`
			} `json:"tag"`
		} `json:"result"`
		DocidStr string `json:"docid_str"`
	} `json:"results"`
}

func TestModNSFWimgs(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	// Analyze the image at https://samples.clarifai.com/nsfw.jpg
	url := fmt.Sprintf("https://api.clarifai.com/v1/tag/?model=nsfw-v1.0&url=https://upload.wikimedia.org/wikipedia/commons/3/3a/Cat03.jpg")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+cfg.Configs.Keys.ModerationPATKey)

	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	var record NSFWResp

	if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
		t.Fatal(err)
	}

	fmt.Println(record.Results[0].Result.Tag.Probs)

}
