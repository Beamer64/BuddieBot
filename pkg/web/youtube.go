package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// "github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/chromedp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var StopPlaying chan bool
var IsPlaying bool
var MpFileQueue []string

func GetYtAudioLink(s *discordgo.Session, m *discordgo.Message, link string) (mpFileLink string, fileName string, err error) {
	// replacer := strings.NewReplacer("m.", "", "https", "http", "youtube", "youtubex2")
	replacer := strings.NewReplacer("m.", "", "youtube.com", "notube.net")
	url := replacer.Replace(link)

	ctx, cancel := chromedp.NewContext(context.Background()) // options: chromedp.WithDebugf(log.Printf)
	ctx, cancel = context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	var res string
	var ok *bool

	msg, err := s.ChannelMessageEdit(m.ChannelID, m.ID, "Prepping vidya...20% [##        ]")
	if err != nil {
		return "", "", err
	}

	// navigate to url and get redirect url
	err = chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		chromedp.Location(&res),
	)
	if err != nil {
		return "", "", err
	}

	// navigate to redirect and click button
	// Grey 'Download file MP3' button
	// button := "/html/body/div[1]/main/section[2]/div[2]/div/div[2]/div/div[2]/div/a"
	button := "/html/body/div[2]/main/section/div/div/a[1]"

	err = chromedp.Run(
		ctx,
		chromedp.Navigate(res),
		// chromedp.Click(button),
		chromedp.AttributeValue(button, "href", &res, ok),
	)
	if err != nil {
		return "", "", err
	}

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...40% [####      ]")
	if err != nil {
		return "", "", err
	}

	/*// wait for page to load and get button redirect url
	searchElem := "/html/body/div/main/section[1]/div/div/div[5]/div/div[1]/div"
	err = chromedp.Run(
		ctx,
		chromedp.WaitNotPresent(searchElem),
		chromedp.Location(&res),
	)
	if err != nil {
		return "", "", err
	}*/

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...50% [#####     ]")
	if err != nil {
		return "", "", err
	}

	/*// navigate to button redirect and get download link
	button = "/html/body/div[1]/main/section/div/div[2]/div/div[2]/div[1]/div[3]/a[1]"
	resURL := res
	err = chromedp.Run(
		ctx,
		chromedp.Navigate(resURL),
		chromedp.AttributeValue(button, "href", &res, ok),
	)
	if err != nil {
		return "", "", err
	}*/

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...70% [#######   ]")
	if err != nil {
		return "", "", err
	}

	// listen for response containing mp3 link
	mpLink := res
	/*chromedp.ListenTarget(
		ctx, func(ev interface{}) {
			if ev, ok := ev.(*network.EventResponseReceived); ok {
				if strings.Contains(ev.Response.URL, ".mp3") {
					mpLink = ev.Response.URL
				}
			}
		},
	)

	// navigate to download link to parse network response
	err = chromedp.Run(ctx, chromedp.Navigate(res))
	if err != nil {
		if !strings.Contains(err.Error(), "net::ERR_ABORTED") {
			return "", "", err
		}
	}*/

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...90% [######### ]")
	if err != nil {
		return "", "", err
	}

	time.AfterFunc(
		2*time.Second, func() {
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		},
	)

	// fileName = strings.SplitAfterN(mpLink, "/", 12)[11]
	fileName = strings.SplitAfterN(mpLink, "=", 3)[2]
	fileName = strings.ReplaceAll(fileName, "=", "")

	return mpLink, fileName, nil
}

func DownloadMpFile(m *discordgo.MessageCreate, link string, fileName string) error {
	// Get the data
	resp, err := http.Get(link)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Create the dir
	dir := fmt.Sprintf("%s/Audio", m.GuildID)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		// does not exist
		err = os.MkdirAll(dir, 0777)
		log.Println(fmt.Sprintf("Dir created: %s", dir))
	}
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(filepath.Join(dir, filepath.Base(fileName)))
	if err != nil {
		return err
	}

	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Println("Created File")

	return nil
}

func PlayAudioFile(vc *discordgo.VoiceConnection, fileName string, m *discordgo.MessageCreate, s *discordgo.Session) error {
	dir := fmt.Sprintf("%s/Audio", m.GuildID)

	var err error

	cleanFileName, err := formatAudioFileName(fileName)
	if err != nil {
		return err
	}

	if !IsPlaying {
		if fileName != "" {
			MpFileQueue = append(MpFileQueue, filepath.Join(dir, filepath.Base(fileName)))
		}

		defer vc.Speaking(false)

		IsPlaying = true
		for i, v := range MpFileQueue {
			log.Println("PlayAudioFile: ", v)

			_, err = s.ChannelMessageSend(m.ChannelID, "Now playing: "+cleanFileName)
			if err != nil {
				return err
			}

			ffmpeg := exec.Command(
				"ffmpeg",
				"-i", v, // input file
				"-f", "s16le", // raw PCM output
				"-ar", "48000", // 48kHz sample rate
				"-ac", "2", // 2 channels (stereo)
				"pipe:1", // output to stdout
			)

			stdout, err := ffmpeg.StdoutPipe()
			if err != nil {
				return err
			}

			if err = ffmpeg.Start(); err != nil {
				return err
			}

			// Mark speaking
			err = vc.Speaking(true)
			if err != nil {
				return err
			}

			// Send PCM data to Discord
			sendPCM(vc, stdout)

			// Wait for ffmpeg to finish
			if err = ffmpeg.Wait(); err != nil {
				log.Println("ffmpeg exited with error:", err)
			}

			// dgvoice.PlayAudioFile(dgv, v, StopPlaying)

			MpFileQueue = append(MpFileQueue[:i], MpFileQueue[i+1:]...)
		}
		// remove file from queue
		// MpFileQueue = nil

		if vc != nil {
			IsPlaying = false
			err = vc.Disconnect()
			if err != nil {
				return err
			}
		}

		err = MpFileCleanUp(dir)
		if err != nil {
			return err
		}

	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added to queue: %s", cleanFileName))
		if err != nil {
			return err
		}

		MpFileQueue = append(MpFileQueue, filepath.Join(dir, filepath.Base(fileName)))
	}

	return nil
}

func sendPCM(vc *discordgo.VoiceConnection, pcm io.Reader) {
	const frameSize = 960 * 2 * 2 // 960 samples * 2 bytes * 2 channels = 3840 bytes per 20ms frame
	buffer := make([]byte, frameSize)

	for {
		n, err := io.ReadFull(pcm, buffer)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			log.Println("error reading PCM:", err)
			break
		}

		// Send audio frame to Discord
		vc.OpusSend <- buffer[:n]
	}

	// Cleanly end transmission
	close(vc.OpusSend)
}

// formatAudioFileName formats audio file name to look better
func formatAudioFileName(fileName string) (string, error) {
	// replace characters
	replacer := strings.NewReplacer("/", "", "_", " ", "-", "", ".mp3", "")
	fileName = replacer.Replace(fileName)

	// remove numbers
	numRegex := regexp.MustCompile("[0-9]")
	fileName = numRegex.ReplaceAllString(fileName, "")

	// capitalize first letters
	c := cases.Title(language.AmericanEnglish)
	fileName = c.String(fileName)

	return fileName, nil
}

// MpFileCleanUp clear out Audio directory
func MpFileCleanUp(dir string) error {
	MpFileQueue = nil

	log.Println("Running Cleanup")
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".mp3") {
			err = os.Remove(filepath.Join(dir, f.Name()))
			if err != nil {
				return err
			}
		}
	}

	log.Println("Cleanup Finished")
	return nil
}
