package gcp

import (
	"github.com/beamer64/buddieBot/pkg/config"
	"os"
	"testing"
)

func TestStopMachine(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewGCPClient("../config_files/auth.json", cfg.Configs.Server.Project_ID, cfg.Configs.Server.Zone)
	if err != nil {
		t.Fatal(err)
	}
	err = client.StopMachine("instance-2-minecraft")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartMachine(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewGCPClient("../config_files/auth.json", cfg.Configs.Server.Project_ID, cfg.Configs.Server.Zone)
	if err != nil {
		t.Fatal(err)
	}
	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		t.Fatal(err)
	}
}
