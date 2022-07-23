package events

import (
	"fmt"
	"github.com/beamer64/buddieBot/pkg/gcp"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/beamer64/buddieBot/pkg/ssh"
	"github.com/beamer64/buddieBot/pkg/voice_chat"
	"github.com/beamer64/buddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"net/url"
	"strings"
	"time"
)

// functions here should mostly be used for the prefix commands ($)

//region dev commands
func (d *MessageCreateHandler) testMethod(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	if helper.IsLaunchedByDebugger() {
		//err := d.playAudioLink(s, m, param)
		err := d.sendStartUpMessages(s, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *MessageCreateHandler) sendReleaseNotes(s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := &discordgo.MessageEmbed{
		Title: "Release Notes!",
		URL:   "https://github.com/Beamer64/BuddieBot/blob/master/res/release.md",
		Description: "SUM MOAR BIG BOI CHANGES\n\nDetailed list can be found in the Title link above." +
			"\nCheck it out\n-----------------------------------------------------------------------------\n\n- Command changes:",
		Color: 11091696,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    m.Author.Username,
			IconURL: m.Author.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "New Command Group: /ratethis",
				Value:  "Give/Get some new ratings",
				Inline: false,
			},
			{
				Name:   "New Commands: /pick album",
				Value:  "<@!282722418093719556>'s Album recommender api. Recommends a music album based on liked tags.",
				Inline: false,
			},
			{
				Name:   "New Commands: /pick poll",
				Value:  "Poll command...for polling things..",
				Inline: false,
			},
			{
				Name:   "New Commands: ${COMMAND} SpongeBob easter egg",
				Value:  "It's my bot, I can do what I want.",
				Inline: false,
			},
			{
				Name:   "Bug Fix: Youtube mobile links",
				Value:  "(When working..) Audio will play with the mobile link 'm.youtube.com...'",
				Inline: false,
			},
			{
				Name:   "Enhancement: Audio Queue",
				Value:  "The Audio Queue will show a cleaned title from the old 'Name-Title-Sum_Numbers.mp3'.",
				Inline: false,
			},
		},
	}

	msg := &discordgo.MessageSend{
		Content: "@everyone",
		Embed:   embed,
	}

	if helper.IsLaunchedByDebugger() {
		_, err := s.ChannelMessageSendComplex(m.ChannelID, msg)
		if err != nil {
			return err
		}
	} else {
		for _, guild := range s.State.Guilds {
			for _, channel := range guild.Channels {
				if channel.Type == discordgo.ChannelTypeGuildText {
					_, err := s.ChannelMessageSendComplex(channel.ID, msg)
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}
	return nil
}

//endregion

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

func (d *MessageCreateHandler) sendStartUpMessages(s *discordgo.Session, m *discordgo.MessageCreate) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	sm := 0
	for i := 1; i < 5; i++ {
		loadingMessage := helper.GetRandomStringFromSet(d.cfg.LoadingMessages)
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
		client, err := gcp.NewGCPClient("config_files/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
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

			client, errr := gcp.NewGCPClient("config_files/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
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
	client, err := gcp.NewGCPClient("config_files/auth.json", d.cfg.Configs.Server.Project_ID, d.cfg.Configs.Server.Zone)
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

//region audio commands
func (d *MessageCreateHandler) playAudioLink(s *discordgo.Session, m *discordgo.MessageCreate, link string) error {
	msg, err := s.ChannelMessageSend(m.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	link, fileName, err := web.GetYtAudioLink(s, msg, link)
	if err != nil {
		return err
	}

	err = web.DownloadMpFile(m, link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, m.Author.ID, m.GuildID)
	if err != nil {
		return err
	}

	err = web.PlayAudioFile(dgv, fileName, m, s)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) stopAudioPlayback() error {
	//vc := voice_chat.VoiceConnection{}

	if web.StopPlaying != nil {
		close(web.StopPlaying)
		web.IsPlaying = false

		/*if vc.Dgv != nil {
			vc.Dgv.Close()

		}*/
	}

	return nil
}

func (d *MessageCreateHandler) sendQueue(s *discordgo.Session, m *discordgo.MessageCreate) error {
	queue := ""
	if len(web.MpFileQueue) > 0 {
		queue = strings.Join(web.MpFileQueue, "\n")
	} else {
		queue = "Uh owh, song queue is wempty (>.<)"
	}

	_, err := s.ChannelMessageSend(m.ChannelID, queue)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) sendSkipMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	audio := ""
	if len(web.MpFileQueue) > 0 {
		audio = fmt.Sprintf("Skipping %s", web.MpFileQueue[0])
	} else {
		audio = "Queue is empty, my guy"
	}

	_, err := s.ChannelMessageSend(m.ChannelID, audio)
	if err != nil {
		return err
	}

	err = d.skipPlayback(s, m)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) skipPlayback(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(web.MpFileQueue) > 0 {
		err := d.stopAudioPlayback()
		if err != nil {
			return err
		}

		dgv, err := voice_chat.ConnectVoiceChannel(s, m.Author.ID, m.GuildID)
		if err != nil {
			return err
		}

		err = web.PlayAudioFile(dgv, "", m, s)
		if err != nil {
			return err
		}
	}

	return nil
}

//endregion

func (d *MessageCreateHandler) sendWeasterEgg(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(
		m.ChannelID,
		"Is mayonnaise an instrument?\n───────────────▄████████▄────────\n──────────────██▒▒▒▒▒▒▒▒██───────\n─────────────██▒▒▒▒▒▒▒▒▒██───────\n────────────██▒▒▒▒▒▒▒▒▒▒██───────\n"+
			"───────────██▒▒▒▒▒▒▒▒▒██▀────────\n"+
			"──────────██▒▒▒▒▒▒▒▒▒▒██─────────\n─────────██▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒████▒████▒▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒─██─▒▒─██─▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n───────██▒▒█▀▀▀▀▀▀▀█▒▒▒▒██───────\n─────██▒▒▒▒▒█▄▄▄▄▄█▒▒▒▒▒▒▒██─────\n───██▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒██───\n─██▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒██─\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒████▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒████▒▒█\n▀████▒▒▒▒▒▒▒▒▒▓▓▓▓▒▒▒▒▒▒▒▒▒▒████▀\n──█▌▌▌▌▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▌▌▌███──\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n────▀█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌██▀─────\n─────█▌▌▌▌▌▌████████▌▌▌▌▌██──────\n──────██▒▒██────────██▒▒██───────\n──────▀████▀────────▀████▀───────",
	)
	if err != nil {
		return err
	}

	return nil
}
