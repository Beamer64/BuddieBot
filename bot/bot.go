package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/gcp"
	"github.com/beamer64/discordBot/voiceChat"
	"log"
	"strings"
	"time"

	"github.com/beamer64/discordBot/config"
	"github.com/beamer64/discordBot/ssh"
	"github.com/beamer64/discordBot/webScrape"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var DiscordBotID string
var cfg *config.Config
var ath *config.Auth
var comm *config.Command

func Start(c *config.Config, a *config.Auth, com *config.Command) error {
	cfg = c
	ath = a
	comm = com

	goBot, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return err
	}

	user, err := goBot.User("@me")
	if err != nil {
		return err
	}
	DiscordBotID = user.ID

	goBot.AddHandler(messageHandler)
	err = goBot.Open()
	if err != nil {
		return err
	}

	fmt.Println("DiscordBot is running!")
	return nil
}

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, cfg.BotPrefix) {
		if message.Author.ID == DiscordBotID {
			return
		}

		/*var author *discordgo.Member
				channel, err := session.State.Channel(message.ChannelID)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
		return
				}

				member, _ := session.GuildMember(channel.GuildID, message.Author.ID)
				fmt.Println(message.MentionRoles)

				if memberHasRole(member, "The Big Gays") {
				}*/ //TODO mess with this more

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
		if strings.Contains(ToLower(message.Content), "$horoscope/") {
			signSlices := strings.SplitAfter(message.Content, "/")
			sign := signSlices[1]
			horoscope := webScrape.ScrapeSign(sign)
			SendMessage(session, message, horoscope)
			return

		} else
		// Sends first searched gif
		if strings.Contains(ToLower(message.Content), "$gif/") {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				return
			}

			searchSlices := strings.SplitAfter(message.Content, "/")
			searchStr := searchSlices[1]
			gifURL := webScrape.RequestGif(searchStr, cfg)
			SendMessage(session, message, gifURL)
			return

		}

		switch ToLower(method) {
		// Sends command list
		case "$tuuck":
			SendMessage(session, message, comm.Tuuck+"\n"+comm.McStatus+"\n"+comm.Start+
				"\n"+comm.Stop+"\n"+comm.Horoscope+"\n"+comm.Gif)
			return

		// Starts the Minecraft Server
		case "$start":
			StartServer(session, message)
			return

		// Stops the Minecraft Server
		case "$stop":
			StopServer(session, message)
			return

		// Stops the Minecraft Server
		case "$mcstatus":
			SendServerStatusAsMessage(session, message)
			return

		case "$play":
			PlayYoutubeLink(session, message)

		// Sends the "Invalid" command Message
		default:
			SendMessage(session, message, comm.Invalid)
			return
		}
	}
}

func SendMessage(session *discordgo.Session, message *discordgo.MessageCreate, outMessage string) {
	_, err := session.ChannelMessageSend(message.ChannelID, outMessage)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}
}

func ToLower(content string) string {
	return strings.ToLower(content)
}

func PlayYoutubeLink(session *discordgo.Session, message *discordgo.MessageCreate) {
	channel, _ := session.Channel(message.ChannelID)
	serverID := channel.GuildID

	youtubeLink, youtubeTitle, err := webScrape.GetYoutubeURL(strings.Split(message.Content, " ")[1])
	if err != nil {
		fmt.Println(err)
		SendMessage(session, message, "No vidya dood.")
		return
	}

	if voiceChat.VoiceInstances[serverID] != nil {
		voiceChat.VoiceInstances[serverID].QueueVideo(youtubeLink)
		SendMessage(session, message, fmt.Sprintf("Queued: %s", youtubeTitle))
	} else {
		SendMessage(session, message, fmt.Sprintf("Playing: %s", youtubeTitle))
		go voiceChat.CreateVoiceInstance(youtubeLink, serverID)
	}
}

func SendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 5; i++ {
		loadingMessage := config.GrabLoadingMessage()
		time.Sleep(3 * time.Second)
		SendMessage(session, message, loadingMessage)
		m += i
	}
	time.Sleep(3 * time.Second)
}

func StartServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	c := ssh.NewConfigStruct()

	client, err := gcp.NewGCPClient("config/auth.json", c.Ath.Project_id, c.Ath.Zone)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	sshClient, err := ssh.NewSSHClient(c.Cfg.SSHKeyBody, c.Cfg.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, c.Comm.ServerUP+status)

	} else {
		SendMessage(session, message, c.Comm.WindUp)

		_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}
		SendStartUpMessages(session, message)
		SendMessage(session, message, c.Comm.FinishOpperation)
	}
}

func StopServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	c := ssh.NewConfigStruct()

	sshClient, err := ssh.NewSSHClient(c.Cfg.SSHKeyBody, c.Cfg.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, c.Comm.WindDown)

		_, err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		client, err := gcp.NewGCPClient("config/auth.json", c.Ath.Project_id, c.Ath.Zone)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		err = client.StopMachine("instance-2-minecraft")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		SendMessage(session, message, c.Comm.FinishOpperation)

	} else {
		SendMessage(session, message, c.Comm.ServerDOWN+status)
	}
}

// SendServerStatusAsMessage Sends the current server status as a message in discord
func SendServerStatusAsMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	c := ssh.NewConfigStruct()

	client, err := gcp.NewGCPClient("config/auth.json", c.Ath.Project_id, c.Ath.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	sshClient, err := ssh.NewSSHClient(c.Cfg.SSHKeyBody, c.Cfg.MachineIP)
	if err != nil {
		log.Fatal(err)
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, c.Comm.CheckStatusUp+status)
	} else {
		SendMessage(session, message, c.Comm.CheckStatusDown+status)
	}
}
