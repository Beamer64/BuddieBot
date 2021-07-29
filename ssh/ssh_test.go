package ssh

import (
	"fmt"
	"github.com/beamer64/discordBot/gcp"
	"log"
	"testing"

	"github.com/beamer64/discordBot/config"
)

func TestRunCommand(t *testing.T) {
	config, _, _, err := config.ReadConfig("../config/config.json", "../config/config.json", "../config/config.json")
	if err != nil {
		t.Fatal(err)
	}

	sshClient, err := NewSSHClient(config.SSHKeyBody, "34.68.22.97:22")
	if err != nil {
		t.Fatal(err)
	}

	_, err = sshClient.RunCommand("sudo echo hello")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckServerStatus(t *testing.T) {
	config, auth, _, err := config.ReadConfig("../config/config.json", "../config/auth.json", "../config/config.json")
	if err != nil {
		t.Fatal(err)
	}

	client, err := gcp.NewGCPClient("../config/auth.json", auth.Project_id, auth.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	sshClient, err := NewSSHClient(config.SSHKeyBody, config.MachineIP)
	if err != nil {
		t.Fatal(err)
	}

	status, serverUp := CheckServerStatus(sshClient)
	if serverUp {
		fmt.Println("Server is up. " + status)
		return

	} else {
		fmt.Println("Server is down. " + status)
	}
}
