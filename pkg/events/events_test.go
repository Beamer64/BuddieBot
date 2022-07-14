package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
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

	//AuditLogActionMemberKick = 20
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
