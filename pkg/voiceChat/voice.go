package voiceChat

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/oleiade/lane"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

var (
	VoiceInstances = map[string]*VoiceInstance{}
)

const (
	channels  int = 2
	frameRate int = 48000
	frameSize int = 960
)

// VoiceInstance is created for each connected server
type VoiceInstance struct {
	discord      *discordgo.Session
	discordVoice *discordgo.VoiceConnection
	queue        *lane.Queue
	pcmChannel   chan []int16
	serverID     string
	skip         bool
	stop         bool
	trackPlaying bool

	cfg *config.Config
}

// QueueVideo places a Youtube link in a queue
func (vi *VoiceInstance) QueueVideo(youtubeLink string) {
	fmt.Println("Queuing video")
	vi.queue.Enqueue(youtubeLink)
}

// CreateVoiceInstance accepts both a youtube query and a server id, boots up the voice connection, and plays the track.
func CreateVoiceInstance(youtubeLink, serverID string, guild *discordgo.Guild, channelID string, cfg *config.Config) {
	vi := new(VoiceInstance)
	VoiceInstances[serverID] = vi

	vi.cfg = cfg

	fmt.Println("Connecting Voice...")
	vi.serverID = serverID
	vi.queue = lane.NewQueue()

	err := vi.connectVoice(guild, channelID)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	vi.pcmChannel = make(chan []int16, 2)
	go SendPCM(vi.discordVoice, vi.pcmChannel)

	vi.QueueVideo(youtubeLink)
	vi.processQueue()
}

func (vi *VoiceInstance) connectVoice(guild *discordgo.Guild, channelID string) error {
	var err error
	vi.discord, err = discordgo.New("Bot " + vi.cfg.ExternalServicesConfig.Token)
	// vi.discord, err = discordgo.New(vi.cfg.ExternalServicesConfig.DiscordEmail, vi.cfg.ExternalServicesConfig.DiscordPassword)
	if err != nil {
		return err
	}

	// Open the websocket and begin listening.
	err = vi.discord.Open()
	if err != nil {
		return err
	}

	channels, err := vi.discord.GuildChannels(vi.serverID)
	if err != nil {
		return err
	}

	var voiceChannel string
	var voiceChannels []string
	for _, channel := range channels {
		channelType := fmt.Sprintf("%b", channel.Type)
		if channelType == "10" { // 10 = voice channels
			if channel.ID != guild.AfkChannelID {
				if channel.ID == channelID {
					voiceChannel = channel.ID
				}
				voiceChannels = append(voiceChannels, channel.ID)
				/*if strings.Contains(strings.ToLower(channel.Name), "wtf we doing? room") && voiceChannel == "" {
					voiceChannel = channel.ID
				}*/
			}
		}
	}

	if voiceChannel == "" {
		fmt.Println("Selecting first channel")
		voiceChannel = voiceChannels[0]
	}

	_, err = vi.discord.ChannelVoiceJoin(vi.serverID, voiceChannel, false, true)
	if err != nil {
		return err
	}

	// Hacky loop to prevent returning when voice isn't ready
	// TODO: Find a better way.
	/*for vi.discord.VoiceReady == false {
		runtime.Gosched()
	}*/

	return nil
}

func (vi *VoiceInstance) processQueue() {
	if vi.trackPlaying == false {
		for {
			vi.skip = false
			link := vi.queue.Dequeue()
			if link == nil || vi.stop == true {
				break
			}
			vi.playVideo(link.(string))
		}

		// No more tracks in queue? Cleanup.
		fmt.Println("Closing connections")
		close(vi.pcmChannel)
		vi.discordVoice.Close()

		err := vi.discord.Close()
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}

		delete(VoiceInstances, vi.serverID)
		fmt.Println("Done")
	}
}

func (vi *VoiceInstance) playVideo(url string) {
	vi.trackPlaying = true

	/*url = strings.Replace(url, "\"", "", -1)
	url = "https://" + url
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Http.Get\nerror: %s\ntarget: %s\n", err, url)
		return
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("reading answer: non 200 status code received: '%s'", err)
	}

	run := exec.Command("ffmpeg.exe", "-i", "-", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	run.Stdin = resp.Body
	stdout, err := run.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe Error:", err)
		return
	}

	err = run.Start()
	if err != nil {
		fmt.Println("RunStart Error:", err)
		return
	}*/

	f, err := os.Open("../../res/Lame_Drivers_-_01_-_Frozen_Egg.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer func(streamer beep.StreamSeekCloser) {
		err := streamer.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(streamer)

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	// buffer used during loop below
	//audiobuf := make([]int16, frameSize*channels)

	err = vi.discordVoice.Speaking(true)
	if err != nil {
		fmt.Printf("%+v", errors.WithStack(err))
		return
	}

	defer func(discordVoice *discordgo.VoiceConnection, b bool) {
		err = discordVoice.Speaking(b)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			return
		}
	}(vi.discordVoice, false)

	/*for {
		// read data from ffmpeg stdout
		err = binary.Read(stdout, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			fmt.Println("error reading from ffmpeg stdout :", err)
			break
		}
		if vi.stop == true || vi.skip == true {
			err = run.Process.Kill()
			if err != nil {
				return
			}
			break
		}
		vi.pcmChannel <- audiobuf
	}*/

	vi.trackPlaying = false

	speaker.Play(streamer)
	select {}
}
