package webScrape

import (
	"context"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetYtAudioLink(s *discordgo.Session, m *discordgo.Message, link string) (mpFileLink string, fileNAme string, err error) {
	url := strings.Replace(link, "youtube", "youtubex2", 1)

	// create context
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var res string
	var ok *bool

	msg, err := s.ChannelMessageEdit(m.ChannelID, m.ID, "Prepping vidya...20% [##        ]")
	if err != nil {
		return "", "", err
	}

	// navigate to url and get redirect url
	NavTasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Location(&res),
	}
	// run navigate task list
	err = chromedp.Run(ctx, NavTasks)
	if err != nil {
		return "", "", err
	}

	// navigate to redirect and click button
	button := "/html/body/div[1]/main/section[2]/div[2]/div/div[2]/div/div[2]/div/a"
	clickTasks := chromedp.Tasks{
		chromedp.Navigate(res),
		chromedp.Click(button),
	}
	// run clickTask list
	err = chromedp.Run(ctx, clickTasks)
	if err != nil {
		return "", "", err
	}

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...40% [####      ]")
	if err != nil {
		return "", "", err
	}

	// wait for page to load and get button redirect url
	searchElem := "/html/body/div/main/section[1]/div/div/div[5]/div/div[1]/div"
	waitTasks := chromedp.Tasks{
		chromedp.WaitNotPresent(searchElem),
		chromedp.Location(&res),
	}

	// run waitTasks list
	err = chromedp.Run(ctx, waitTasks)
	if err != nil {
		return "", "", err
	}

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...50% [#####     ]")
	if err != nil {
		return "", "", err
	}

	// navigate to button redirect and get download link
	button = "/html/body/div[1]/main/section/div/div[2]/div/div[2]/div[1]/div[3]/a[1]"
	resURL := res
	navTasks := chromedp.Tasks{
		chromedp.Navigate(resURL),
		chromedp.AttributeValue(button, "href", &res, ok),
	}

	// run navTasks list
	err = chromedp.Run(ctx, navTasks)
	if err != nil {
		return "", "", err
	}

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...70% [#######   ]")
	if err != nil {
		return "", "", err
	}

	// navigate to download link to parse network response
	getLinkTasks := chromedp.Tasks{
		chromedp.Navigate(res),
	}

	// listen for response containing mp3 link
	mpLink := ""
	chromedp.ListenTarget(
		ctx, func(ev interface{}) {
			if ev, ok := ev.(*network.EventResponseReceived); ok {
				if strings.Contains(ev.Response.URL, ".mp3") {
					mpLink = ev.Response.URL
					//fmt.Println("closing alert:", ev.Response)
				}
			}
		},
	)

	// run getLinkTasks list
	err = chromedp.Run(ctx, getLinkTasks)
	if !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		return "", "", err
	}

	msg, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Prepping vidya...90% [######### ]")
	if err != nil {
		return "", "", err
	}

	time.AfterFunc(
		2*time.Second, func() {
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		},
	)

	fileName := strings.SplitAfterN(mpLink, "/", 12)[11]

	return mpLink, fileName, nil
}

func DownloadMpFile(link string, fileName string) error {
	// Get the data
	resp, err := http.Get(link)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	fmt.Println("Created File")

	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func PlayAudioFile(dgv *discordgo.VoiceConnection, fileQueue []string) error {
	// Start loop and attempt to play all files in the given dir

	for _, v := range fileQueue {
		fmt.Println("PlayAudioFile:", v)

		dgvoice.PlayAudioFile(dgv, v, make(chan bool))
		err := os.Remove(v)
		fileQueue = fileQueue[1:]
		if err != nil {
			return err
		}
	}

	// Close connections
	dgv.Close()

	return nil
}
