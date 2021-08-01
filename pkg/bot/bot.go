package bot

import (
	"fmt"
	"log"
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

		// Sends Daily Horoscope
		if strings.Contains(strings.ToLower(message.Content), "$horoscope/") {
			signSlices := strings.SplitAfter(message.Content, "/")
			sign := signSlices[1]
			horoscope := webScrape.ScrapeSign(sign)
			d.sendMessage(session, message, horoscope)
			return

		} else
		// Sends first searched gif
		if strings.Contains(strings.ToLower(message.Content), "$gif/") {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				return
			}

			searchSlices := strings.SplitAfter(message.Content, "/")
			searchStr := searchSlices[1]
			gifURL := webScrape.RequestGif(searchStr, d.cfg.ExternalServicesConfig.TenorAPIkey)
			d.sendMessage(session, message, gifURL)
			return

		}

		switch strings.ToLower(method) {
		// Sends command list
		case "$tuuck":
			d.sendMessage(session, message, d.cfg.Command.Tuuck+"\n"+d.cfg.Command.McStatus+"\n"+d.cfg.Command.Start+
				"\n"+d.cfg.Command.Stop+"\n"+d.cfg.Command.Horoscope+"\n"+d.cfg.Command.Gif+"\n"+d.cfg.Command.Version)
			return

		case "$version":
			d.sendMessage(session, message, "We're running version "+d.cfg.Version+" right now")

		// Starts the Minecraft Server
		case "$start":
			d.startServer(session, message)
			return

		// Stops the Minecraft Server
		case "$stop":
			d.stopServer(session, message)
			return

		// Stops the Minecraft Server
		case "$mcstatus":
			d.sendServerStatusAsMessage(session, message)
			return

		case "$play":
			d.playYoutubeLink(session, message)

		// Sends the "Invalid" command Message
		default:
			d.sendMessage(session, message, d.cfg.Command.Invalid)
			return
		}
	}
}

func (d *DiscordBot) sendMessage(session *discordgo.Session, message *discordgo.MessageCreate, outMessage string) {
	_, err := session.ChannelMessageSend(message.ChannelID, outMessage)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}
}

func (d *DiscordBot) playYoutubeLink(session *discordgo.Session, message *discordgo.MessageCreate) {
	channel, _ := session.Channel(message.ChannelID)
	serverID := channel.GuildID

	youtubeLink, youtubeTitle, err := webScrape.GetYoutubeURL(strings.Split(message.Content, " ")[1], d.cfg.ExternalServicesConfig.YoutubeAPIKey)
	if err != nil {
		fmt.Println(err)
		d.sendMessage(session, message, "No vidya dood.")
		return
	}

	if voiceChat.VoiceInstances[serverID] != nil {
		voiceChat.VoiceInstances[serverID].QueueVideo(youtubeLink)
		d.sendMessage(session, message, fmt.Sprintf("Queued: %s", youtubeTitle))
	} else {
		d.sendMessage(session, message, fmt.Sprintf("Playing: %s", youtubeTitle))
		go voiceChat.CreateVoiceInstance(youtubeLink, serverID, d.cfg)
	}
}

func (d *DiscordBot) sendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 5; i++ {
		loadingMessage := getRandomLoadingMessage(d.cfg.LoadingMessages)
		time.Sleep(3 * time.Second)
		d.sendMessage(session, message, loadingMessage)
		m += i
	}
	time.Sleep(3 * time.Second)
}

func (d *DiscordBot) startServer(session *discordgo.Session, message *discordgo.MessageCreate) {

	client, err := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		d.sendMessage(session, message, d.cfg.Command.ServerUP+status)

	} else {
		d.sendMessage(session, message, d.cfg.Command.WindUp)

		_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}
		d.sendStartUpMessages(session, message)
		d.sendMessage(session, message, d.cfg.Command.FinishOpperation)
	}
}

func (d *DiscordBot) stopServer(session *discordgo.Session, message *discordgo.MessageCreate) {

	sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		d.sendMessage(session, message, d.cfg.Command.WindDown)

		_, err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		client, err := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		err = client.StopMachine("instance-2-minecraft")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		d.sendMessage(session, message, d.cfg.Command.FinishOpperation)

	} else {
		d.sendMessage(session, message, d.cfg.Command.ServerDOWN+status)
	}
}

// d.sendServerStatusAsMessage Sends the current server status as a message in discord
func (d *DiscordBot) sendServerStatusAsMessage(session *discordgo.Session, message *discordgo.MessageCreate) {

	client, err := gcp.NewGCPClient("config/auth.json", d.cfg.GCPAuth.Project_ID, d.cfg.GCPAuth.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	sshClient, err := ssh.NewSSHClient(d.cfg.ExternalServicesConfig.SSHKeyBody, d.cfg.ExternalServicesConfig.MachineIP)
	if err != nil {
		log.Fatal(err)
	}

	status, serverUp := sshClient.CheckServerStatus(sshClient)
	if serverUp {
		d.sendMessage(session, message, d.cfg.Command.CheckStatusUp+status)
	} else {
		d.sendMessage(session, message, d.cfg.Command.CheckStatusDown+status)
	}
}

func getRandomLoadingMessage(possibleMessages []string) string {
	rand.Seed(time.Now().Unix())
	return possibleMessages[rand.Intn(len(possibleMessages))]
}
