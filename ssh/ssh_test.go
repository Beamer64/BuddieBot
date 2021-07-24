package ssh

import (
	"testing"

	"github.com/beamer64/discordBot/config"
)

func TestRunCommand(t *testing.T) {
	config, err := config.ReadConfig("../config/config.json")
	if err != nil {
		t.Fatal(err)
	}

	sshClient, err := NewSSHClient(config.SSHKeyBody, "34.68.22.97:22")
	if err != nil {
		t.Fatal(err)
	}

	err = sshClient.RunCommand("sudo echo hello")
	if err != nil {
		t.Fatal(err)
	}
}
