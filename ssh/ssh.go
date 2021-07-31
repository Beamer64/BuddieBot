package ssh

import (
	"encoding/json"
	"github.com/beamer64/discordBot/config"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type SSHClient struct {
	config *ssh.ClientConfig
	ip     string
}

type ConfigStructs struct {
	Cfg  *config.Config
	Ath  *config.Auth
	Comm *config.Command
}

func NewConfigStruct() *ConfigStructs {
	cfig, auth, commd, err := config.ReadConfig("config/config.json", "config/auth.json", "config/command.json")
	if err != nil {
		log.Fatal(err)
	}

	return &ConfigStructs{
		Cfg:  cfig,
		Ath:  auth,
		Comm: commd,
	}
}

func NewSSHClient(sshKeyContents, ip string) (*SSHClient, error) {
	sshFileName := "./minecraft_rsa"

	// check if auth file exists
	if _, err := os.Stat(sshFileName); os.IsNotExist(err) {
		// if file doesn't exist, make it
		f, err := os.OpenFile(sshFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(sshKeyContents))
		if err != nil {
			return nil, err
		}

		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	pubKeyAuth, err := publicKey(sshFileName)
	if err != nil {
		return nil, err
	}

	return &SSHClient{
		config: &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				pubKeyAuth,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		ip: ip,
	}, nil
}
func publicKey(path string) (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func (client *SSHClient) RunCommand(commandText string) (string, error) {
	conn, err := ssh.Dial("tcp", client.ip, client.config)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}

	defer sess.Close()

	outPutByte, err := sess.CombinedOutput(commandText)
	if err != nil {
		return "", err
	}

	outPut := string(outPutByte)
	return outPut, nil
}

// CheckServerStatus Returns any cmd output along with whether server is running as bool
func CheckServerStatus(sshClient *SSHClient) (string, bool) {
	serverOutput, err := sshClient.RunCommand("docker container ls --format='{{json .}}'")
	if err != nil {
		log.Fatal(err)
	}

	status := ParseServerStatus(serverOutput)

	if strings.Contains(status, "Up") {
		return status, true
	}
	return status, false
}

// ParseServerStatus Parses out the server status from the cmd output
func ParseServerStatus(serverOut string) string {
	var commOut *config.CommandOut

	if serverOut != "" {
		in := []byte(serverOut)

		err := json.Unmarshal(in, &commOut)
		if err != nil {
			log.Fatal(err)
		}

		return commOut.Status
	} else {
		return serverOut
	}
}
