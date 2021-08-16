package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/gcp"
	"github.com/beamer64/discordBot/pkg/ssh"
	"github.com/beamer64/discordBot/pkg/voiceChat"
	"github.com/beamer64/discordBot/pkg/webScrape"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"time"
)

func (d *DiscordBot) sendMessage(session *discordgo.Session, message *discordgo.MessageCreate, outMessage string) error {
	_, err := session.ChannelMessageSend(message.ChannelID, outMessage)
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

	lmgtfyShortURL, err := webScrape.ShortenURL(lmgtfyURL, provider)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, "\""+lmgtfyMsg.Content+"\""+"\n"+lmgtfyShortURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) coinFlip(session *discordgo.Session, message *discordgo.MessageCreate) error {
	err := d.sendMessage(session, message, "Flipping...")
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)
	err = d.sendMessage(session, message, "...")
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
	err := session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		return err
	}

	gifURL, err := webScrape.RequestGif(param, d.cfg.ExternalServicesConfig.TenorAPIkey)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, param)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, gifURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) displayHoroscope(session *discordgo.Session, message *discordgo.MessageCreate, param string) error {
	horoscope, err := webScrape.ScrapeSign(param)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, horoscope)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) playYoutubeLink(session *discordgo.Session, message *discordgo.MessageCreate, param string) error {
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
		err = d.sendMessage(session, message, fmt.Sprintf("Queued: %s", youtubeTitle))
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, fmt.Sprintf("Playing: %s", youtubeTitle))
		if err != nil {
			return err
		}
		go voiceChat.CreateVoiceInstance(youtubeLink, serverID, d.cfg)
	}

	return nil
}

func (d *DiscordBot) sendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 5; i++ {
		loadingMessage := getRandomLoadingMessage(d.cfg.LoadingMessages)
		time.Sleep(3 * time.Second)

		err := d.sendMessage(session, message, loadingMessage)
		if err != nil {
			return err
		}

		m += i
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (d *DiscordBot) startServer(session *discordgo.Session, message *discordgo.MessageCreate) error {

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
		err = d.sendMessage(session, message, d.cfg.CommandMessages.ServerUP+status)
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, d.cfg.CommandMessages.WindUp)
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

		err = d.sendMessage(session, message, d.cfg.CommandMessages.FinishOpperation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DiscordBot) stopServer(session *discordgo.Session, message *discordgo.MessageCreate) error {

	sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
	if err != nil {
		return err
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		err = d.sendMessage(session, message, d.cfg.CommandMessages.WindDown)

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

		err = d.sendMessage(session, message, d.cfg.CommandMessages.FinishOpperation)
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, d.cfg.CommandMessages.ServerDOWN+status)
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
		err = d.sendMessage(session, message, d.cfg.CommandMessages.CheckStatusUp+status)
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, d.cfg.CommandMessages.CheckStatusDown+status)
		if err != nil {
			return err
		}
	}
	return nil
}

func getRandomLoadingMessage(possibleMessages []string) string {
	rand.Seed(time.Now().Unix())
	return possibleMessages[rand.Intn(len(possibleMessages))]
}
