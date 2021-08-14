package webScrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/parnurzeal/gorequest"
)

type stream map[string]string
type youtube struct {
	streamList []stream
	videoID    string
	videoInfo  *http.Response
}

// GetYoutubeURL converts a standard Youtube URL or ID to an mp4 download link,
// or searches Youtube and picks the first result.
func GetYoutubeURL(query, apiKey string) (string, string, error) {
	y := new(youtube)

	if len(query) < 4 || query[:4] != "http" {
		link, err := searchYoutube(query, apiKey)
		if err != nil {
			return "", "", err
		}
		query = link
	}

	err := y.findVideoID(query)
	if err != nil {
		return "", "", fmt.Errorf("findVideoID error=%s", err)
	}

	err = y.getVideoInfo(apiKey)
	if err != nil {
		return "", "", fmt.Errorf("getVideoInfo error=%s", err)
	}

	err = y.parseVideoInfo()
	if err != nil {
		return "", "", fmt.Errorf("parse video info failed, err=%s", err)
	}

	targetStream := y.streamList[0]
	// downloadURL := targetStream["url"] // + "&signature=" + targetStream["sig"]
	downloadURL := "https://r5---sn-8vap5-ig3e.googlevideo.com/videoplayback?expire=1628847584&ei=gOkVYZDXA9rAsALfpbnIDA&ip=77.120.162.238&id=o-AICdw3FfnCt2zRHyaKFSLTU1jYIhXwq_ETyFXzhVoTJZ&itag=22&source=youtube&requiressl=yes&mh=xB&mm=31%2C29&mn=sn-8vap5-ig3e%2Csn-8vap5-3c2k&ms=au%2Crdu&mv=m&mvi=5&pl=20&initcwndbps=1455000&vprv=1&mime=video%2Fmp4&ns=cWnZqnBlGeQufnNSCewjzB8G&cnr=14&ratebypass=yes&dur=13015.666&lmt=1628721465342371&mt=1628825562&fvip=2&fexp=24001373%2C24007246&c=WEB&txp=7316222&n=FaUCC17TfXyDde2nbLL&sparams=expire%2Cei%2Cip%2Cid%2Citag%2Csource%2Crequiressl%2Cvprv%2Cmime%2Cns%2Ccnr%2Cratebypass%2Cdur%2Clmt&sig=AOq0QJ8wRQIhALuniCICt5AV6v9ijiASrxyEpwS8NKEUIpYlyUtFSQuMAiAl0EgfwpCP0rDtRuTPtPDMHOddaGIdW13qUBm25R0W_g%3D%3D&lsparams=mh%2Cmm%2Cmn%2Cms%2Cmv%2Cmvi%2Cpl%2Cinitcwndbps&lsig=AG3C_xAwRQIgPFGMDz-u-Gvj8i-iH-NsqUBRFI2xSVM76ar4xNqGaIICIQCzacneKFNaj0hD6HhMZ71NNjWFcwmvvKNV88KTQwobqQ%3D%3D&title=Rimworld%20IS%20A%20PERFECTLY%20BALANCED%20GAME%20WITH%20NO%20EXPLOITS%20-%20Organ%20Harvesting%20Live%20(BIRTHDAY%20STREAM)"
	return downloadURL, targetStream["title"], nil
}

func searchYoutube(text, apiKey string) (string, error) {

	formattedURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&key=%s", url.QueryEscape(text), apiKey)

	_, body, err := gorequest.New().Get(formattedURL).EndBytes()
	if len(err) > 0 { // ??? never seen an []err before
		return "", err[0]
	}

	jsonParsed, _ := gabs.ParseJSON(body)
	children, _ := jsonParsed.S("items").Children()
	if len(children) == 0 {
		return "", fmt.Errorf("No video found")
	}

	videoID, _ := children[0].Path("id.videoId").Data().(string)
	return videoID, nil
}

func (y *youtube) findVideoID(videoID string) error {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		reList := []*regexp.Regexp{
			regexp.MustCompile(`(?:v|embed|watch\?v)(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`([^"&?/=%]{11})`),
		}
		for _, re := range reList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
				break
			}
		}
	}

	y.videoID = videoID
	if strings.ContainsAny(videoID, "?&/<%=") {
		return errors.New("invalid characters in video id")
	}
	if len(videoID) < 10 {
		return errors.New("the video id must be at least 10 characters long")
	}
	return nil
}

func (y *youtube) getVideoInfo(apiKey string) error {
	URL := "https://www.googleapis.com/youtube/v3/videos?id=" + y.videoID + "&key=" + apiKey + "&part=snippet,status,player"

	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	y.videoInfo = resp
	return nil
}

func (y *youtube) parseVideoInfo() error {
	var responseResults map[string]interface{}
	videoTitle := ""
	videoURL := ""

	// answer, err := url.ParseQuery(y.videoInfo)
	err := json.NewDecoder(y.videoInfo.Body).Decode(&responseResults)
	if err != nil {
		return err
	}

	for key, responseValue := range responseResults {
		if key == "items" {
			for key, itemsValue := range responseValue.([]interface{})[0].(map[string]interface{}) {
				if key == "status" {
					for key, statusValue := range itemsValue.(map[string]interface{}) {
						if key == "uploadStatus" {
							if statusValue != "processed" {
								err = fmt.Errorf("no response status found in the server's answer")
								return err
							}
							break
						}
					}
				}
			}
		}
	}
	for key, responseValue := range responseResults {
		if key == "items" {
			for key, itemsValue := range responseValue.([]interface{})[0].(map[string]interface{}) {
				if key == "player" {
					for key, playerValue := range itemsValue.(map[string]interface{}) {
						if key == "embedHtml" {
							rawURLValue := fmt.Sprintf("%v", playerValue)
							rawURLString := strings.SplitAfterN(rawURLValue, "\"", 7)[5]
							URL := strings.Replace(rawURLString, "/", "", -1)
							URL = strings.Replace(URL, "\"", "", -1)
							videoURL = URL
						}
						break
					}
					break
				}
			}
		}
	}

	stream := stream{
		"quality": "",
		"type":    "",
		"url":     videoURL,
		"sig":     "",
		"title":   videoTitle,
		"author":  "",
	}

	y.streamList = append(y.streamList, stream)
	return nil
}
