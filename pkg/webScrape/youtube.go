package webScrape

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io"
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

	// create click tasks
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

	// create capTask list for download link
	searchElem = "#btnDown"
	capTasks := chromedp.Tasks{
		chromedp.Sleep(15 * time.Second),
		chromedp.Location(&res),
	}

	// run clickTask list
	err = chromedp.Run(ctx, capTasks)
	if err != nil {
		return "", "", err
	}

	searchElem = "/html/body/div[1]/div[1]/div/div[2]/div[3]/a[1]"
	resURL := res
	navTasks := chromedp.Tasks{
		chromedp.Navigate(resURL),
		chromedp.AttributeValue(searchElem, "href", &res, ok),
	}

	// run clickTask list
	err = chromedp.Run(ctx, navTasks)
	if err != nil {
		return "", "", err
	}

	getLinkTasks := chromedp.Tasks{
		chromedp.Navigate(res),
	}

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

	// run clickTask list
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

	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
