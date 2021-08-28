package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/gcp"
	"github.com/beamer64/discordBot/pkg/ssh"
	"github.com/beamer64/discordBot/pkg/voiceChat"
	"github.com/beamer64/discordBot/pkg/webScrape"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"math/rand"
	"strings"
	"time"
)

func (d *DiscordBot) testMethod(session *discordgo.Session, message *discordgo.MessageCreate) {
	err := d.playYoutubeLink(session, message, "https://www.youtube.com/watch?v=7tC6DUaPqfM")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
	}
}

func (d *DiscordBot) sendHelpMessage(session *discordgo.Session, message *discordgo.MessageCreate) error {
	var cmds []string
	if d.memberHasRole(session, message, d.cfg.ExternalServicesConfig.BotAdminRole) { // bot mod
		cmds = append(cmds, d.cfg.CommandDescriptions.Tuuck, d.cfg.CommandDescriptions.Horoscope, d.cfg.CommandDescriptions.Version,
			d.cfg.CommandDescriptions.CoinFlip, d.cfg.CommandDescriptions.LMGTFY)

		if d.cfg.ExternalServicesConfig.TenorAPIkey != "" {
			cmds = append(cmds, d.cfg.CommandDescriptions.Gif)
		}
		if d.cfg.ExternalServicesConfig.MachineIP != "" {
			cmds = append(cmds, d.cfg.CommandDescriptions.McStatus, d.cfg.CommandDescriptions.Start, d.cfg.CommandDescriptions.Stop)
		}
		if d.cfg.ExternalServicesConfig.InsultAPI != "" {
			cmds = append(cmds, d.cfg.CommandDescriptions.Insult)
		}

	} else {
		cmds = append(cmds, d.cfg.CommandDescriptions.Tuuck, d.cfg.CommandDescriptions.Horoscope,
			d.cfg.CommandDescriptions.CoinFlip, d.cfg.CommandDescriptions.LMGTFY)

		if d.cfg.ExternalServicesConfig.TenorAPIkey != "" {
			cmds = append(cmds, d.cfg.CommandDescriptions.Gif)
		}
		if d.cfg.ExternalServicesConfig.InsultAPI != "" {
			cmds = append(cmds, d.cfg.CommandDescriptions.Insult)
		}
	}

	cmdDesc := ""
	for _, command := range cmds {
		cmdDesc = cmdDesc + "\n" + command
	}

	_, err := session.ChannelMessageSend(message.ChannelID, cmdDesc)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) sendLmgtfy(session *discordgo.Session, message *discordgo.MessageCreate) error {
	err := session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		return err
	}

	provider := "tinyurl"
	lmgtfyMsg, err := webScrape.FindLMGTFY(session, message, d.botID)
	if err != nil {
		return err
	}

	lmgtfyURL := webScrape.LmgtfyURL(lmgtfyMsg.Content)

	lmgtfyShortURL, err := ShortenURL(lmgtfyURL, provider)
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSend(message.ChannelID, "\""+lmgtfyMsg.Content+"\""+"\n"+lmgtfyShortURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) coinFlip(session *discordgo.Session, message *discordgo.MessageCreate) error {
	_, err := session.ChannelMessageSend(message.ChannelID, "Flipping...")
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)
	_, err = session.ChannelMessageSend(message.ChannelID, "...")
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)
	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	if randNum%2 == 0 {
		err = d.sendGif(session, message, "heads")
		if err != nil {
			return err
		}

	} else {
		err = d.sendGif(session, message, "tails")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DiscordBot) sendGif(session *discordgo.Session, message *discordgo.MessageCreate, param string) error {
	if d.cfg.ExternalServicesConfig.TenorAPIkey != "" { // check if Tenor API set up
		err := session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			return err
		}

		gifURL, err := webScrape.RequestGif(param, d.cfg.ExternalServicesConfig.TenorAPIkey)
		if err != nil {
			return err
		}

		_, err = session.ChannelMessageSend(message.ChannelID, param)
		if err != nil {
			return err
		}

		_, err = session.ChannelMessageSend(message.ChannelID, gifURL)
		if err != nil {
			return err
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.TenorAPIError)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DiscordBot) displayHoroscope(session *discordgo.Session, message *discordgo.MessageCreate, param string) error {
	horoscope, err := webScrape.ScrapeSign(param)
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSend(message.ChannelID, horoscope)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) playYoutubeLink(session *discordgo.Session, message *discordgo.MessageCreate, param string) error {
	if d.cfg.ExternalServicesConfig.TenorAPIkey != "" { // check if YouTube API set up
		guild, err := session.State.Guild(message.GuildID)
		if err != nil {
			return err
		}

		channel, _ := session.Channel(message.ChannelID)
		serverID := channel.GuildID

		youtubeLink, youtubeTitle, err := webScrape.GetYoutubeURL(param, d.cfg.ExternalServicesConfig.YoutubeAPIKey)
		if err != nil {
			return err
		}

		/*youtubeLink, youtubeTitle, err := webScrape.GetYoutubeURL("https://www.youtube.com/watch?v=72hjeHtSEfg&pp=sAQA", d.cfg.ExternalServicesConfig.YoutubeAPIKey)
		if err != nil {
			return err
		}*/

		if voiceChat.VoiceInstances[serverID] != nil {
			voiceChat.VoiceInstances[serverID].QueueVideo(youtubeLink)
			_, err = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Queued: %s", youtubeTitle))
			if err != nil {
				return err
			}

		} else {
			_, err = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Playing: %s", youtubeTitle))
			if err != nil {
				return err
			}

			go voiceChat.CreateVoiceInstance(youtubeLink, serverID, guild, channel.ID, d.cfg)
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.YoutubeAPIError)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DiscordBot) sendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 5; i++ {
		loadingMessage := getRandomLoadingMessage(d.cfg.LoadingMessages)
		time.Sleep(3 * time.Second)

		_, err := session.ChannelMessageSend(message.ChannelID, loadingMessage)
		if err != nil {
			return err
		}

		m += i
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (d *DiscordBot) startServer(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if d.cfg.ExternalServicesConfig.MachineIP != "" { // check if Minecraft server is set up
		client, err := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
		if err != nil {
			return err
		}

		err = client.StartMachine("instance-2-minecraft")
		if err != nil {
			return err
		}

		sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
		if err != nil {
			return err
		}

		status, serverUp := sshClient.CheckServerStatus(sshClient)
		if serverUp {
			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.ServerUP+status)
			if err != nil {
				return err
			}

		} else {
			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.WindUp)
			if err != nil {
				return err
			}

			_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
			if err != nil {
				return err
			}

			err = d.sendStartUpMessages(session, message)
			if err != nil {
				return err
			}

			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.FinishOpperation)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.MCServerError)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DiscordBot) stopServer(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if d.cfg.ExternalServicesConfig.MachineIP != "" { // check if Minecraft server is set up
		sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
		if err != nil {
			return err
		}

		status, serverUp := sshClient.CheckServerStatus(sshClient)
		if serverUp {
			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.WindDown)

			_, err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
			if err != nil {
				return err
			}

			client, errr := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
			if errr != nil {
				return err
			}

			err = client.StopMachine("instance-2-minecraft")
			if err != nil {
				return err
			}

			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.FinishOpperation)
			if err != nil {
				return err
			}

		} else {
			_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.ServerDOWN+status)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.MCServerError)
		if err != nil {
			return err
		}
	}
	return nil
}

// d.sendServerStatusAsMessage Sends the current server status as a message in discord
func (d *DiscordBot) sendServerStatusAsMessage(session *discordgo.Session, message *discordgo.MessageCreate) error {
	client, err := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
	if err != nil {
		return err
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		return err
	}

	sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
	if err != nil {
		return err
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.CheckStatusUp+status)
		if err != nil {
			return err
		}

	} else {
		_, err = session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.CheckStatusDown+status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DiscordBot) postInsult(session *discordgo.Session, message *discordgo.MessageCreate, memberName string) error {
	if d.cfg.ExternalServicesConfig.MachineIP != "" { // check if insult API is set up
		insult, err := webScrape.GetInsult(d.cfg.ExternalServicesConfig.InsultAPI)
		if err != nil {
			return err
		}

		// get some info before deleting msg
		msgChannelID := message.ChannelID
		msgAuthorID := message.Author.ID

		err = session.ChannelMessageDelete(msgChannelID, message.ID)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(memberName, "<@") {
			if strings.ToLower(memberName) == "me" || strings.ToLower(memberName) == "@me" {
				_, err = session.ChannelMessageSend(msgChannelID, "<@!"+msgAuthorID+">"+"\n"+insult)
				if err != nil {
					return err
				}

			} else {
				channel, err := session.UserChannelCreate(msgAuthorID)
				if err != nil {
					return err
				}

				_, err = session.ChannelMessageSend(channel.ID, "You need to '@Mention' the user for insults. eg: @UserName")
				if err != nil {
					return err
				}
			}

		} else {
			_, err = session.ChannelMessageSend(msgChannelID, memberName+"\n"+insult)
			if err != nil {
				return err
			}
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.InsultAPIError)
		if err != nil {
			return err
		}
	}

	return nil
}
