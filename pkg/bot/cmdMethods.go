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

func (d *DiscordBot) sendHelpMessage(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if d.memberHasRole(session, message, d.cfg.ExternalServicesConfig.BotAdminRole) {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandDescriptions.Tuuck+"\n"+d.cfg.CommandDescriptions.McStatus+"\n"+d.cfg.CommandDescriptions.Start+
			"\n"+d.cfg.CommandDescriptions.Stop+"\n"+d.cfg.CommandDescriptions.Horoscope+"\n"+d.cfg.CommandDescriptions.Gif+"\n"+d.cfg.CommandDescriptions.Version+
			"\n"+d.cfg.CommandDescriptions.CoinFlip+"\n"+d.cfg.CommandDescriptions.LMGTFY+"\n"+d.cfg.CommandDescriptions.Insult)
		if err != nil {
			return err
		}

	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandDescriptions.Tuuck+"\n"+d.cfg.CommandDescriptions.Horoscope+
			"\n"+d.cfg.CommandDescriptions.Gif+"\n"+d.cfg.CommandDescriptions.CoinFlip+"\n"+d.cfg.CommandDescriptions.LMGTFY+"\n"+d.cfg.CommandDescriptions.Insult)
		if err != nil {
			return err
		}
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
	return nil
}

func (d *DiscordBot) stopServer(session *discordgo.Session, message *discordgo.MessageCreate) error {

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
	err := session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		return err
	}

	insult, err := webScrape.GetInsult(d.cfg.ExternalServicesConfig.InsultAPI)
	if err != nil {
		return err
	}

	members, err := GetGuildMembers(session, message.GuildID)
	if err != nil {
		return err
	}

	atMember := GetMentionedMemberFromList(memberName, members)

	if atMember != "" {
		_, err = session.ChannelMessageSend(message.ChannelID, atMember)
		if err != nil {
			return err
		}

		_, err = session.ChannelMessageSend(message.ChannelID, insult)
		if err != nil {
			return err
		}

	} else {
		_, err = session.ChannelMessageSend(message.ChannelID, "UwUser must be wost ~.~")
		if err != nil {
			return err
		}
	}
	return nil
}
