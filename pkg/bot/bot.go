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

		method := strings.Split(message.Content, " ")[0][0:]

		switch strings.ToLower(method) {
		// Sends command list
		case "$tuuck":
			err := d.sendMessage(session, message, d.cfg.Command.Tuuck+"\n"+d.cfg.Command.McStatus+"\n"+d.cfg.Command.Start+
				"\n"+d.cfg.Command.Stop+"\n"+d.cfg.Command.Horoscope+"\n"+d.cfg.Command.Gif+"\n"+d.cfg.Command.Version)
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
			err := d.playYoutubeLink(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_ = d.sendMessage(session, message, "No vidya dood.")
			}
			return

		// Sends daily horoscope
		case "$horoscope":
			err := d.displayHoroscope(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends gif response
		case "$gif":
			err := d.sendGif(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends the "Invalid" command Message
		default:
			err := d.sendMessage(session, message, d.cfg.Command.Invalid)
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

func (d *DiscordBot) sendGif(session *discordgo.Session, message *discordgo.MessageCreate) error {
	err := session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		return err
	}

	searchStr := strings.Split(message.Content, " ")[1]
	gifURL, err := webScrape.RequestGif(searchStr, d.cfg.ExternalServicesConfig.TenorAPIkey)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, gifURL)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) displayHoroscope(session *discordgo.Session, message *discordgo.MessageCreate) error {
	sign := strings.Split(message.Content, " ")[1]
	horoscope, err := webScrape.ScrapeSign(sign)
	if err != nil {
		return err
	}

	err = d.sendMessage(session, message, horoscope)
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscordBot) playYoutubeLink(session *discordgo.Session, message *discordgo.MessageCreate) error {
	channel, _ := session.Channel(message.ChannelID)
	serverID := channel.GuildID

	youtubeLink, youtubeTitle, err := webScrape.GetYoutubeURL(strings.Split(message.Content, " ")[1], d.cfg.ExternalServicesConfig.YoutubeAPIKey)
	if err != nil {
		return err
	}

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
		err = d.sendMessage(session, message, d.cfg.Command.ServerUP+status)
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, d.cfg.Command.WindUp)
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

		err = d.sendMessage(session, message, d.cfg.Command.FinishOpperation)
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
		err = d.sendMessage(session, message, d.cfg.Command.WindDown)

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

		err = d.sendMessage(session, message, d.cfg.Command.FinishOpperation)
		if err != nil {
			return err
		}

	} else {
		err = d.sendMessage(session, message, d.cfg.Command.ServerDOWN+status)
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
		err = d.sendMessage(session, message, d.cfg.Command.CheckStatusUp+status)
		if err != nil {
			return err
		}
	} else {
		err = d.sendMessage(session, message, d.cfg.Command.CheckStatusDown+status)
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
