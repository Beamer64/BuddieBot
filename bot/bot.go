package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamer64/discordBot/config"
	"github.com/beamer64/discordBot/gcp"
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

		//deletes Justins Messages
		if message.Author.ID == "282722418093719556" {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				return
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
				fmt.Printf("%+v", errors.WithStack(err))
				return
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
			SendMessage(session, message, comm.Tuuck+"\n"+comm.McStatus+"\n"+comm.Start+
				"\n"+comm.Stop+"\n"+comm.Horoscope+"\n"+comm.Gif)
			return

		//Starts the Minecraft Server
		case "$start":
			StartServer(session, message)
			return

		//Stops the Minecraft Server
		case "$stop":
			StopServer(session, message)
			return

		//Stops the Minecraft Server
		case "$mcstatus":
			GetServerStatus(session, message)
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
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}
}

func GetServerStatus(session *discordgo.Session, message *discordgo.MessageCreate) {
	client, err := gcp.NewGCPClient("config/auth.json", ath.Project_id, ath.Zone)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	sshClient, err := ssh.NewSSHClient(cfg.SSHKeyBody, cfg.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, comm.CheckStatusUp+status)
	} else {
		SendMessage(session, message, comm.CheckStatusDown+status)
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
		time.Sleep(3 * time.Second)
		SendMessage(session, message, loadingMessage)
		m += i
	}
	time.Sleep(3 * time.Second)
}

func StartServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	client, err := gcp.NewGCPClient("config/auth.json", ath.Project_id, ath.Zone)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	sshClient, err := ssh.NewSSHClient(cfg.SSHKeyBody, cfg.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, comm.ServerUP+status)

	} else {
		SendMessage(session, message, comm.WindUp)

		_, err = sshClient.RunCommand("docker container start 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}
		SendStartUpMessages(session, message)
		SendMessage(session, message, comm.FinishOpperation)
	}
}

func StopServer(session *discordgo.Session, message *discordgo.MessageCreate) {
	sshClient, err := ssh.NewSSHClient(cfg.SSHKeyBody, cfg.MachineIP)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	status, serverUp := ssh.CheckServerStatus(sshClient)
	if serverUp {
		SendMessage(session, message, comm.WindDown)

		_, err = sshClient.RunCommand("docker container stop 06ae729f5c2b")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		client, err := gcp.NewGCPClient("config/auth.json", ath.Project_id, ath.Zone)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		err = client.StopMachine("instance-2-minecraft")
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		SendMessage(session, message, comm.FinishOpperation)

	} else {
		SendMessage(session, message, comm.ServerDOWN+status)
	}
}
