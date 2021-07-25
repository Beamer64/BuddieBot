package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/beamer64/discordBot/config"
	"github.com/beamer64/discordBot/gcp"
	"github.com/beamer64/discordBot/ssh"
	"github.com/beamer64/discordBot/webScrape"
	"github.com/bwmarrin/discordgo"
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
			log.Fatal(err)
		}

		member, _ := session.GuildMember(channel.GuildID, message.Author.ID)
		fmt.Println(message.MentionRoles)

		if memberHasRole(member, "The Big Gays") {
		}*/ //TODO mess with this more

		//deletes Justins Messages
		if message.Author.ID == "282722418093719556" {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				log.Fatal(err)
			}
			return
		} else

		//Sends Daily Horoscope
		if strings.Contains(ToLower(message.Content), "$horoscope/") {
			signSlices := strings.SplitAfter(message.Content, "/")
			sign := signSlices[1]
			horoscope := webScrape.ScrapeSign(sign)
			SendMessage(session, message, horoscope)
			return

		} else
		//Sends first searched gif
		if strings.Contains(ToLower(message.Content), "$gif/") {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				log.Fatal(err)
			}

			searchSlices := strings.SplitAfter(message.Content, "/")
			searchStr := searchSlices[1]
			gifURL := webScrape.RequestGif(searchStr, cfg)
			SendMessage(session, message, gifURL)
			return
		}

		switch ToLower(message.Content) {
		//Sends command list
		case "$tuuck":
			SendMessage(session, message, comm.Tuuck+"\n"+comm.Start+"\n"+comm.Stop+"\n"+comm.Horoscope+"\n"+comm.Gif)
			return

		//Starts the Minecraft Server
		case "$start":
			StartServer(session, message)
			return

		//Stops the Minecraft Server
		case "$stop":
			StopServer(session, message)
			return

		//Sends the "Invalid" command Message
		default:
			SendMessage(session, message, comm.Invalid)
			return
		}
	}
}

func SendMessage(session *discordgo.Session, message *discordgo.MessageCreate, outMessage string) {
	_, err := session.ChannelMessageSend(message.ChannelID, outMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func ToLower(content string) string {
	return strings.ToLower(content)
}

func SendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 5; i++ {
		loadingMessage := config.GrabLoadingMessage()
		time.Sleep(5 * time.Second)
		SendMessage(session, message, loadingMessage)
		m += i
	}
	time.Sleep(5 * time.Second)
}

func StartServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	SendMessage(session, message, comm.WindUp)

	client, err := gcp.NewGCPClient("config/auth.json", ath.Project_id, ath.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	SendStartUpMessages(session, message)

	sshClient, err := ssh.NewSSHClient(cfg.SSHKeyBody, cfg.MachineIP)
	if err != nil {
		log.Fatal(err)
	}

	err = sshClient.RunCommand("docker container start 06ae729f5c2b")
	if err != nil {
		log.Fatal(err)
	}

	SendMessage(session, message, comm.FinishOpperation)
}

func StopServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	SendMessage(session, message, comm.WindDown)

	sshClient, err := ssh.NewSSHClient(cfg.SSHKeyBody, cfg.MachineIP)
	if err != nil {
		log.Fatal(err)
	}

	err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
	if err != nil {
		log.Fatal(err)
	}

	client, err := gcp.NewGCPClient("config/auth.json", ath.Project_id, ath.Zone)
	if err != nil {
		log.Fatal(err)
	}

	err = client.StopMachine("instance-2-minecraft")
	if err != nil {
		log.Fatal(err)
	}

	SendMessage(session, message, comm.FinishOpperation)
}
