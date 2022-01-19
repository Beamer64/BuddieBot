package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/games"
	"github.com/beamer64/discordBot/pkg/gcp"
	"github.com/beamer64/discordBot/pkg/ssh"
	"github.com/beamer64/discordBot/pkg/voice_chat"
	"github.com/beamer64/discordBot/pkg/web_scrape"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
	"time"
)

func (d *MessageCreateHandler) testMethod(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	err := d.playYoutubeLink(s, m, param)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) sendHelpMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	var cmds []string
	if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) { // bot mod
		cmds = append(
			cmds, d.cfg.Cmd.Desc.Tuuck, d.cfg.Cmd.Desc.Horoscope, d.cfg.Cmd.Desc.Version,
			d.cfg.Cmd.Desc.CoinFlip, d.cfg.Cmd.Desc.LMGTFY, d.cfg.Cmd.Desc.Play, d.cfg.Cmd.Desc.Stop,
			d.cfg.Cmd.Desc.Queue,
		)

		if d.cfg.Configs.Server.MachineIP != "" {
			cmds = append(cmds, d.cfg.Cmd.Desc.ServerStatus, d.cfg.Cmd.Desc.StartServer, d.cfg.Cmd.Desc.StopServer)
		}
		if d.cfg.Configs.Keys.InsultAPI != "" {
			cmds = append(cmds, d.cfg.Cmd.Desc.Insult)
		}

	} else {
		cmds = append(
			cmds, d.cfg.Cmd.Desc.Tuuck, d.cfg.Cmd.Desc.Horoscope,
			d.cfg.Cmd.Desc.CoinFlip, d.cfg.Cmd.Desc.LMGTFY,
			d.cfg.Cmd.Desc.Play, d.cfg.Cmd.Desc.Stop, d.cfg.Cmd.Desc.Queue,
		)

		if d.cfg.Configs.Keys.InsultAPI != "" {
			cmds = append(cmds, d.cfg.Cmd.Desc.Insult)
		}
	}

	cmdDesc := ""
	for _, command := range cmds {
		cmdDesc = cmdDesc + "\n" + command
	}

	_, err := s.ChannelMessageSend(m.ChannelID, cmdDesc)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReactionHandler) sendLmgtfy(s *discordgo.Session, m *discordgo.Message) error {
	lmgtfyURL := CreateLmgtfyURL(m.Content)

	lmgtfyShortURL, err := ShortenURL(lmgtfyURL)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "\""+m.Content+"\""+"\n"+lmgtfyShortURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) coinFlip(s *discordgo.Session, m *discordgo.MessageCreate) error {
	gifURL, err := api.RequestGifURL("Coin Flip", d.cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Coin Flip",
		Description: "Flipping...",
		Color:       16761856,
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}
	t := &discordgo.WebhookParams{
		Username:  "BuddieBot",
		AvatarURL: "https://camo.githubusercontent.com/97c16e17070b00f5c5db3447703233bf007dd60706c46db66aa5042a417277a7/68747470733a2f2f696d6167652e666c617469636f6e2e636f6d2f69636f6e732f706e672f3531322f343639382f343639383738372e706e67",
		Embeds: []*discordgo.MessageEmbed{
			embed,
		},
	}

	_, err = s.WebhookEdit(d.cfg.Configs.DiscordIDs.WebHookID, "", "", m.ChannelID)
	if err != nil {
		return err
	}
	whMessage, err := s.WebhookExecute(d.cfg.Configs.DiscordIDs.WebHookID, d.cfg.Configs.Keys.WebHookToken, true, t)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)
	err = s.ChannelMessageDelete(whMessage.ChannelID, whMessage.ID)
	if err != nil {
		return err
	}

	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	results := ""
	if randNum%2 == 0 {
		results = "Heads"
		gifURL, err = api.RequestGifURL(results, d.cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return err
		}

	} else {
		results = "Tails"
		gifURL, err = api.RequestGifURL(results, d.cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return err
		}
	}

	embed.Description = fmt.Sprintf("It's %s!", results)
	embed.Image = &discordgo.MessageEmbedImage{
		URL: gifURL,
	}

	_, err = s.WebhookExecute(d.cfg.Configs.DiscordIDs.WebHookID, d.cfg.Configs.Keys.WebHookToken, false, t)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) displayHoroscope(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	horoscope, err := web_scrape.ScrapeSign(param)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, horoscope)
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

func (d *MessageCreateHandler) playYoutubeLink(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	msg, err := s.ChannelMessageSend(m.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	//yas
	if m.Author.ID == "932843527870742538" {
		param = "https://www.youtube.com/watch?v=kJQP7kiw5Fk"
	}

	link, fileName, err := web_scrape.GetYtAudioLink(s, msg, param)
	if err != nil {
		return err
	}

	err = web_scrape.DownloadMpFile(link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, m, m.GuildID, d.cfg.Configs.DiscordIDs.ErrorLogChannelID)
	if err != nil {
		return err
	}

	err = web_scrape.PlayAudioFile(dgv, fileName, m.ChannelID, s)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) stopAudioPlayback() error {
	vc := voice_chat.VoiceConnection{}

	if web_scrape.StopPlaying != nil {
		close(web_scrape.StopPlaying)
		web_scrape.IsPlaying = false

		if vc.Dgv != nil {
			vc.Dgv.Close()
		}
	}

	return nil
}

func (d *MessageCreateHandler) getSongQueue(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(web_scrape.MpFileQueue) > 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, strings.Join(web_scrape.MpFileQueue, "\n"))
		if err != nil {
			return err
		}

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Uh owh, song queue is wempty (>.<)")
		if err != nil {
			return err
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

func (d *MessageCreateHandler) postInsult(s *discordgo.Session, m *discordgo.MessageCreate, memberName string) error {
	if d.cfg.Configs.Keys.InsultAPI != "" { // check if insult API is set up
		insult, err := api.GetInsult(d.cfg.Configs.Keys.InsultAPI)
		if err != nil {
			return err
		}

		// get some info before deleting msg
		msgChannelID := m.ChannelID
		msgAuthorID := m.Author.ID

		err = s.ChannelMessageDelete(msgChannelID, m.ID)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(memberName, "<@") {
			if strings.ToLower(memberName) == "me" || strings.ToLower(memberName) == "@me" {
				_, err = s.ChannelMessageSend(msgChannelID, "<@!"+msgAuthorID+">"+"\n"+insult)
				if err != nil {
					return err
				}

			} else {
				channel, err := s.UserChannelCreate(msgAuthorID)
				if err != nil {
					return err
				}

				_, err = s.ChannelMessageSend(channel.ID, "You need to '@Mention' the user for insults. eg: @UserName")
				if err != nil {
					return err
				}
			}

		} else {
			_, err = s.ChannelMessageSend(msgChannelID, memberName+"\n"+insult)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.InsultAPIError)
		if err != nil {
			return err
		}
	}

	return nil
}
