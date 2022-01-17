package ssh

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/gcp"
)

func TestRunCommand(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	sshClient, err := NewSSHClient(cfg.Configs.Server.SSHKeyBody, cfg.Configs.Server.MachineIP)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sshClient.RunCommand("sudo echo hello")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunCommandStartContainer(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	sshClient, err := NewSSHClient(cfg.Configs.Server.SSHKeyBody, cfg.Configs.Server.MachineIP)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckServerStatus(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	// need to be careful importing from other packages in tests, could cause issues in future
	client, err := gcp.NewGCPClient("../config/auth.json", cfg.Configs.Server.Project_ID, cfg.Configs.Server.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	sshClient, err := NewSSHClient(cfg.Configs.Server.SSHKeyBody, cfg.Configs.Server.MachineIP)
	if err != nil {
		t.Fatal(err)
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		fmt.Println("Server is up. " + status)
		return

	} else {
		fmt.Println("Server is down. " + status)
	}
}
