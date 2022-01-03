package webScrape

import (
	"context"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetYtAudioLink(link string) (mpFileLink string, fileNAme string, err error) {
	url := "https://x2convert.com/en159/download-youtube-to-mp3-music"
	searchElem := "/html/body/div/div[1]/div[2]/div[2]/div[1]/input"

	// create context
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var res string
	var ok *bool

	// create submit task
	// submit the youtube link
	subTasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(searchElem),
		chromedp.SendKeys(searchElem, link),
	}

	// run submit task list
	err = chromedp.Run(ctx, subTasks)
	if err != nil {
		return "", "", err
	}

	// create click tasks to click convert button
	searchElem = "/html/body/div/div[1]/div[2]/div[4]/button"
	clickTasks := chromedp.Tasks{
		chromedp.WaitVisible(searchElem),
		chromedp.Click(searchElem),
	}

	// run clickTask list
	err = chromedp.Run(ctx, clickTasks)
	if err != nil {
		return "", "", err
	}

	// create waitTasks list to get redirect URL
	searchElem = "/html/body/div/div[1]/div[2]/div[6]/div/div[1]/div"
	waitTasks := chromedp.Tasks{
		chromedp.WaitNotPresent(searchElem),
		chromedp.Location(&res),
	}

	// run waitTasks list
	err = chromedp.Run(ctx, waitTasks)
	if err != nil {
		return "", "", err
	}

	// create navTasks to get download link
	searchElem = "/html/body/div[1]/div[1]/div/div[2]/div[3]/a[1]"
	resURL := res
	navTasks := chromedp.Tasks{
		chromedp.Navigate(resURL),
		chromedp.AttributeValue(searchElem, "href", &res, ok),
	}

	// run navTasks list
	err = chromedp.Run(ctx, navTasks)
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func PlayAudioFile(dgv *discordgo.VoiceConnection, fileName string) error {
	// Start loop and attempt to play all files in the given dir
	files, err := ioutil.ReadDir(".")

	var mpFileNames []string
	for _, f := range files {
		if contains(mpFileNames, f.Name()) {
			err = os.Remove(f.Name())
			if err != nil {
				return err
			}
		} else {
			mpFileNames = append(mpFileNames, f.Name())
			fmt.Println("PlayAudioFile:", f.Name())

			dgvoice.PlayAudioFile(dgv, f.Name(), make(chan bool))
		}
	}
	if err != nil {
		return err
	}

	// Close connections
	dgv.Close()

	return nil
}
