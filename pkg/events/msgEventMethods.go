package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/games"
	"github.com/beamer64/discordBot/pkg/gcp"
	"github.com/beamer64/discordBot/pkg/ssh"
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"net/url"
	"strings"
	"time"
)

// functions here should mostly be used for the prefix commands ($)

func (d *MessageCreateHandler) testMethod(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	if IsLaunchedByDebugger() {

	}
	return nil
}

func (r *ReactionHandler) sendLmgtfy(s *discordgo.Session, m *discordgo.Message) error {
	strEnc := url.QueryEscape(m.Content)
	lmgtfyURL := fmt.Sprintf("http://lmgtfy.com/?q=%s", strEnc)

	lmgtfyShortURL, err := shorturl.Shorten(lmgtfyURL, "tinyurl")
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\"%s\"\n%s", m.Content, string(lmgtfyShortURL)))
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) playNIM(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	if strings.HasPrefix(param, "<@") {
		err := games.StartNim(s, m, param, true)
		if err != nil {
			return err
		}

	} else {
		if param == "" {
			err := games.StartNim(s, m, param, false)
			if err != nil {
				return err
			}

		} else {
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *MessageCreateHandler) sendStartUpMessages(s *discordgo.Session, m *discordgo.MessageCreate) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	sm := 0
	for i := 1; i < 5; i++ {
		loadingMessage := getRandomLoadingMessage(d.cfg.LoadingMessages)
		time.Sleep(3 * time.Second)

		_, err := s.ChannelMessageSend(m.ChannelID, loadingMessage)
		if err != nil {
			return err
		}

		sm += i
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (d *MessageCreateHandler) startServer(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if d.cfg.Configs.Server.MachineIP != "" { // check if Minecraft server is set up
		client, err := gcp.NewGCPClient("config/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
		if err != nil {
			return err
		}

		err = client.StartMachine("instance-2-minecraft")
		if err != nil {
			return err
		}

		sshClient, err := ssh.NewSSHClient(d.cfg.Configs.Server.SSHKeyBody, d.cfg.Configs.Server.MachineIP)
		if err != nil {
			return err
		}

		status, serverUp := sshClient.CheckServerStatus(sshClient)
		if serverUp {
			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.ServerUP+status)
			if err != nil {
				return err
			}

		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.WindUp)
			if err != nil {
				return err
			}

			_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
			if err != nil {
				return err
			}

			err = d.sendStartUpMessages(s, m)
			if err != nil {
				return err
			}

			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.FinishOpperation)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.MCServerError)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *MessageCreateHandler) stopServer(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if d.cfg.Configs.Server.MachineIP != "" { // check if Minecraft server is set up
		sshClient, err := ssh.NewSSHClient(d.cfg.Configs.Server.SSHKeyBody, d.cfg.Configs.Server.MachineIP)
		if err != nil {
			return err
		}

		status, serverUp := sshClient.CheckServerStatus(sshClient)
		if serverUp {
			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.WindDown)

			_, err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
			if err != nil {
				return err
			}

			client, errr := gcp.NewGCPClient("config/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
			if errr != nil {
				return err
			}

			err = client.StopMachine("instance-2-minecraft")
			if err != nil {
				return err
			}

			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.FinishOpperation)
			if err != nil {
				return err
			}

		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.ServerDOWN+status)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.MCServerError)
		if err != nil {
			return err
		}
	}
	return nil
}

// d.sendServerStatusAsMessage Sends the current server status as a message in discord
func (d *MessageCreateHandler) sendServerStatusAsMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	client, err := gcp.NewGCPClient("config/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
	if err != nil {
		return err
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		return err
	}

	sshClient, err := ssh.NewSSHClient(d.cfg.Configs.Server.SSHKeyBody, d.cfg.Configs.Server.MachineIP)
	if err != nil {
		return err
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.CheckStatusUp+status)
		if err != nil {
			return err
		}

	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.CheckStatusDown+status)
		if err != nil {
			return err
		}
	}
	return nil
}
