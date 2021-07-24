package ssh

import (
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	config *ssh.ClientConfig
	ip     string
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

func (client *SSHClient) RunCommand(commandText string) error {
	conn, err := ssh.Dial("tcp", client.ip, client.config)
	if err != nil {
		return err
	}

	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, sessStdOut)
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stderr, sessStderr)
	err = sess.Run(commandText) // eg., /usr/bin/whoami
	if err != nil {
		return err
	}

	return nil
}
