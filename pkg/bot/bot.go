package bot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/beamer64/discordBot/pkg/gcp"
	"github.com/beamer64/discordBot/pkg/voiceChat"

	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/ssh"
	"github.com/beamer64/discordBot/pkg/webScrape"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type DiscordBot struct {
	cfg   *config.Config
	botID string
}

func NewDiscordBot(cfg *config.Config) *DiscordBot {
	return &DiscordBot{
		cfg: cfg,
	}
}

func (d *DiscordBot) Start() error {
	goBot, err := discordgo.New("Bot " + d.cfg.ExternalServicesConfig.Token)
	if err != nil {
		return err
	}

	user, err := goBot.User("@me")
	if err != nil {
		return err
	}
	d.botID = user.ID

	goBot.AddHandler(d.messageHandler)
	err = goBot.Open()
	if err != nil {
		return err
	}

	fmt.Println("DiscordBot is running!")
	return nil
}

func (d *DiscordBot) messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, d.cfg.ExternalServicesConfig.BotPrefix) {
		if message.Author.ID == d.botID {
			return
		}

		// deletes Justins Messages
		if message.Author.ID == "282722418093719556" {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				return
			}
			return
		}

		messageSlices := strings.Split(message.Content, " ")
		command := messageSlices[0]

		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {
		// Sends command list
		case "$tuuck":
			err := d.sendMessage(session, message, d.cfg.CommandDescriptions.Tuuck+"\n"+d.cfg.CommandDescriptions.McStatus+"\n"+d.cfg.CommandDescriptions.Start+
				"\n"+d.cfg.CommandDescriptions.Stop+"\n"+d.cfg.CommandDescriptions.Horoscope+"\n"+d.cfg.CommandDescriptions.Gif+"\n"+d.cfg.CommandDescriptions.Version+
				"\n"+d.cfg.CommandDescriptions.CoinFlip)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		case "$version":
			err := d.sendMessage(session, message, "We'we wunnying vewsion `"+d.cfg.Version+"` wight nyow")
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}

		// Starts the Minecraft Server
		case "$start":
			err := d.startServer(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Stops the Minecraft Server
		case "$stop":
			err := d.stopServer(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Stops the Minecraft Server
		case "$mcstatus":
			err := d.sendServerStatusAsMessage(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// TODO make this work
		// Plays youtube link in voice chat
		case "$play":
			err := d.playYoutubeLink(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_ = d.sendMessage(session, message, "No vidya dood.")
			}
			return

		// Sends daily horoscope
		case "$horoscope":
			err := d.displayHoroscope(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends gif response
		case "$gif":
			err := d.sendGif(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		case "$coinflip":
			err := d.coinFlip(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends the "Invalid" command Message
		default:
			err := d.sendMessage(session, message, d.cfg.CommandMessages.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}

func (d *DiscordBot) sendMessage(session *discordgo.Session, message *discordgo.MessageCreate, outMessage string) error {
	_, err := session.ChannelMessageSend(message.ChannelID, outMessage)
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
