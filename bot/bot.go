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

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate, member *discordgo.Member) {
	if strings.HasPrefix(message.Content, cfg.BotPrefix) {
		if message.Author.ID == DiscordBotID {
			return
		}

		//deletes Justins Messages
		if message.Author.ID == "282722418093719556" {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		/*if memberHasRole(member, "test") {

		}*/

		//Sends command list
		if message.Content == "$tuuck" {
			SendMessage(session, message, comm.Tuuck+"\n"+comm.Start+"\n"+comm.Stop+"\n"+comm.Horoscope)
			return

		} else
		//Sends Daily Horoscope
		if strings.Contains(message.Content, "$horoscope-") {
			signSlices := strings.SplitAfter(message.Content, "-")
			sign := signSlices[1]
			horoscope := webScrape.ScrapeSign(sign)
			SendMessage(session, message, horoscope)
			return

		} else
		//Starts the Minecraft Server
		if message.Content == "$start" {
			StartServer(session, message)
			return

		} else
		//Stops the Minecraft Server
		if message.Content == "$stop" {
			StopServer(session, message)
			return

		} else {
			//Sends the "Invalid" command Message
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

func SendStartUpMessages(session *discordgo.Session, message *discordgo.MessageCreate) {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	m := 0
	for i := 1; i < 6; i++ {
		loadingMessage := config.GrabLoadingMessage("bot/loadingMessages.txt")
		time.Sleep(10 * time.Second)
		SendMessage(session, message, loadingMessage)
		m += i
	}
	time.Sleep(10 * time.Second)
}

func StartServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	loadingMessage := config.GrabLoadingMessage("bot/loadingMessages.txt")
	SendMessage(session, message, loadingMessage+comm.WindUp)

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
